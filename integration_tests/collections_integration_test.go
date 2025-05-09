package integration_tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/unfaiyted/plexgo"
	"github.com/unfaiyted/plexgo/integration_tests/internal"
)

// These tests are designed to be run against a real Plex server.
// Set the following environment variables to run these tests:
// PLEX_TOKEN: Your Plex authentication token
// PLEX_SERVER_IP: The IP address of your Plex server
// PLEX_SERVER_PORT: The port of your Plex server
// PLEX_SECTION_ID: The ID of a movie or TV show library section to test with

func getClient() *plexgo.PlexAPI {
	// Load environment variables from .env file
	_ = internal.LoadEnv() // Ignore error, we'll check for env vars next
	
	client, err := internal.GetPlexClient()
	if err != nil {
		return nil
	}
	
	return client
}

func getSectionID() int {
	// Load environment variables from .env file
	_ = internal.LoadEnv() // Ignore error, we'll check for env vars next
	
	sectionID, err := internal.GetSectionID()
	if err != nil {
		return 1 // Default to section ID 1
	}
	return sectionID
}

func TestIntegration_GetAllCollections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getClient()
	if client == nil {
		t.Skip("Skipping test; PLEX_TOKEN, PLEX_SERVER_IP, or PLEX_SERVER_PORT not set")
	}

	collections, err := client.Collections.GetAllCollections(context.Background(), getSectionID())
	if err != nil {
		t.Fatalf("Error getting collections: %v", err)
	}

	t.Logf("Found %d collections", len(collections))
	for i, collection := range collections {
		if i < 5 { // Log only first 5 collections to avoid too much output
			t.Logf("Collection: %s (ID: %s, Items: %d)", collection.Title, collection.RatingKey, collection.ChildCount)
		}
	}
}

func TestIntegration_CreateAndDeleteCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getClient()
	if client == nil {
		t.Skip("Skipping test; PLEX_TOKEN, PLEX_SERVER_IP, or PLEX_SERVER_PORT not set")
	}

	ctx := context.Background()
	sectionID := getSectionID()

	// Create a test collection
	collection, err := client.Collections.CreateCollection(
		ctx,
		sectionID,
		"Test Integration Collection",
		[]string{}, // Empty collection
	)

	if err != nil {
		t.Fatalf("Error creating collection: %v", err)
	}

	t.Logf("Created collection: %s (ID: %s)", collection.Title, collection.RatingKey)

	// Get the collection ID
	collectionID := 0
	fmt.Sscanf(collection.RatingKey, "%d", &collectionID)

	// Clean up - delete the collection
	err = client.Collections.DeleteCollection(ctx, collectionID)
	if err != nil {
		t.Fatalf("Error deleting collection: %v", err)
	}

	t.Logf("Successfully deleted collection")
}

func TestIntegration_CollectionMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := getClient()
	if client == nil {
		t.Skip("Skipping test; PLEX_TOKEN, PLEX_SERVER_IP, or PLEX_SERVER_PORT not set")
	}

	ctx := context.Background()
	sectionID := getSectionID()

	// First, create a test collection
	collection, err := client.Collections.CreateCollection(
		ctx,
		sectionID,
		"Test Mode Collection",
		[]string{}, // Empty collection
	)

	if err != nil {
		t.Fatalf("Error creating collection: %v", err)
	}

	t.Logf("Created collection: %s (ID: %s)", collection.Title, collection.RatingKey)

	// Get the collection ID
	collectionID := 0
	fmt.Sscanf(collection.RatingKey, "%d", &collectionID)

	// Test updating the collection mode
	err = client.Collections.UpdateCollectionMode(ctx, collectionID, plexgo.CollectionModeShowItems)
	if err != nil {
		t.Fatalf("Error updating collection mode: %v", err)
	}

	t.Logf("Successfully updated collection mode")

	// Clean up - delete the collection
	err = client.Collections.DeleteCollection(ctx, collectionID)
	if err != nil {
		t.Fatalf("Error deleting collection: %v", err)
	}
}

// Add more integration tests as needed...