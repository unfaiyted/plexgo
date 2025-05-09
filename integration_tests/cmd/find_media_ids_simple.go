package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Simple struct for JSON parsing
type PlexResponse struct {
	MediaContainer struct {
		Metadata []struct {
			RatingKey string `json:"ratingKey"`
			Title     string `json:"title"`
			Type      string `json:"type"`
		} `json:"Metadata"`
	} `json:"MediaContainer"`
}

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

	// Build URL
	url := fmt.Sprintf("%s://%s:%s/library/sections/%s/all", protocol, ip, port, sectionID)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
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

	// Parse JSON
	var plexResp PlexResponse
	err = json.Unmarshal(body, &plexResp)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	// Display item IDs and titles
	fmt.Printf("Found %d items in section %s:\n\n", len(plexResp.MediaContainer.Metadata), sectionID)
	fmt.Println("RatingKey | Title | Type")
	fmt.Println("----------|------------------|--------")

	// Show first 20 items
	count := 0
	var mediaIDs []string

	for _, item := range plexResp.MediaContainer.Metadata {
		if count >= 20 {
			fmt.Println("\n...and more. Showing only first 20 items.")
			break
		}

		fmt.Printf("%s | %s | %s\n", item.RatingKey, item.Title, item.Type)
		mediaIDs = append(mediaIDs, item.RatingKey)
		count++
	}

	// Show example for .env file
	if len(mediaIDs) > 0 {
		fmt.Println("\nExample for .env file:")
		ids := []string{mediaIDs[0]}
		if len(mediaIDs) > 1 {
			ids = append(ids, mediaIDs[1])
		}
		if len(mediaIDs) > 2 {
			ids = append(ids, mediaIDs[2])
		}
		fmt.Printf("PLEX_TEST_MEDIA_IDS=%s\n", strings.Join(ids, ","))
	}
}