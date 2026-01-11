package main

import (
	"fmt"
	"net/http"
	"os"
	"time" // Timestamp ke liye
)

// Step 6: Ek mast struct banao result hold karne ke liye
// Bilkul waise hi jaise tu DynamoDB ke liye schema banata hai
type result struct {
	url        string
	statusCode int
	status     string
	err        error
	timestamp  string
}

func main() {
	links := os.Args[1:]

	if len(links) == 0 {
		fmt.Println("Usage: go run main.go <url1> <url2> ...")
		os.Exit(1)
	}

	// Ab channel string ka nahi, hamare custom 'result' struct ka hoga
	resultsChannel := make(chan result)

	for _, link := range links {
		go checkLink(link, resultsChannel)
	}

	// Results collect karo
	for i := 0; i < len(links); i++ {
		res := <-resultsChannel

		// Ab humare paas structured data hai, hum kuch bhi kar sakte hain!
		if res.err != nil {
			fmt.Printf("[%s] ❌ %s | Error: %v\n", res.timestamp, res.url, res.err)
		} else {
			fmt.Printf("[%s] ✅ %s | Status: %d (%s)\n", res.timestamp, res.url, res.statusCode, res.status)
		}
	}
}

func checkLink(link string, c chan result) {
	// Har result ke liye ek naya struct instance taiyar karo
	res := result{
		url:       link,
		timestamp: time.Now().Format("15:04:05"), // Current time (HH:MM:SS)
	}

	resp, err := http.Get(link)
	if err != nil {
		res.err = err
		c <- res // Poora struct pipe mein daal do
		return
	}
	defer resp.Body.Close()

	// Struct ki fields bharo
	res.statusCode = resp.StatusCode
	res.status = resp.Status
	
	c <- res
}