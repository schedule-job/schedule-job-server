package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	parser "github.com/Sotaneum/go-args-parser"
	"github.com/gin-gonic/gin"
	ginsession "github.com/go-session/gin-session"
	"github.com/schedule-job/schedule-job-server/internal/agent"
	"github.com/schedule-job/schedule-job-server/internal/batch"
	"github.com/schedule-job/schedule-job-server/internal/job"
	"github.com/schedule-job/schedule-job-server/internal/oauth"
	"github.com/schedule-job/schedule-job-server/internal/pg"
)

type Options struct {
	Port           string
	PostgresSqlDsn string
	TrustedProxies []string
	AgentUrls      []string
	BatchUrls      []string
}

var DEFAULT_OPTIONS = map[string]string{
	"PORT":             "8080",
	"POSTGRES_SQL_DSN": "",
	"TRUSTED_PROXIES":  "",
	"AGENT_URLS":       "",
	"BATCH_URLS":       "",
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
	if rawOptions["AGENT_URLS"] != "" {
		options.AgentUrls = strings.Split(rawOptions["AGENT_URLS"], ",")
	} else {
		options.AgentUrls = []string{}
	}
	if rawOptions["BATCH_URLS"] != "" {
		options.BatchUrls = strings.Split(rawOptions["BATCH_URLS"], ",")
	} else {
		options.BatchUrls = []string{}
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
		panic("not found 'POSTGRES_SQL_DSN' options")
	}
	if len(options.Port) == 0 {
		panic("not found 'PORT' options")
	}

	database := pg.New(options.PostgresSqlDsn)

	agentApi := agent.Agent{}
	agentApi.SetAgentUrls(options.AgentUrls)

	batchApi := batch.Batch{}
	batchApi.SetBatchUrls(options.BatchUrls)

	jobApi := job.Job{}
	jobApi.SetDatabase(database)

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
		user, userErr := oauth.Core.GetUser(name, ctx.Query("code"))
		if userErr != nil {
			ctx.AbortWithError(500, userErr)
			return
		}

		store := ginsession.FromContext(ctx)
		store.Set("userName", user.Name)
		store.Set("userEmail", user.Email)

		storeErr := store.Save()

		if storeErr != nil {
			ctx.AbortWithError(500, storeErr)
			return
		}

		ctx.Redirect(302, "/")
	})

	router.GET("/auth/providers", func(ctx *gin.Context) {
		providers, err := oauth.Core.GetProviders()

		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}

		encoder := json.NewEncoder(ctx.Writer)
		encoder.SetEscapeHTML(false)
		encoder.Encode(gin.H{"code": 200, "data": providers})
	})

	router.GET("/api/v1/logs/:job_id", func(ctx *gin.Context) {
		limit := 0
		jobId := ctx.Param("job_id")
		lastId := ctx.Query("last_id")

		if ctx.Query("limit") != "" {
			cnvI, err := strconv.Atoi(ctx.Query("limit"))
			if err != nil {
				limit = cnvI
			}
		}

		body, err := agentApi.GetLogs(jobId, lastId, limit)

		if err != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		ctx.Data(200, "application/json", body)
	})

	router.GET("/api/v1/logs/:job_id/:id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")
		id := ctx.Param("id")

		body, err := agentApi.GetLog(jobId, id)

		if err != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		ctx.Data(200, "application/json", body)
	})

	router.POST("/api/v1/pre-next/schedule/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		payload := make(map[string]interface{})
		bindErr := ctx.BindJSON(&payload)

		if bindErr != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": bindErr.Error()})
			return
		}

		data, err := batchApi.GetPreNextSchedule(name, payload)

		if err != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": data})
	})

	router.POST("/api/v1/next/schedule/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")

		data, err := batchApi.GetNextSchedule(jobId)

		if err != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": data})
	})

	router.POST("/api/v1/pre-next/info/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		payload := make(map[string]interface{})
		bindErr := ctx.BindJSON(&payload)

		if bindErr != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": bindErr.Error()})
			return
		}

		data, err := batchApi.GetPreNextInfo(name, payload)

		if err != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": data})
	})

	router.POST("/api/v1/next/info/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")

		data, err := batchApi.GetNextInfo(jobId)

		if err != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": data})
	})

	router.POST("/api/v1/progress/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")

		err := batchApi.ProgressOnce(jobId)

		if err != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": "ok"})
	})

	router.POST("/api/v1/job", func(ctx *gin.Context) {
		payload := job.InsertItem{}
		bindErr := ctx.BindJSON(&payload)

		if bindErr != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": bindErr.Error()})
			return
		}

		id, insertErr := jobApi.InsertJob(payload)

		if insertErr != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": insertErr.Error()})
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": id})
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
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": info})
	})

	router.PUT("/api/v1/job/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")
		payload := job.Info{}
		bindErr := ctx.BindJSON(&payload)

		if bindErr != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": bindErr.Error()})
			return
		}

		_, err := database.SelectJob(jobId)

		if err != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		updateErr := database.UpdateJob(jobId, payload.Name, payload.Description, payload.Author, payload.Members)

		if updateErr != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": updateErr.Error()})
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": "ok"})
	})

	router.GET("/api/v1/action/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")

		info, err := database.SelectAction(jobId)

		if err != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": info})
	})

	router.PUT("/api/v1/action/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")
		payload := job.Item{}
		bindErr := ctx.BindJSON(&payload)

		if bindErr != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": bindErr.Error()})
			return
		}

		_, err := database.SelectAction(jobId)

		if err != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		insertErr := database.UpdateAction(jobId, payload.Name, payload.Payload)

		if insertErr != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": insertErr.Error()})
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": "ok"})
	})

	router.GET("/api/v1/trigger/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")

		info, err := database.SelectTrigger(jobId)

		if err != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"code": 200, "data": info})
	})

	router.PUT("/api/v1/trigger/:job_id", func(ctx *gin.Context) {
		jobId := ctx.Param("job_id")
		payload := job.Item{}
		bindErr := ctx.BindJSON(&payload)

		if bindErr != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": bindErr.Error()})
			return
		}

		_, err := database.SelectTrigger(jobId)

		if err != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": err.Error()})
			return
		}

		insertErr := database.UpdateTrigger(jobId, payload.Name, payload.Payload)

		if insertErr != nil {
			ctx.JSON(400, gin.H{"code": 400, "message": insertErr.Error()})
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
