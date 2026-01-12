package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type result struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	ErrorMsg   string `json:"error,omitempty"`
	Timestamp  string `json:"timestamp"`
}

func main() {
	links := os.Args[1:]

	if len(links) == 0 {
		fmt.Println("Usage: go run main.go <url1> <url2> ...")
		os.Exit(1)
	}

	resultsChannel := make(chan result)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, link := range links {
		go checkLink(link, resultsChannel, client)
	}

	// Step 10: Log file open karna (Append mode mein)
	// Agar file nahi hai toh Create hogi, hai toh naya data peeche judta jayega
	f, err := os.OpenFile("log_results.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Fatal: Could not open log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	for res := range resultsChannel {
		jsonData, err := json.Marshal(res)
		if err != nil {
			fmt.Printf("Error marshaling: %v\n", err)
			continue
		}

		// Terminal par bhi dikhao
		fmt.Printf("Logging: %s\n", string(jsonData))

		// Step 10: File mein likhna
		// Har result ke baad \n (newline) dena zaroori hai taaki JSONL format rahe
		_, err = f.WriteString(string(jsonData) + "\n")
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
		}

		go func(l string) {
			time.Sleep(10 * time.Second)
			checkLink(l, resultsChannel, client)
		}(res.URL)
	}
}

func checkLink(link string, c chan result, client *http.Client) {
	res := result{
		URL:       link,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	resp, err := client.Get(link)
	if err != nil {
		res.ErrorMsg = err.Error()
		c <- res
		return
	}
	defer resp.Body.Close()

	res.StatusCode = resp.StatusCode
	c <- res
}