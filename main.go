package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	// os.Args[1:] ka matlab hai 1st index se lekar end tak saare arguments lapet lo
	// Isse hume ek string slice mil jayega URLs ka
	links := os.Args[1:]

	if len(links) == 0 {
		fmt.Println("Usage: go run main.go <url1> <url2> ...")
		os.Exit(1)
	}

	// Step 3: Loop chalao aur har link ko check karo
	for _, link := range links {
		checkLink(link)
	}
}

// checkLink function humne alag se banaya hai taaki code neat rahe
// Ye function abhi "serial" kaam kar raha hai (ek ke baad ek)
func checkLink(link string) {
	resp, err := http.Get(link)
	if err != nil {
		fmt.Printf("❌ Error checking %s: %v\n", link, err)
		return // Function se bahar aa jao, main loop chalta rahega
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ [OK] %s\n", link)
	} else {
		fmt.Printf("⚠️ [BROKEN] %s (Status: %d)\n", link, resp.StatusCode)
	}
}
