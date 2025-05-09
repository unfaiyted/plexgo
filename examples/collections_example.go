package main

import (
	"context"
	"fmt"
	"log"

	"github.com/LukeHagar/plexgo"
)

func main() {
	// Create a new Plex API client with your X-Plex-Token
	client := plexgo.New(
		plexgo.WithSecurity("<YOUR-PLEX-TOKEN>"),
		plexgo.WithIP("<YOUR-PLEX-SERVER-IP>"),
		plexgo.WithPort("<YOUR-PLEX-SERVER-PORT>"),
	)

	// Create a background context
	ctx := context.Background()

	// Example 1: Get all collections from a library section
	fmt.Println("Getting all collections from library section 1...")
	collections, err := client.Collections.GetAllCollections(ctx, 1)
	if err != nil {
		log.Fatalf("Error getting collections: %v", err)
	}

	// Print collection info
	fmt.Printf("Found %d collections\n", len(collections))
	for _, collection := range collections {
		fmt.Printf("- %s (ID: %s, Items: %d)\n", collection.Title, collection.RatingKey, collection.ChildCount)
	}

	// Example 2: Create a new collection
	fmt.Println("\nCreating a new collection...")
	newCollection, err := client.Collections.CreateCollection(
		ctx,
		1,                 // Library section ID
		"My New Collection", // Collection title
		[]string{"1234", "5678"}, // Item IDs to add to the collection
	)
	if err != nil {
		log.Fatalf("Error creating collection: %v", err)
	}
	fmt.Printf("Created collection: %s (ID: %s)\n", newCollection.Title, newCollection.RatingKey)

	// Example 3: Get collection details
	fmt.Println("\nGetting collection details...")
	collectionID := 123 // Replace with an actual collection ID
	collection, err := client.Collections.GetCollection(ctx, collectionID)
	if err != nil {
		log.Fatalf("Error getting collection: %v", err)
	}
	fmt.Printf("Collection: %s (ID: %s, Items: %d)\n", collection.Title, collection.RatingKey, collection.ChildCount)

	// Example 4: Add items to a collection
	fmt.Println("\nAdding items to a collection...")
	err = client.Collections.AddToCollection(
		ctx,
		collectionID,
		[]string{"9876"}, // New item IDs to add
	)
	if err != nil {
		log.Fatalf("Error adding items to collection: %v", err)
	}
	fmt.Println("Items added to collection successfully")

	// Example 5: Update collection mode
	fmt.Println("\nUpdating collection mode...")
	err = client.Collections.UpdateCollectionMode(
		ctx,
		collectionID,
		plexgo.CollectionModeShowItems, // Use the predefined constant
	)
	if err != nil {
		log.Fatalf("Error updating collection mode: %v", err)
	}
	fmt.Println("Collection mode updated successfully")

	// Example 6: Create a smart collection
	fmt.Println("\nCreating a smart collection...")
	smartFilter := "?type=1&genre=action" // Example filter: action movies
	smartCollection, err := client.Collections.CreateSmartCollection(
		ctx,
		1,                   // Library section ID
		"Action Movies",     // Collection title
		1,                   // Type 1 for movies
		smartFilter,         // Filter args
	)
	if err != nil {
		log.Fatalf("Error creating smart collection: %v", err)
	}
	fmt.Printf("Created smart collection: %s (ID: %s)\n", smartCollection.Title, smartCollection.RatingKey)

	// Example 7: Delete a collection
	fmt.Println("\nDeleting a collection...")
	deleteCollectionID := 456 // Replace with an actual collection ID to delete
	err = client.Collections.DeleteCollection(ctx, deleteCollectionID)
	if err != nil {
		log.Fatalf("Error deleting collection: %v", err)
	}
	fmt.Println("Collection deleted successfully")
}