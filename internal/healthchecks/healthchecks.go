package healthchecks

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func SignalStart() {
	if os.Getenv("APP_ENV") != "production" {
		return
	}
	log.Debug("Signaling start")
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/%s", os.Getenv("HC_URL"), "start"), nil)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	_, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warn("Error occurred while signalling Healthcheck start")
	}
}

func SignalEnd() {
	if os.Getenv("APP_ENV") != "production" {
		return
	}
	log.Debug("Signaling end")
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	req, _ := http.NewRequest("GET", os.Getenv("HC_URL"), nil)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	_, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warn("Error occurred while signalling Healthcheck end")
	}
}

func SignalFailure(errorMsg string) {
	if os.Getenv("APP_ENV") != "production" {
		return
	}
	log.Debug("Signaling failure")
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/%s", os.Getenv("HC_URL"), "fail"), strings.NewReader(errorMsg))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	_, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warn("Error occurred while signalling Healthcheck failure")
	}
}
