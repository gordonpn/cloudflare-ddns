package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gordonpn/cloudflare-ddns/types"

	"github.com/jasonlvhit/gocron"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func getIPAddress() (ipAddress string, err error) {
	log.Info("Fetching external public IP")
	resp, err := http.Get("https://1.1.1.1/cdn-cgi/trace")
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "ip=") {
			ipAddress = strings.TrimPrefix(scanner.Text(), "ip=")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return ipAddress, nil
}

func fetchRecord() (types.Record, error) {
	var recordsResponse types.ListRecordsResponse
	client := http.Client{}
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=A&name=%s", os.Getenv("ZONE_ID"), os.Getenv("RECORD_NAME"))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return types.Record{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("API_TOKEN")))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return types.Record{}, err
	}
	err = json.NewDecoder(res.Body).Decode(&recordsResponse)
	if err != nil {
		return types.Record{}, err
	}

	if len(recordsResponse.Result) > 1 {
		log.Warn("More than one matching Record returned, ambiguous. Choosing first record.")
		// maybe better to update all?
	}

	return recordsResponse.Result[0], nil
}

func updateRecord(ipAddress string) error {
	log.Debugf("Updating with %s", ipAddress)
	return nil
}

func checkCriticalConfig(key string) (err error) {
	if value, exists := os.LookupEnv(key); !exists {
		log.Fatalf("Missing %s environment variable, exiting...", key)
	} else if value == "" {
		log.Fatalf("%s cannot be empty, exiting...", key)
	}
	return nil
}

func verifyConfig() {
	checkCriticalConfig("API_TOKEN")
	checkCriticalConfig("ZONE_ID")
	checkCriticalConfig("RECORD_NAME")
	if value, exists := os.LookupEnv("RECORD_TTL"); !exists {
		log.Warn("Missing RECORD_TTL environment variable")
		log.Warn("Will use \"1\" as RECORD_TTL")
		os.Setenv("RECORD_TTL", "1")
	} else if _, err := strconv.Atoi(value); err != nil {
		log.Warnf("RECORD_TTL: %q doesn't looks like a number", value)
		log.Warn("Will use \"1\" as RECORD_TTL")
		os.Setenv("RECORD_TTL", "1")
	}
	log.Infof("Running in \"%s\" environment", os.Getenv("APP_ENV"))
}

func task() {
	log.Info("Starting task")
	ipAddress, err := getIPAddress()
	// maybe wait and try again?
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{"ipAddress": ipAddress}).Info("Fetched external public IP")
	currentIP, _ := fetchRecord()
	// handle error, retries or return?
	if currentIP.Content == ipAddress {
		log.WithFields(log.Fields{"currentIP": currentIP.Content}).Info("IP address has not changed")
		// nothing to do, return early
	}
	updateRecord(ipAddress)
	// refactor ip address variable names to be more concise
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

	verifyConfig()

	periodicPtr := flag.Bool("periodic", false, "Run periodically, automatically")

	flag.Parse()

	task()

	log.WithFields(log.Fields{"periodic": *periodicPtr}).Info("Periodic flag")
	if *periodicPtr {
		gocron.Every(2).Hours().From(gocron.NextTick()).Do(task)
		<-gocron.Start()
	}
}
