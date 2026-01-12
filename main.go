package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// --- 1. DATA STRUCTURES ---

// result struct mein JSON tags hain taaki data universal bhasha mein rahe.
// Ismein humne DynamoDB tags bhi rakhe hain (for future real AWS integration).
type result struct {
	URL        string `json:"url" dynamodbav:"URL"`
	StatusCode int    `json:"status_code" dynamodbav:"StatusCode"`
	ErrorMsg   string `json:"error,omitempty" dynamodbav:"ErrorMsg"`
	Timestamp  string `json:"timestamp" dynamodbav:"Timestamp"`
}

// --- 2. INTERFACES (DECOUPLING) ---

// ResultStore ek contract hai. Isse humara code flexible banta hai.
type ResultStore interface {
	Record(res result) error
}

// FileStore: Local JSONL file mein data save karne ke liye.
type FileStore struct {
	file *os.File
}

func (fs *FileStore) Record(res result) error {
	jsonData, _ := json.Marshal(res)
	_, err := fs.file.WriteString(string(jsonData) + "\n")
	return err
}

// CloudStore: AWS (S3/DynamoDB) simulation ke liye.
type CloudStore struct {
	// Real AWS clients yahan aayenge
}

func (cs *CloudStore) Record(res result) error {
	// Agar link down hai, toh hum 'Alert' simulate karte hain
	if res.ErrorMsg != "" || res.StatusCode != http.StatusOK {
		fmt.Printf("⚠️  [CLOUD ALERT] %s is problematic! Syncing to AWS...\n", res.URL)
	}
	return nil
}

// --- 3. MAIN LOGIC ---

func main() {
	// CLI Flags: Program ko dynamic banane ke liye
	workerCount := flag.Int("workers", 3, "Number of concurrent workers")
	checkInterval := flag.Duration("interval", 5*time.Second, "Time between checks")
	timeout := flag.Duration("timeout", 5*time.Second, "HTTP timeout")
	flag.Parse()

	links := flag.Args()
	if len(links) == 0 {
		fmt.Println("Usage: go run main.go -workers=5 -interval=10s <url1> <url2> ...")
		os.Exit(1)
	}

	// Channel Setup: Communication ke liye pipes
	jobs := make(chan string)
	results := make(chan result)
	client := &http.Client{Timeout: *timeout}

	// Storage Setup: Interface use karke multiple jagah save karo
	f, _ := os.OpenFile("log_results.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	stores := []ResultStore{
		&FileStore{file: f},
		&CloudStore{},
	}

	// Worker Pool: Background workers ko kaam par lagao
	var workerWg sync.WaitGroup
	for w := 1; w <= *workerCount; w++ {
		workerWg.Add(1)
		go worker(w, jobs, results, client, &workerWg)
	}

	// Graceful Shutdown: Ctrl+C handle karne ke liye
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Initial jobs queue mein daalo
	go func() {
		for _, link := range links {
			jobs <- link
		}
	}()

	fmt.Printf("🚀 Go-Link Monitoring Started! Workers: %d, Timeout: %v\n", *workerCount, *timeout)
	fmt.Println("Press Ctrl+C to stop safely.")

	// Event Loop
	for {
		select {
		case res := <-results:
			// Har store mein result record karo (File + Cloud)
			for _, store := range stores {
				store.Record(res)
			}

			// Print to console for visibility
			statusLabel := "✅ UP"
			if res.ErrorMsg != "" || res.StatusCode != http.StatusOK {
				statusLabel = "❌ DOWN"
			}
			fmt.Printf("[%s] %s | %s (Code: %d)\n", res.Timestamp, statusLabel, res.URL, res.StatusCode)

			// Re-queue the job for continuous monitoring
			go func(l string) {
				time.Sleep(*checkInterval)
				// Check if channel is still open before sending
				select {
				case jobs <- l:
				default:
					// Program is shutting down
				}
			}(res.URL)

		case <-sigs:
			// Shutdown logic
			fmt.Println("\n🛑 Signal received. Cleaning up and exiting...")
			close(jobs)      // Workers ko bolo kaam khatam karo
			workerWg.Wait()  // Unka wait karo
			fmt.Println("✅ All workers finished. Logs saved. Bye!")
			return
		}
	}
}

// --- 4. HELPER FUNCTIONS ---

func worker(id int, jobs <-chan string, results chan<- result, client *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	for link := range jobs {
		results <- checkLink(link, client)
	}
}

func checkLink(link string, client *http.Client) result {
	res := result{
		URL:       link,
		Timestamp: time.Now().Format("15:04:05"),
	}

	resp, err := client.Get(link)
	if err != nil {
		res.ErrorMsg = err.Error()
		return res
	}
	defer resp.Body.Close()

	res.StatusCode = resp.StatusCode
	return res
}