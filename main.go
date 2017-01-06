package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/YoSarin/gm-mentions-bot/lib"
	"github.com/garyburd/redigo/redis"
	"github.com/julienschmidt/httprouter"
	"github.com/martin-reznik/logger"
)

var log *logger.Log

type Router struct {
	*httprouter.Router
	*logger.Log
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// TODO: validate source
	r.Router.ServeHTTP(w, req)
}

func main() {
	flagDebug := flag.Bool("debug", false, "log verbosity: debug")
	flagPort := flag.Int("port", 8080, "Port to run server at")
	flagRedis := flag.String("redis", ":6379", "Address (including port) to redis server")

	flag.Parse()
	log = logger.NewLog(func(line *logger.LogLine) {
		line.Print()
	}, &logger.Config{GoRoutinesLogTicker: 5 * time.Second})

	if *flagDebug {
		log.LogSeverity[logger.DEBUG] = true
	}

	web := &Router{
		httprouter.New(),
		log,
	}

	rdb, err := redis.Dial("tcp", *flagRedis)
	if err != nil {
		panic(err)
	}
	defer rdb.Close()

	handler := lib.NewHandler(
		log,
		lib.NewStorage(
			rdb,
			log,
		),
	)

	web.GET("/message/:token", handler.ProcessMessage)

	panic(http.ListenAndServe(fmt.Sprintf(":%v", *flagPort), web))
}
