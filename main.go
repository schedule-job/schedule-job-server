package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	parser "github.com/Sotaneum/go-args-parser"
	"github.com/gin-gonic/gin"
	ginsession "github.com/go-session/gin-session"
	oauth "github.com/schedule-job/schedule-job-authorization/core"
	oauthGithub "github.com/schedule-job/schedule-job-authorization/github"
	db "github.com/schedule-job/schedule-job-database/core"
	"github.com/schedule-job/schedule-job-database/pg"
	schedule_errors "github.com/schedule-job/schedule-job-errors"
	gateway "github.com/schedule-job/schedule-job-gateway"
)

type Options struct {
	Port           string
	PostgresSqlDsn string
	TrustedProxies []string
}

var DEFAULT_OPTIONS = map[string]string{
	"PORT":             "8080",
	"POSTGRES_SQL_DSN": "",
	"TRUSTED_PROXIES":  "",
}

func getOptions() *Options {
	rawOptions := parser.ArgsJoinEnv(DEFAULT_OPTIONS)

	options := new(Options)
	options.Port = rawOptions["PORT"]
	options.PostgresSqlDsn = rawOptions["POSTGRES_SQL_DSN"]
	if rawOptions["TRUSTED_PROXIES"] != "" {
		options.TrustedProxies = strings.Split(rawOptions["TRUSTED_PROXIES"], ",")
	} else {
		options.TrustedProxies = []string{}
	}

	return options
}

func safeGo(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered from panic:", r)
			}
		}()
		f()
	}()
}

func main() {
	options := getOptions()
	if len(options.PostgresSqlDsn) == 0 {
		panic(schedule_errors.InvalidArgumentError{Param: "POSTGRES_SQL_DSN", Message: "not found"})
	}
	if len(options.Port) == 0 {
		panic(schedule_errors.InvalidArgumentError{Param: "PORT", Message: "not found"})
	}

	database := pg.New(options.PostgresSqlDsn)
	agentApi := gateway.Agent{DB: database}
	batchApi := gateway.Batch{DB: database}
	jobApi := gateway.Job{DB: database}

	var authorizations, queryError = database.SelectAuthorizations()

	if queryError != nil {
		for _, authorization := range authorizations {
			if authorization.Name == "github" {
				github := authorization.Payload.(oauthGithub.Github)
				oauth.Core.AddProvider(authorization.Name, &github)
			}
		}
	}

	router := gin.Default()
	router.Use(ginsession.New())
	router.SetTrustedProxies(options.TrustedProxies)

	router.GET("/auth/:name/login", func(ctx *gin.Context) {
		name := ctx.Param("name")
		url, err := oauth.Core.GetLoginUrl(name)

		if err != nil {
			ctx.Redirect(302, "/404")
			return
		}

		ctx.Redirect(302, url)
	})

	router.GET("/auth/:name/callback", func(ctx *gin.Context) {
		name := ctx.Param("name")
		user, errUser := oauth.Core.GetUser(name, ctx.Query("code"))
		if errUser != nil {
			ctx.AbortWithError(403, &schedule_errors.ForbiddenError{})
			return
		}

		store := ginsession.FromContext(ctx)
		store.Set("userName", user.Name)
		store.Set("userEmail", user.Email)

		errStore := store.Save()

		if errStore != nil {
			ctx.AbortWithError(500, &schedule_errors.InternalServerError{Err: errStore})
			return
		}

		ctx.Redirect(302, "/")
	})

	router.GET("/auth/providers", func(ctx *gin.Context) {
		providers := oauth.Core.GetProviders()

		encoder := json.NewEncoder(ctx.Writer)
		encoder.SetEscapeHTML(false)
		encoder.Encode(gin.H{"code": 200, "data": providers})
	})

	router.GET("/api/v1/logs/:job_id", func(ctx *gin.Context) {
		limit := 0
		jobId := ctx.Param("job_id")
		lastId := ctx.Query("last_id")

		if ctx.Query("limit") != "" {
			newLimit, errAtoi := strconv.Atoi(ctx.Query("limit"))
			if errAtoi != nil {
				limit = newLimit
			}
		}

		body, errApi := agentApi.GetLogs(jobId, lastId, limit)

		if errApi != nil {
			ctx.AbortWithError(500, errApi)
			return
		}

		ctx.Data(200, "application/json", body)
	})

	router.GET("/api/v1/logs/:job_id/:id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")
		id := ctx.Param("id")

		body, err := agentApi.GetLog(jobId, id)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.Data(200, "application/json", body)
	})

	router.POST("/api/v1/pre-next/schedule/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		payload := make(map[string]interface{})
		errBind := ctx.BindJSON(&payload)

		if errBind != nil {
			ctx.AbortWithError(400, &schedule_errors.InvalidArgumentError{Param: "body", Message: errBind.Error()})
			return
		}

		data, err := batchApi.GetPreNextSchedule(name, payload)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": data})
	})

	router.POST("/api/v1/next/schedule/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")

		data, err := batchApi.GetNextSchedule(jobId)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": data})
	})

	router.POST("/api/v1/pre-next/info/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		payload := make(map[string]interface{})
		errBind := ctx.BindJSON(&payload)

		if errBind != nil {
			ctx.AbortWithError(400, &schedule_errors.InvalidArgumentError{Param: "body", Message: errBind.Error()})
			return
		}

		data, err := batchApi.GetPreNextInfo(name, payload)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": data})
	})

	router.POST("/api/v1/next/info/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")

		data, err := batchApi.GetNextInfo(jobId)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": data})
	})

	router.POST("/api/v1/progress/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")

		err := batchApi.ProgressOnce(jobId)

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": "ok"})
	})

	router.POST("/api/v1/job", func(ctx *gin.Context) {
		payload := gateway.Item{}
		errBind := ctx.BindJSON(&payload)

		if errBind != nil {
			ctx.AbortWithError(400, &schedule_errors.InvalidArgumentError{Param: "body", Message: errBind.Error()})
			return
		}

		id, errInsert := jobApi.InsertJob(payload)

		if errInsert != nil {
			ctx.AbortWithError(400, errInsert)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": id})
	})

	router.GET("/api/v1/jobs", func(ctx *gin.Context) {
		lastId := ctx.Query("last_id")
		limit, errAtoi := strconv.Atoi(ctx.Query("limit"))
		if errAtoi != nil {
			limit = 20
		}

		store := ginsession.FromContext(ctx)
		email, has := store.Get("userEmail")

		if !has {
			ctx.AbortWithError(401, &schedule_errors.UnauthorizedError{Reason: "not logged in"})
			return
		}

		body, err := database.SelectJobs(email.(string), lastId, limit)

		if err != nil {
			ctx.AbortWithError(400, err)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": body})
	})

	router.DELETE("/api/v1/job/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")

		jobApi.DeleteJob(jobId)

		ctx.JSON(200, gin.H{"code": 200, "data": "ok"})
	})

	router.GET("/api/v1/job/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")

		info, err := database.SelectJob(jobId)

		if err != nil {
			ctx.AbortWithError(400, err)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": info})
	})

	router.PUT("/api/v1/job/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")
		payload := db.Job{}
		errBind := ctx.BindJSON(&payload)

		if errBind != nil {
			ctx.AbortWithError(400, &schedule_errors.InvalidArgumentError{Param: "body", Message: errBind.Error()})
			return
		}

		_, err := database.SelectJob(jobId)

		if err != nil {
			ctx.AbortWithError(400, err)
			return
		}

		errUpdate := database.UpdateJob(jobId, payload.Name, payload.Description, payload.Author, payload.Members)

		if errUpdate != nil {
			ctx.AbortWithError(400, errUpdate)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": "ok"})
	})

	router.GET("/api/v1/action/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")

		info, err := database.SelectAction(jobId)

		if err != nil {
			ctx.AbortWithError(400, err)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": info})
	})

	router.PUT("/api/v1/action/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")
		payload := db.Action{}
		errBind := ctx.BindJSON(&payload)

		if errBind != nil {
			ctx.AbortWithError(400, &schedule_errors.InvalidArgumentError{Param: "body", Message: errBind.Error()})
			return
		}

		_, err := database.SelectAction(jobId)

		if err != nil {
			ctx.AbortWithError(400, err)
			return
		}

		errInsert := database.UpdateAction(jobId, payload.Name, payload.Type, payload.Payload)

		if errInsert != nil {
			ctx.AbortWithError(400, errInsert)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": "ok"})
	})

	router.GET("/api/v1/trigger/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")

		info, err := database.SelectTrigger(jobId)

		if err != nil {
			ctx.AbortWithError(400, err)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": info})
	})

	router.PUT("/api/v1/trigger/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")
		payload := db.Trigger{}
		errBind := ctx.BindJSON(&payload)

		if errBind != nil {
			ctx.AbortWithError(400, &schedule_errors.InvalidArgumentError{Param: "body", Message: errBind.Error()})
			return
		}

		_, err := database.SelectTrigger(jobId)

		if err != nil {
			ctx.AbortWithError(400, err)
			return
		}

		errInsert := database.UpdateTrigger(jobId, payload.Name, payload.Payload)

		if errInsert != nil {
			ctx.AbortWithError(400, errInsert)
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": "ok"})
	})
	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(404, gin.H{"code": 404, "message": "접근 할 수 없는 페이지입니다!"})
	})

	fmt.Println("Started Agent! on " + options.Port)

	router.Run(":" + options.Port)
}
