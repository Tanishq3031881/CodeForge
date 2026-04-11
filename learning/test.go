package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	// Step 1: check input
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <url>")
		return
	}

	url := os.Args[1]

	// Step 2: make request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	// Step 3: read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading body:", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: received status code %d\n", resp.StatusCode)
	}

	// Step 4: print results
	fmt.Println("Status:", resp.Status)
	fmt.Println("Content Length:", len(body))
}