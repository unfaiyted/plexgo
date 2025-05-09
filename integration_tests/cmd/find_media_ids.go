package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/unfaiyted/plexgo"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get Plex client configuration
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
	sectionIDStr := os.Getenv("PLEX_SECTION_ID")
	if sectionIDStr == "" {
		log.Fatal("PLEX_SECTION_ID environment variable is not set")
	}
	
	sectionID, err := strconv.Atoi(sectionIDStr)
	if err != nil {
		log.Fatalf("Invalid section ID: %v", err)
	}

	// Create PlexGo client
	var serverProtocol plexgo.ServerProtocol
	if protocol == "https" {
		serverProtocol = plexgo.ServerProtocolHTTPS
	} else {
		serverProtocol = plexgo.ServerProtocolHTTP
	}

	client := plexgo.New(
		plexgo.WithSecurity(token),
		plexgo.WithProtocol(serverProtocol),
		plexgo.WithIP(ip),
		plexgo.WithPort(port),
	)

	// Get all media from library (similar to GetLibraryItems but more reliable)
	fmt.Printf("Getting media items from library section %d...\n", sectionID)
	
	response, err := client.Library.GetAllMediaLibrary(context.Background(), sectionID)
	if err != nil {
		log.Fatalf("Error getting library items: %v", err)
	}

	// Check if we have metadata from the response
	if response.Object == nil || response.Object.MediaContainer.Metadata == nil {
		log.Fatalf("No items found in section %d or error retrieving items", sectionID)
	}

	metadata := response.Object.MediaContainer.Metadata

	// Display item IDs and titles
	fmt.Printf("Found %d items in section %d:\n\n", len(metadata), sectionID)
	fmt.Println("RatingKey | Title")
	fmt.Println("----------|------------------")
	
	// Show first 20 items
	count := 0
	var itemsToUse []string

	for i, item := range metadata {
		if count >= 20 {
			fmt.Println("\n...and more. Showing only first 20 items.")
			break
		}
		
		if item.RatingKey != nil && item.Title != nil {
			fmt.Printf("%s | %s\n", *item.RatingKey, *item.Title)
			itemsToUse = append(itemsToUse, *item.RatingKey)
			count++
		} else {
			fmt.Printf("Item %d has nil RatingKey or Title\n", i)
		}
	}
	
	// Show example for .env file
	if len(itemsToUse) > 0 {
		fmt.Println("\nExample for .env file:")
		fmt.Printf("PLEX_TEST_MEDIA_IDS=%s", itemsToUse[0])
		if len(itemsToUse) > 1 {
			fmt.Printf(",%s", itemsToUse[1])
		}
		if len(itemsToUse) > 2 {
			fmt.Printf(",%s", itemsToUse[2])
		}
		fmt.Println()
	}
}