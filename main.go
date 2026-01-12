package main

import (
	"encoding/json" // JSON conversion ke liye
	"fmt"
	"net/http"
	"os"
	"time"
)

// Step 9: Struct mein JSON tags add karna
// Jab hum Marshal karenge, toh Go in tags ko dekh kar JSON keys banayega.
// Note: Humne 'Error' ko string mein badla hai kyunki 'error' interface JSON mein marshal nahi hota.
type result struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	ErrorMsg   string `json:"error,omitempty"` // omitempty ka matlab agar error nahi hai toh JSON mein mat dikhao
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

	for res := range resultsChannel {
		// Step 9: Struct ko JSON mein badalna (Marshaling)
		// json.Marshal returns: byte slice ([]byte) and error
		jsonData, err := json.Marshal(res)
		if err != nil {
			fmt.Printf("Error marshaling to JSON: %v\n", err)
			continue
		}

		// Ab hum terminal par poora JSON string print karenge
		// %s isliye use kiya kyunki jsonData ek byte slice hai
		fmt.Printf("DATA: %s\n", string(jsonData))

		go func(l string) {
			time.Sleep(10 * time.Second)
			checkLink(l, resultsChannel, client)
		}(res.URL)
	}
}

func checkLink(link string, c chan result, client *http.Client) {
	res := result{
		URL:       link,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"), // Proper ISO-ish format
	}

	resp, err := client.Get(link)
	if err != nil {
		res.ErrorMsg = err.Error() // error ko string mein convert kiya
		c <- res
		return
	}
	defer resp.Body.Close()

	res.StatusCode = resp.StatusCode
	c <- res
}