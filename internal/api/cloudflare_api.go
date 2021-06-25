package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gordonpn/cloudflare-ddns/internal/types"

	log "github.com/sirupsen/logrus"
)

func GetIPAddress() (address string, err error) {
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

func addAuthHeaders(req *http.Request) *http.Request {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("API_TOKEN")))
	req.Header.Set("Content-Type", "application/json")

	return req
}

func FetchRecord() (types.Record, error) {
	var apiResponse types.ListRecordsResponse
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=A&name=%s", os.Getenv("ZONE_ID"), os.Getenv("RECORD_NAME"))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return types.Record{}, err
	}
	req = addAuthHeaders(req)
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

func UpdateRecord(record types.Record, externalAddress string) error {
	log.Debug("Updating record")
	var apiResponse types.UpdateRecordResponse
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", os.Getenv("ZONE_ID"), record.ID)
	ttl, _ := strconv.Atoi(os.Getenv("RECORD_TTL"))
	reqBody, _ := json.Marshal(types.UpdateRecordRequest{Type: "A", Name: "gordon-pn.com", Content: record.Content, TTL: ttl, Proxied: true})
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req = addAuthHeaders(req)
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
