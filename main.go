package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type FirewallRequest struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

var (
	apiURLs = []string{
		"https://chuanxiang-backend.int.devcloudhub.org/api/public/firewall",
		"https://chuanxiang-backend.prod.devcloudhub.org/api/public/firewall",
	}
)

const (
	ipFetchURL  = "https://ifconfig.me/ip"
	requestFreq = 5 * time.Minute // Adjust as needed
)

func getPublicIP() (string, error) {
	resp, err := http.Get(ipFetchURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func postFirewallRule(url string, ip string, apiKey string) error {
	data := FirewallRequest{Type: "Allow", Address: ip}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("accept", "application/hal+json")
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Response: %s\n", body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed with status: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	apiKey := flag.String("apikey", "", "API Key for authentication")
	flag.Parse()

	if *apiKey == "" {
		fmt.Println("API key is required. Use -apikey=<your_api_key>")
		os.Exit(1)
	}

	for {
		ip, err := getPublicIP()
		if err != nil {
			fmt.Println("Error fetching IP:", err)
		} else {
			fmt.Println("Public IP:", ip)
			for _, url := range apiURLs {
				err = postFirewallRule(url, ip, *apiKey)
				if err != nil {
					errMessage := fmt.Sprintf("Error posting firewall rule to %s: %v", url, err)
					fmt.Println(errMessage)
				}

			}

		}
		time.Sleep(requestFreq)
	}
}
