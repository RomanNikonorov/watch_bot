package watch

import (
	"crypto/tls"
	"log"
	"net/http"
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
