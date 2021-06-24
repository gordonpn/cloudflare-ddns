package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
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

func getIPAddress() (address string, err error) {
	log.Debug("Fetching external public IP")
	resp, err := http.Get("https://1.1.1.1/cdn-cgi/trace")
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "ip=") {
			address = strings.TrimPrefix(scanner.Text(), "ip=")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return address, nil
}

func fetchRecord() (types.Record, error) {
	var apiResponse types.ListRecordsResponse
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
	err = json.NewDecoder(res.Body).Decode(&apiResponse)
	if err != nil {
		return types.Record{}, err
	}

	if len(apiResponse.Result) > 1 {
		log.Warn("More than one matching Record returned, ambiguous. Choosing first record.")
	}

	return apiResponse.Result[0], nil
}

func updateRecord(record types.Record, externalAddress string) error {
	log.Debug("Updating record")
	var apiResponse types.UpdateRecordResponse
	client := http.Client{}
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", os.Getenv("ZONE_ID"), record.ID)
	ttl, _ := strconv.Atoi(os.Getenv("RECORD_TTL"))
	reqBody, _ := json.Marshal(types.UpdateRecordRequest{Type: "A", Name: "gordon-pn.com", Content: record.Content, TTL: ttl, Proxied: true})
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("API_TOKEN")))
	req.Header.Set("Content-Type", "application/json")
	// TODO: common headers are set
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	err = json.NewDecoder(res.Body).Decode(&apiResponse)
	if err != nil {
		return err
	}
	if !apiResponse.Success {
		return errors.New("record update was not successful")
	}
	log.WithFields(log.Fields{"Update success": apiResponse.Success}).Debug("Update API response")
	log.Debug("Done updating record")
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
	externalAddress, err := getIPAddress()
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{"ipAddress": externalAddress}).Debug("Fetched external public IP")
	currentRecord, err := fetchRecord()
	if err != nil {
		log.Fatal(err)
	}
	if currentRecord.Content != externalAddress || os.Getenv("APP_ENV") != "production" {
		updateRecord(currentRecord, externalAddress)
	} else {
		log.WithFields(log.Fields{"currentIP": currentRecord.Content}).Debug("IP address has not changed")
		log.Debug("Nothing do to")
	}
	log.Info("Task completed")
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

	log.WithFields(log.Fields{"periodic": *periodicPtr}).Debug("Periodic flag")
	if *periodicPtr {
		log.Info("Setting schedule")
		gocron.Every(10).Minutes().From(gocron.NextTick()).Do(task)
		<-gocron.Start()
	}
}
