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
	client := &http.Client{Timeout: 5 * time.Second}

	for _, link := range links {
		go checkLink(link, resultsChannel, client)
	}

	// Step 11: Ticker setup (Har 30 second mein trigger hoga)
	// Ye background mein ek ghadi ki tarah chalta rahega
	uploadTicker := time.NewTicker(30 * time.Second)

	f, _ := os.OpenFile("log_results.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	// Step 11: Select Statement ka introduction
	// Ab humein do cheezein listen karni hain: results aur timer
	for {
		select {
		case res := <-resultsChannel:
			// Purana logic: JSON banao aur file mein likho
			jsonData, _ := json.Marshal(res)
			f.WriteString(string(jsonData) + "\n")
			fmt.Printf("Logged: %s\n", res.URL)

			go func(l string) {
				time.Sleep(10 * time.Second)
				checkLink(l, resultsChannel, client)
			}(res.URL)

		case <-uploadTicker.C:
			// Timer baj gaya! Ab S3 par "upload" karne ka natak karte hain
			fmt.Println("\n--- ☁️  AWS S3 UPLOAD TRIGGERED ---")
			uploadToS3("log_results.jsonl")
			
			// Upload ke baad file ko truncate (khaali) kar dena chahiye
			// f.Truncate(0) // Asli logic mein hum ye karte hain
		}
	}
}

// Step 11: S3 Upload ka Skeleton
// Jab tu real AWS SDK seekhega, toh yahan S3 PutObject ka code aayega
func uploadToS3(filename string) {
	fmt.Printf("Uploading %s to S3 Bucket: 'blood-bank-logs'...\n", filename)
	// Yahan hum simulation kar rahe hain
	time.Sleep(1 * time.Second)
	fmt.Println("✅ Upload Successful! S3 Storage Updated.")
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