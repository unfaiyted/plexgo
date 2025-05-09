package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get Plex configuration
	token := os.Getenv("PLEX_TOKEN")
	if token == "" {
		log.Fatal("PLEX_TOKEN environment variable is not set")
	}

	protocol := os.Getenv("PLEX_SERVER_PROTOCOL")
	if protocol == "" {
		protocol = "https" // Default to https
	}

	ip := os.Getenv("PLEX_SERVER_IP")
	if ip == "" {
		log.Fatal("PLEX_SERVER_IP environment variable is not set")
	}

	port := os.Getenv("PLEX_SERVER_PORT")
	if port == "" {
		log.Fatal("PLEX_SERVER_PORT environment variable is not set")
	}

	// Get section ID
	sectionID := os.Getenv("PLEX_SECTION_ID")
	if sectionID == "" {
		log.Fatal("PLEX_SECTION_ID environment variable is not set")
	}

	// Get test media IDs
	mediaIDs := os.Getenv("PLEX_TEST_MEDIA_IDS")
	if mediaIDs == "" {
		log.Fatal("PLEX_TEST_MEDIA_IDS environment variable is not set")
	}

	// Generate a unique collection name
	timestamp := time.Now().Unix()
	collectionName := fmt.Sprintf("Debug Collection %d", timestamp)

	// Build URL
	baseURL := fmt.Sprintf("%s://%s:%s", protocol, ip, port)
	apiURL := fmt.Sprintf("%s/library/collections", baseURL)

	// Create query parameters
	params := url.Values{}
	params.Add("type", "1") // Type 1 for movies
	params.Add("title", collectionName)
	params.Add("smart", "0") // Regular collection, not smart
	params.Add("sectionId", sectionID)
	
	// Add item IDs as a comma-separated list in the uri parameter
	uri := fmt.Sprintf("server://plex-server-machine-id/com.plexapp.plugins.library/library/metadata/%s", mediaIDs)
	params.Add("uri", uri)

	// Append query parameters to URL
	apiURL = fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Add headers
	req.Header.Add("X-Plex-Token", token)
	req.Header.Add("Accept", "application/json")

	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	// Log response details
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Headers:\n")
	for k, v := range resp.Header {
		fmt.Printf("  %s: %s\n", k, v)
	}
	fmt.Printf("Response Body:\n%s\n", string(body))

	// Check for Location header
	location := resp.Header.Get("Location")
	if location == "" {
		fmt.Println("\nNOTE: No Location header found in response!")
		fmt.Println("This will cause the SDK integration tests to fail.")
	} else {
		fmt.Printf("\nLocation Header Found: %s\n", location)
	}
}