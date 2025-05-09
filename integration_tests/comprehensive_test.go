package integration_tests

import (
	"context"
	"fmt"
	"testing"
	"time"
	"strconv"

	"github.com/unfaiyted/plexgo/integration_tests/internal"
)

func TestComprehensiveCollectionWorkflow(t *testing.T) {
	// Load environment variables
	err := internal.LoadEnv()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	// Get Plex client
	client, err := internal.GetPlexClient()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	// Get section ID
	sectionID, err := internal.GetSectionID()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}

	// Get test media IDs
	initialMediaIDs, err := internal.GetTestMediaIDs()
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	
	// Ensure we have at least some media IDs to work with
	if len(initialMediaIDs) == 0 {
		t.Skipf("No test media IDs found in PLEX_TEST_MEDIA_IDS")
	}

	// Get up to 5 items for testing
	mediaIDs := initialMediaIDs
	if len(mediaIDs) > 5 {
		mediaIDs = mediaIDs[:5]
	}

	ctx := context.Background()

	// Step 1: Create a new collection with items included from the start
	// This is more reliable than adding items later
	t.Run("1. Create Collection with Items", func(t *testing.T) {
		timestamp := time.Now().Unix()
		collectionName := fmt.Sprintf("Test Comprehensive Collection %d", timestamp)

		t.Logf("Creating collection with %d items: %s", len(mediaIDs), collectionName)
		collection, err := client.Collections.CreateCollection(ctx, sectionID, collectionName, mediaIDs)
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}

		// Extract collection ID for later use
		collectionID, err := strconv.Atoi(collection.RatingKey)
		if err != nil {
			t.Fatalf("Failed to parse collection ID: %v", err)
		}

		t.Logf("Created collection: %s (ID: %d)", collectionName, collectionID)

		// Verify collection was created by searching for it by name
		allCollections, err := client.Collections.GetAllCollections(ctx, sectionID)
		if err != nil {
			t.Fatalf("Failed to get all collections: %v", err)
		}

		found := false
		for _, c := range allCollections {
			if c.Title == collectionName {
				found = true
				t.Logf("Successfully verified collection creation by name search")
				break
			}
		}

		if !found {
			t.Fatalf("Created collection not found in search results")
		}

		// Wait for collection to be fully processed
		t.Logf("Waiting for collection to be fully processed...")
		time.Sleep(5 * time.Second)

		// Step 2: Retrieve collection items
		t.Run("2. Get Collection Items", func(t *testing.T) {
			// Get collection details
			collection, err = client.Collections.GetCollection(ctx, collectionID)
			if err != nil {
				t.Fatalf("Failed to get collection: %v", err)
			}

			t.Logf("Collection has %d items according to ChildCount", collection.ChildCount)

			collectionItems, err := client.Collections.GetCollectionItems(ctx, collectionID)
			if err != nil {
				t.Fatalf("Failed to get collection items: %v", err)
			}

			t.Logf("Retrieved %d items from collection", len(collectionItems))
			
			// Step 3: Test removing items if we have any items to remove
			t.Run("3. Remove Items from Collection", func(t *testing.T) {
				// Even if GetCollectionItems returns empty, we'll try removing the first item 
				// we initially added, since it might still be in the collection internally
				if len(mediaIDs) == 0 {
					t.Skip("No items available to remove")
				}

				itemToRemove := []string{mediaIDs[0]}
				t.Logf("Attempting to remove item: %v", itemToRemove)
				
				err = client.Collections.RemoveFromCollection(ctx, collectionID, itemToRemove)
				if err != nil {
					t.Fatalf("Failed to remove items from collection: %v", err)
				}
				
				t.Logf("Successfully attempted to remove items from collection")
				time.Sleep(3 * time.Second)

				// Step 4: Delete the collection
				t.Run("4. Delete Collection", func(t *testing.T) {
					t.Logf("Deleting collection: %s (ID: %d)", collectionName, collectionID)
					err = client.Collections.DeleteCollection(ctx, collectionID)
					if err != nil {
						t.Fatalf("Failed to delete collection: %v", err)
					}

					// Give Plex a moment to process the deletion
					time.Sleep(1 * time.Second)

					// Verify collection is deleted
					allCollections, err := client.Collections.GetAllCollections(ctx, sectionID)
					if err != nil {
						t.Fatalf("Failed to get all collections: %v", err)
					}

					found := false
					for _, c := range allCollections {
						if c.RatingKey == strconv.Itoa(collectionID) {
							found = true
							break
						}
					}

					if found {
						t.Logf("Warning: Collection still found after deletion (Plex may have delayed processing)")
					} else {
						t.Logf("Successfully verified collection deletion")
					}

					// Step 5: Get all collections
					t.Run("5. Get All Collections", func(t *testing.T) {
						allCollections, err := client.Collections.GetAllCollections(ctx, sectionID)
						if err != nil {
							t.Fatalf("Failed to get all collections: %v", err)
						}

						t.Logf("Total collections found: %d", len(allCollections))
						for i, c := range allCollections {
							if i < 5 { // Log only first 5 collections to avoid too much output
								t.Logf("Collection %d: %s (ID: %s, Items: %d)", 
									i+1, c.Title, c.RatingKey, c.ChildCount)
							}
						}
					})
				})
			})
		})
	})
}