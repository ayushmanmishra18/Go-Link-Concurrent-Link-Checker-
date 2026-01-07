package main

import( 
	"fmt"
	"os"
) 

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usasge: go run main.go <url>")
		os.Exit(1)
	}
	url := os.Args[1]

	fmt.Printf("checking links: %s\n", url)

}