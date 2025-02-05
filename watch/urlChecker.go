package watch

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"
)

type URLChecker interface {
	IsUrlOk(url string, unhealthyThreshold int, unhealthyDelay int) bool
}

type RealURLChecker struct{}

func (r RealURLChecker) IsUrlOk(url string, unhealthyThreshold int, unhealthyDelay int) bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	status := checkURLOnce(url, client)

	if !status {
		for i := 0; i < unhealthyThreshold; i++ {
			time.Sleep(time.Duration(unhealthyDelay) * time.Second)
			status = checkURLOnce(url, client)
			if status {
				return true
			}
		}
	}
	return true
}

func checkURLOnce(url string, client *http.Client) bool {
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("failed to get URL: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}
