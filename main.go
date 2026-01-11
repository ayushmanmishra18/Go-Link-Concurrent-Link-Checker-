package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	links := os.Args[1:]

	if len(links) == 0 {
		fmt.Println("Usage: go run main.go <url1> <url2> ...")
		os.Exit(1)
	}

	// Step 5: Channel create karo
	// Ye string type ka channel hai, matlab is pipe mein sirf strings travel karenge
	resultsChannel := make(chan string)

	for _, link := range links {
		// Har worker ko channel de do taaki wo result bhej sake
		go checkLink(link, resultsChannel)
	}

	// Jitne links bheje, utne hi results receive karne padenge
	// Ye loop tab tak wait karega jab tak channel se data na mile (Blocking nature!)
	for i := 0; i < len(links); i++ {
		fmt.Println(<-resultsChannel)
	}

	fmt.Println("\nKaam khatam, paisa hazam! Saare results collect ho gaye.")
}

// checkLink ab WaitGroup ki jagah channel use karega data wapas bhejne ke liye
func checkLink(link string, c chan string) {
	resp, err := http.Get(link)
	
	if err != nil {
		// Result ko channel mein phenk do (Arrow pointing towards channel)
		c <- fmt.Sprintf("❌ %s is DOWN! (Error: %v)", link, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c <- fmt.Sprintf("✅ %s is UP!", link)
	} else {
		c <- fmt.Sprintf("⚠️ %s returned status %d", link, resp.StatusCode)
	}
}