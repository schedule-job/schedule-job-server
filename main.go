package main

import (
	"fmt"
	"strings"

	parser "github.com/Sotaneum/go-args-parser"
	"github.com/gin-gonic/gin"
	ginsession "github.com/go-session/gin-session"
	"github.com/schedule-job/schedule-job-server/internal/pg"
)

type Options struct {
	Port           string
	PostgresSqlDsn string
	TrustedProxies string
	AgentUrl       string
}

var DEFAULT_OPTIONS = map[string]string{
	"PORT":             "8080",
	"POSTGRES_SQL_DSN": "",
	"TRUSTED_PROXIES":  "",
	"AGENT_URL":        "",
}

func getOptions() *Options {
	rawOptions := parser.ArgsJoinEnv(DEFAULT_OPTIONS)

	options := new(Options)
	options.Port = rawOptions["PORT"]
	options.PostgresSqlDsn = rawOptions["POSTGRES_SQL_DSN"]
	options.TrustedProxies = rawOptions["TRUSTED_PROXIES"]
	options.AgentUrl = rawOptions["AGENT_URL"]

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
	if len(options.AgentUrl) == 0 {
		panic("not found 'AGENT_URL' options")
	}

	database := pg.New(options.PostgresSqlDsn)

	router := gin.Default()
	router.Use(ginsession.New())

	if options.TrustedProxies != "" {
		trustedProxies := strings.Split(options.TrustedProxies, ",")
		router.SetTrustedProxies(trustedProxies)
	}

	router.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(404, gin.H{"code": 404, "message": "접근 할 수 없는 페이지입니다!"})
	})

	fmt.Println("Started Agent! on " + options.Port)

	router.Run(":" + options.Port)
}
