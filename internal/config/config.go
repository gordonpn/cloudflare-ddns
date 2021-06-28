package config

import (
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func checkCriticalConfig(key string) {
	if value, exists := os.LookupEnv(key); !exists {
		log.Fatalf("Missing %s environment variable, exiting...", key)
	} else if value == "" {
		log.Fatalf("%s cannot be empty, exiting...", key)
	}
}

func VerifyConfig() {
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
	if os.Getenv("APP_ENV") == "production" {
		checkCriticalConfig("HC_URL")
	}
}
