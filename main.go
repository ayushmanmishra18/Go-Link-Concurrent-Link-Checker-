package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

type result struct {
	url        string
	statusCode int
	err        error
	timestamp  string
}

func main() {
	links := os.Args[1:]

	if len(links) == 0 {
		fmt.Println("Usage: go run main.go <url1> <url2> ...")
		os.Exit(1)
	}

	resultsChannel := make(chan result)

	// Step 8: Custom HTTP Client with Timeout
	// Go ka default client infinite wait kar sakta hai, jo ki bahut risky hai.
	// Hum yahan 5-second ka timeout set kar rahe hain.
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, link := range links {
		go checkLink(link, resultsChannel, client)
	}

	for res := range resultsChannel {
		if res.err != nil {
			// Ab agar timeout hoga, toh yahan "context deadline exceeded" dikhega
			fmt.Printf("[%s] ❌ %s | Status: DOWN/TIMEOUT | Error: %v\n", res.timestamp, res.url, res.err)
		} else {
			fmt.Printf("[%s] ✅ %s | Status: %d\n", res.timestamp, res.url, res.statusCode)
		}

		go func(l string) {
			time.Sleep(10 * time.Second) // Thoda aur chill gap
			checkLink(l, resultsChannel, client)
		}(res.url)
	}
}

// Ab hum function mein client bhi pass kar rahe hain
func checkLink(link string, c chan result, client *http.Client) {
	res := result{
		url:       link,
		timestamp: time.Now().Format("15:04:05"),
	}

	// Default http.Get ki jagah ab hum apna timed-out client use karenge
	resp, err := client.Get(link)
	if err != nil {
		res.err = err
		c <- res
		return
	}
	defer resp.Body.Close()

	res.statusCode = resp.StatusCode
	c <- res
}