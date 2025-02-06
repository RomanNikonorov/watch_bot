package watch

import (
	"log"
	"net/http"
	"time"
)

type URLChecker interface {
	IsUrlOk(url string, unhealthyThreshold int, unhealthyDelay int, client HTTPClient) bool
}

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

type RealURLChecker struct{}

func (r RealURLChecker) IsUrlOk(url string, unhealthyThreshold int, unhealthyDelay int, client HTTPClient) bool {
	status := checkURLOnce(url, client)
	if status {
		return true
	}
	for i := 0; i < unhealthyThreshold; i++ {
		time.Sleep(time.Duration(unhealthyDelay) * time.Second)
		status = checkURLOnce(url, client)
		if status {
			return true
		}
	}
	return false
}

func checkURLOnce(url string, client HTTPClient) bool {
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("failed to get URL: %v", err)
		return false
	}
	if resp == nil || resp.Body == nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}
