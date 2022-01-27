package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/caseymrm/menuet"
	"github.com/pkg/errors"
)

const (
	refreshingPeriod = 5 * time.Second
	timeoutPeriod    = 10 * time.Second
)

func refreshLoop() {
	lastIP := ""
	httpClient := createHTTPClient()
	for range time.Tick(refreshingPeriod) {
		menuet.App().SetMenuState(&menuet.MenuState{Title: "Checking IP..."})

		ip, err := checkIP(httpClient)
		if err != nil {
			menuet.App().SetMenuState(&menuet.MenuState{Title: "Error checking IP"})
			log.Println(err)
			continue
		}

		menuet.App().SetMenuState(&menuet.MenuState{Title: "IP: " + ip})

		if lastIP == ip {
			continue
		}

		lastIP = ip
		log.Printf("IP changed: %s", ip)
	}
}

func createHTTPClient() *http.Client {
	transport := http.Transport{
		DisableKeepAlives: true,
	}
	return &http.Client{
		Transport: &transport,
		Timeout:   timeoutPeriod,
	}
}

func checkIP(httpClient *http.Client) (string, error) {
	req, err := http.NewRequest("GET", "https://ip.me", nil)
	if err != nil {
		return "", errors.Wrap(err, "http.NewRequest failed")
	}
	req.Header.Set("User-Agent", "curl/7.64.1")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "http.Get failed")
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "ioutil.ReadAll failed")
	}

	t := bytes.TrimSpace(b)
	parsed := net.ParseIP(string(t))
	if parsed == nil {
		return "", fmt.Errorf("net.ParseIP failed (IP address: %s)", string(t))
	}

	return parsed.String(), nil
}

func main() {
	menuet.App().Label = "com.github.mtojek.ip-menubar"
	menuet.App().SetMenuState(&menuet.MenuState{Title: "Checking IP..."})

	go refreshLoop()
	menuet.App().RunApplication()
}
