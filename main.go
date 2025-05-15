package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/file_upload/config"
	"github.com/file_upload/server"
	"github.com/file_upload/utils"
	"github.com/sirupsen/logrus"
)

func main() {

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	config, err := config.LoadConfig("config/config.json")
	if err != nil {
		log.Fatal(fmt.Errorf("error loading config file, %v", err))
	}

	utils.LoggingMechanism()
	srv := server.SrvInit(&config)

	go srv.Start()
	fmt.Println("Srv", srv)
	<-done
	logrus.Info("Gracefully Shutdown Initiate")
	srv.Stop()
	logrus.Info("Gracefully Shutdown Completed")

}
