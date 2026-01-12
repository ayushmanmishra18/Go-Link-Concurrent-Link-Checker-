package main

import (
	"context"
	"encoding/json"
	"flag" // Professional CLI flags handle karne ke liye
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type result struct {
	URL        string `json:"url" dynamodbav:"URL"`
	StatusCode int    `json:"status_code" dynamodbav:"StatusCode"`
	ErrorMsg   string `json:"error,omitempty" dynamodbav:"ErrorMsg"`
	Timestamp  string `json:"timestamp" dynamodbav:"Timestamp"`
}

func main() {
	// Step 17: Command Line Flags define karna
	// Syntax: flag.Type("flag_name", default_value, "description")
	workerCount := flag.Int("workers", 5, "Number of concurrent workers")
	checkInterval := flag.Duration("interval", 10*time.Second, "Interval between checks for each URL")
	timeout := flag.Duration("timeout", 5*time.Second, "HTTP timeout duration")
	
	// Flags ko parse karna zaroori hai
	flag.Parse()

	// Baaki bache huye arguments (URLs) lene ke liye
	links := flag.Args()

	if len(links) == 0 {
		fmt.Println("Usage: go run main.go [flags] <url1> <url2> ...")
		flag.PrintDefaults() // Saare available flags print kar dega
		os.Exit(1)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		fmt.Printf("Unable to load SDK config, %v\n", err)
		os.Exit(1)
	}

	s3Client := s3.NewFromConfig(cfg)
	dbClient := dynamodb.NewFromConfig(cfg)

	jobs := make(chan string)
	results := make(chan result)
	
	// Humne 'timeout' flag yahan use kiya hai
	client := &http.Client{Timeout: *timeout}

	var workerWg sync.WaitGroup
	// Humne 'workerCount' flag yahan use kiya hai
	for w := 1; w <= *workerCount; w++ {
		workerWg.Add(1)
		go worker(w, jobs, results, client, &workerWg)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for _, link := range links {
			jobs <- link
		}
	}()

	uploadTicker := time.NewTicker(30 * time.Second)
	f, _ := os.OpenFile("log_results.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	fmt.Printf("🚀 Monitoring started with %d workers. Timeout: %v\n", *workerCount, *timeout)

	for {
		select {
		case res := <-results:
			jsonData, _ := json.Marshal(res)
			f.WriteString(string(jsonData) + "\n")

			if res.ErrorMsg != "" || res.StatusCode != http.StatusOK {
				go updateDynamoDBStatus(context.TODO(), dbClient, res)
			}

			go func(l string) {
				// Humne 'checkInterval' flag yahan use kiya hai
				time.Sleep(*checkInterval)
				select {
				case jobs <- l:
				default:
				}
			}(res.URL)

		case <-uploadTicker.C:
			fmt.Println("\n--- ☁️  REAL S3 UPLOAD STARTING ---")
			uploadToS3(context.TODO(), s3Client, "log_results.jsonl")

		case <-sigs:
			fmt.Println("\n\n🛑 Shutting down safely...")
			uploadTicker.Stop()
			close(jobs)
			workerWg.Wait()
			uploadToS3(context.TODO(), s3Client, "log_results.jsonl")
			f.Close()
			fmt.Println("✅ Done!")
			return
		}
	}
}

// ... Baki functions (updateDynamoDBStatus, uploadToS3, worker, checkLink) same rahenge ...

func updateDynamoDBStatus(ctx context.Context, api *dynamodb.Client, res result) {
	item, err := attributevalue.MarshalMap(res)
	if err != nil {
		fmt.Printf("Error marshaling struct: %v\n", err)
		return
	}
	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String("LinkStatus"),
	}
	_, err = api.PutItem(ctx, input)
	if err != nil {
		fmt.Printf("❌ DynamoDB Error for %s: %v\n", res.URL, err)
		return
	}
	fmt.Printf("✅ DynamoDB: Status updated for %s\n", res.URL)
}

func uploadToS3(ctx context.Context, api *s3.Client, filename string) {
	fmt.Printf("📦 AWS: Ready to upload %s to S3...\n", filename)
}

func worker(id int, jobs <-chan string, results chan<- result, client *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	for link := range jobs {
		res := checkLink(link, client)
		results <- res
	}
}

func checkLink(link string, client *http.Client) result {
	res := result{URL: link, Timestamp: time.Now().Format("2006-01-02 15:04:05")}
	resp, err := client.Get(link)
	if err != nil {
		res.ErrorMsg = err.Error()
		return res
	}
	defer resp.Body.Close()
	res.StatusCode = resp.StatusCode
	return res
}