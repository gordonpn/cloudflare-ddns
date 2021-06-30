package main

import (
	"flag"
	"os"
	"time"

	"github.com/gordonpn/cloudflare-ddns/internal/api"
	"github.com/gordonpn/cloudflare-ddns/internal/config"
	"github.com/gordonpn/cloudflare-ddns/internal/healthchecks"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var memoizedIPAddress = ""

func task() {
	healthchecks.SignalStart()
	defer healthchecks.SignalEnd()
	defer log.Info("Task completed")
	log.Info("Starting task")

	externalAddress, err := api.GetIPAddress()
	if err != nil {
		healthchecks.SignalFailure(err.Error())
		log.Fatal(err)
	}
	log.WithFields(log.Fields{"ipAddress": externalAddress}).Debug("Fetched external public IP")

	if externalAddress == memoizedIPAddress {
		log.WithFields(log.Fields{"ipAddress": memoizedIPAddress}).Debug("IP address in memory")
		return
	}

	memoizedIPAddress = externalAddress

	currentRecord, err := api.FetchRecord()
	if err != nil {
		healthchecks.SignalFailure(err.Error())
		log.Fatal(err)
	}
	log.WithFields(log.Fields{"ipAddress": currentRecord.Content}).Debug("Current record IP address")

	if currentRecord.Content == externalAddress && os.Getenv("APP_ENV") == "production" {
		log.Debug("IP address already up to date, nothing to do")
		return
	}

	err = api.UpdateRecord(currentRecord, externalAddress)
	if err != nil {
		healthchecks.SignalFailure(err.Error())
		log.Fatal(err)
	}
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

	log.WithFields(log.Fields{"periodic": *periodicPtr}).Debug("Periodic flag")
	if *periodicPtr {
		log.Info("Setting schedule")
		s := gocron.NewScheduler(time.UTC)

		_, err := s.Every(10).Minutes().Do(task)
		if err != nil {
			healthchecks.SignalFailure(err.Error())
			log.Fatal(err)
		}
		s.StartBlocking()
	} else {
		task()
	}
}
