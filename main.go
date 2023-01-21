package main

import (
	"flag"
	"fmt"
	"log"
	"paydex/config"
	"paydex/pkg/logger"
	"paydex/pkg/version"
	"paydex/services"
	"paydex/worker"

	"github.com/hibiken/asynq"
)

func main() {
	var loc string
	flag.StringVar(&loc, "config", "config file", "provide config file location")

	flag.Parse()

	l := logger.GetLogger()
	log.Printf("App Version [%s]", version.Get())

	conf, err := config.MustLoad(loc)
	if err != nil {
		log.Fatal(err)
	}
	dsn := fmt.Sprintf("%s:%s", conf.Redis.Address, conf.Redis.Port)
	workerService := worker.NewRedisTaskDistributor(asynq.RedisClientOpt{Addr: dsn})
	if err != nil {
		log.Fatal(err)
	}
	server := services.NewServer(workerService, &conf, l)

	go func() {
		if errx := server.RunGrpcServer(); errx != nil {
			log.Fatal(errx)
		}
	}()

	if errx := server.RunHTTPServer(); errx != nil {
		log.Fatal(errx)
	}
}
