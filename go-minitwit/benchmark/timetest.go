package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Endpoint struct {
	Method      string
	URL         string
	Body        interface{}
	IterateUser bool
}

// There order matters here so be careful with ;)
var endpoints = []Endpoint{
	{"POST", "/api/register", nil, false},
	{"POST", "/api/msgs/testTime", map[string]string{"content": "testTime content"}, true},
	{"POST", "/api/fllws/testTime", map[string]string{"follow": "testTime1"}, true},
	{"GET", "/api/fllws/testTime2", nil, false},
	{"GET", "/api/msgs/testTime1", nil, false},
	{"GET", "/api/msgs", nil, false},
	{"GET", "/api/latest", nil, false},
}

func MeasureResponseTime(baseURL string, numRequests int) {
	file, err := os.Create("response_times.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	for _, endpoint := range endpoints {
		var totalResponseTime time.Duration
		var start time.Time

		for i := 0; i < numRequests; i++ {
			var route string
			if endpoint.IterateUser {
				route = endpoint.URL + strconv.Itoa(i)
			} else {
				route = endpoint.URL
			}
			if endpoint.Method == "GET" {
				start = time.Now()
				_, err := http.Get(baseURL + route)
				if err != nil {
					fmt.Printf("Error occurred while making GET request to %s: %v\n", endpoint.URL, err)
					continue
				}
			} else if endpoint.Method == "POST" {
				if endpoint.URL == "/api/register" {
					username := "testTime" + strconv.Itoa(i)
					email := "test_email" + strconv.Itoa(i) + "@example.com"
					pwd := "test_password" + strconv.Itoa(i)
					body := map[string]string{"username": username, "email": email, "pwd": pwd}
					jsonBody, err := json.Marshal(body)
					if err != nil {
						fmt.Printf("Error marshaling JSON for POST request to %s: %v\n", endpoint.URL, err)
						continue
					}
					start = time.Now()
					_, err = http.Post(baseURL+route, "application/json", bytes.NewBuffer(jsonBody))
					if err != nil {
						fmt.Printf("Error occurred while making POST request to %s: %v\n", endpoint.URL, err)
						continue
					}
				} else {
					jsonBody, err := json.Marshal(endpoint.Body)
					if err != nil {
						fmt.Printf("Error marshaling JSON for POST request to %s: %v\n", endpoint.URL, err)
						continue
					}
					start = time.Now()
					_, err = http.Post(baseURL+route, "application/json", bytes.NewBuffer(jsonBody))
					if err != nil {
						fmt.Printf("Error occurred while making POST request to %s: %v\n", endpoint.URL, err)
						continue
					}
				}
			}
			totalResponseTime += time.Since(start)
		}
		avgResponseTime := totalResponseTime / time.Duration(numRequests)
		line := fmt.Sprintf("Average response time for %s: %v\n", endpoint.URL, avgResponseTime)
		fmt.Print(line)
		if _, err := file.WriteString(line); err != nil {
			fmt.Println("Error writing to file:", err)
		}
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("To use this executable please insert also the <base_URL> and <requests_number>")
		return
	}
	baseURL := os.Args[1]
	numRequests, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Error parsing the requests number:", err)
		return
	}
	MeasureResponseTime(baseURL, numRequests)
}
