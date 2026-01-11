package main

import (
	"fmt"
	"net/http"
	"os"
	"sync" // Synchronization primitives ke liye
)

func main() {
	links := os.Args[1:]

	if len(links) == 0 {
		fmt.Println("Usage: go run main.go <url1> <url2> ...")
		os.Exit(1)
	}

	// Step 4: WaitGroup initialize karna
	// Ye ek counter ki tarah hai jo track rakhega kitne workers kaam kar rahe hain
	var wg sync.WaitGroup

	for _, link := range links {
		// Har link ke liye counter ko +1 karo
		wg.Add(1)

		// 'go' keyword lagate hi ye function background mein chalne lagega
		// Humne yahan pointer (&wg) bheja hai taaki function original counter ko update kar sake
		go checkLink(link, &wg)
	}

	// Main function ko bolo: "Ruko! Jab tak counter 0 na ho jaye, kahin mat jana"
	wg.Wait()
	fmt.Println("\nAll links checked. Shanti!")
}

// checkLink mein ab hum WaitGroup ka pointer bhi le rahe hain
func checkLink(link string, wg *sync.WaitGroup) {
	// Function khatam hote hi counter ko -1 kar do
	defer wg.Done()

	resp, err := http.Get(link)
	if err != nil {
		fmt.Printf("❌ Error checking %s: %v\n", link, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ [OK] %s\n", link)
	} else {
		fmt.Printf("⚠️ [BROKEN] %s (Status: %d)\n", link, resp.StatusCode)
	}
}