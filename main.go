package main

import (
	"encoding/json"
	"fmt"
	"strings"

	parser "github.com/Sotaneum/go-args-parser"
	"github.com/gin-gonic/gin"
	ginsession "github.com/go-session/gin-session"
	"github.com/schedule-job/schedule-job-server/internal/oauth"
	"github.com/schedule-job/schedule-job-server/internal/pg"
)

type Options struct {
	Port           string
	PostgresSqlDsn string
	TrustedProxies string
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
	options.TrustedProxies = rawOptions["TRUSTED_PROXIES"]

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

	router := gin.Default()
	router.Use(ginsession.New())

	if options.TrustedProxies != "" {
		trustedProxies := strings.Split(options.TrustedProxies, ",")
		router.SetTrustedProxies(trustedProxies)
	}

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

	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(404, gin.H{"code": 404, "message": "접근 할 수 없는 페이지입니다!"})
	})

	fmt.Println("Started Agent! on " + options.Port)

	router.Run(":" + options.Port)
}
