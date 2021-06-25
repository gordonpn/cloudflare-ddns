package main

import (
	"flag"
	"os"

	"github.com/gordonpn/cloudflare-ddns/internal/api"
	"github.com/gordonpn/cloudflare-ddns/internal/config"
	"github.com/gordonpn/cloudflare-ddns/internal/healthchecks"

	"github.com/jasonlvhit/gocron"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func task() {
	healthchecks.SignalStart()
	log.Info("Starting task")
	externalAddress, err := api.GetIPAddress()
	if err != nil {
		healthchecks.SignalFailure(err.Error())
		log.Fatal(err)
	}

	log.WithFields(log.Fields{"ipAddress": externalAddress}).Debug("Fetched external public IP")
	currentRecord, err := api.FetchRecord()
	if err != nil {
		healthchecks.SignalFailure(err.Error())
		log.Fatal(err)
	}
	if currentRecord.Content != externalAddress || os.Getenv("APP_ENV") != "production" {
		api.UpdateRecord(currentRecord, externalAddress)
	} else {
		log.WithFields(log.Fields{"currentIP": currentRecord.Content}).Debug("IP address has not changed")
		log.Debug("Nothing do to")
	}
	log.Info("Task completed")
	healthchecks.SignalEnd()
}

func main() {
	appEnv, _ := os.LookupEnv("APP_ENV")
	log.SetLevel(log.InfoLevel)
	if appEnv != "production" {
		err := godotenv.Load()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Fatal("Problem with loading .env file")
		}
		log.SetLevel(log.DebugLevel)
	}

	config.VerifyConfig()

	periodicPtr := flag.Bool("periodic", false, "Run periodically, automatically")

	flag.Parse()

	task()

	log.WithFields(log.Fields{"periodic": *periodicPtr}).Debug("Periodic flag")
	if *periodicPtr {
		log.Info("Setting schedule")
		gocron.Every(10).Minutes().From(gocron.NextTick()).Do(task)
		<-gocron.Start()
	}
}
