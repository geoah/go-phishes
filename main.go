package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const phishesJson = "./verified_online.json"
const dataDir = "/tmp/phishes-go"
const concurrency = 5

type Phish struct {
	Details []struct {
		AnnouncingNetwork string `json:"announcing_network"`
		CidrBlock         string `json:"cidr_block"`
		Country           string `json:"country"`
		DetailTime        string `json:"detail_time"`
		IpAddress         string `json:"ip_address"`
		Rir               string `json:"rir"`
	} `json:"details"`
	Online           string `json:"online"`
	PhishDetailURL   string `json:"phish_detail_url"`
	PhishID          string `json:"phish_id"`
	SubmissionTime   string `json:"submission_time"`
	Target           string `json:"target"`
	URL              string `json:"url"`
	VerificationTime string `json:"verification_time"`
	Verified         string `json:"verified"`
}

type Wrapper struct {
	Body     string
	Response http.Response
}

func main() {
	file, e := ioutil.ReadFile(phishesJson)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	var phishes []Phish
	json.Unmarshal(file, &phishes)

	sem := make(chan bool, concurrency)

	for _, phish := range phishes {
		sem <- true
		go func(phish *Phish) {
			defer func() { <-sem }()
			resp, err := http.Get(phish.URL)

			if err == nil {
				body, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					bodyStr := string(body[:])
					wrapper := Wrapper{Body: bodyStr, Response: *resp}
					bytes, err := json.Marshal(wrapper)
					if err == nil {
						filename := fmt.Sprintf("%s/%s.json", dataDir, phish.PhishID)
						err := ioutil.WriteFile(filename, bytes, 0644)
						if err == nil {
							log.Printf("Retrieved %s", phish.URL)
						} else {
							log.Printf("Could save wrapper for %s from %s with %s", phish.PhishID, phish.URL, err)
						}
					} else {
						log.Printf("Could marshal object for %s from %s with %s", phish.PhishID, phish.URL, err)
					}
				} else {
					log.Printf("Could get complete body for %s from %s with %s", phish.PhishID, phish.URL, err)
				}
			} else {
				log.Printf("Could not read URL for %s from %s with %s", phish.PhishID, phish.URL, err)
			}
		}(&phish)
	}
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
}
