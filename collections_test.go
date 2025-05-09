package plexgo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// MockHTTPClient is a mock HTTP client for testing
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestGetAllCollections(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for the expected endpoint
		if r.URL.Path != "/library/sections/1/collections" {
			t.Errorf("Expected request to '/library/sections/1/collections', got: %s", r.URL.Path)
		}

		// Check if the request method is GET
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got: %s", r.Method)
		}

		// Return a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Create a mock response body
		response := CollectionResponse{
			MediaContainer: CollectionMediaContainer{
				Size:      2,
				TotalSize: 2,
				Metadata: []Collection{
					{
						RatingKey:    "1",
						Title:        "Test Collection 1",
						ChildCount:   3,
						SectionID:    1,
						SectionTitle: "Movies",
						Type:         "collection",
					},
					{
						RatingKey:    "2",
						Title:        "Test Collection 2",
						ChildCount:   2,
						SectionID:    1,
						SectionTitle: "Movies",
						Type:         "collection",
					},
				},
				AllowSync:  true,
				Identifier: "com.plexapp.plugins.library",
			},
		}

		// Encode the response
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested
	collections, err := client.Collections.GetAllCollections(context.Background(), 1)
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Check the number of collections
	if len(collections) != 2 {
		t.Errorf("Expected 2 collections, got: %d", len(collections))
	}
	
	// Check the collection titles
	if collections[0].Title != "Test Collection 1" {
		t.Errorf("Expected collection title 'Test Collection 1', got: %s", collections[0].Title)
	}
	
	if collections[1].Title != "Test Collection 2" {
		t.Errorf("Expected collection title 'Test Collection 2', got: %s", collections[1].Title)
	}
}

func TestCreateCollection(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the first request is to create the collection
		if r.URL.Path == "/library/collections" && r.Method == "POST" {
			// Return a mock response with Location header
			w.Header().Set("Location", "/library/collections/3")
			w.WriteHeader(http.StatusCreated)
			return
		}
		
		// Check if the second request is to get the collection details
		if r.URL.Path == "/library/collections/3" && r.Method == "GET" {
			// Return a mock response with collection details
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			
			// Create a mock response body
			response := CollectionResponse{
				MediaContainer: CollectionMediaContainer{
					Size:      1,
					TotalSize: 1,
					Metadata: []Collection{
						{
							RatingKey:    "3",
							Title:        "New Collection",
							ChildCount:   2,
							SectionID:    1,
							SectionTitle: "Movies",
							Type:         "collection",
						},
					},
					AllowSync:  true,
					Identifier: "com.plexapp.plugins.library",
				},
			}
			
			// Encode the response
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()
	
	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested
	collection, err := client.Collections.CreateCollection(
		context.Background(),
		1,
		"New Collection",
		[]string{"1234", "5678"},
	)
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Check the collection details
	if collection.RatingKey != "3" {
		t.Errorf("Expected collection RatingKey '3', got: %s", collection.RatingKey)
	}
	
	if collection.Title != "New Collection" {
		t.Errorf("Expected collection Title 'New Collection', got: %s", collection.Title)
	}
	
	if collection.SectionID != 1 {
		t.Errorf("Expected collection SectionID 1, got: %d", collection.SectionID)
	}
}

func TestGetCollection(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for the expected endpoint
		if r.URL.Path != "/library/collections/5" {
			t.Errorf("Expected request to '/library/collections/5', got: %s", r.URL.Path)
		}

		// Check if the request method is GET
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got: %s", r.Method)
		}

		// Return a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Create a mock response body
		response := CollectionResponse{
			MediaContainer: CollectionMediaContainer{
				Size:      1,
				TotalSize: 1,
				Metadata: []Collection{
					{
						RatingKey:       "5",
						Key:             "/library/collections/5/children",
						GUID:            "collection://5",
						Title:           "Action Movies",
						Summary:         "Collection of action movies",
						Smart:           false,
						AddedAt:         1620000000,
						UpdatedAt:       1620100000,
						ChildCount:      10,
						CollectionMode:  "default",
						CollectionSort:  "release",
						SectionID:       1,
						SectionTitle:    "Movies",
						SectionUUID:     "section-uuid",
						Type:            "collection",
					},
				},
				AllowSync:  true,
				Identifier: "com.plexapp.plugins.library",
			},
		}

		// Encode the response
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested
	collection, err := client.Collections.GetCollection(context.Background(), 5)
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Check the collection details
	if collection.RatingKey != "5" {
		t.Errorf("Expected collection RatingKey '5', got: %s", collection.RatingKey)
	}
	
	if collection.Title != "Action Movies" {
		t.Errorf("Expected collection Title 'Action Movies', got: %s", collection.Title)
	}
	
	if collection.ChildCount != 10 {
		t.Errorf("Expected collection ChildCount 10, got: %d", collection.ChildCount)
	}
	
	if collection.CollectionMode != "default" {
		t.Errorf("Expected collection CollectionMode 'default', got: %s", collection.CollectionMode)
	}
	
	if collection.CollectionSort != "release" {
		t.Errorf("Expected collection CollectionSort 'release', got: %s", collection.CollectionSort)
	}
}

func TestGetCollectionItems(t *testing.T) {
	// This test needs two requests: first to get the collection, then to get the items
	requestCount := 0
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		
		// First request: Get the collection to check if it's smart or not
		if requestCount == 1 {
			if r.URL.Path != "/library/collections/5" || r.Method != "GET" {
				t.Errorf("Expected first request to GET /library/collections/5, got: %s %s", r.Method, r.URL.Path)
			}
			
			// Return a mock collection (non-smart)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CollectionResponse{
				MediaContainer: CollectionMediaContainer{
					Size: 1,
					Metadata: []Collection{
						{
							RatingKey:    "5",
							Key:          "/library/collections/5/children",
							Title:        "Regular Collection",
							Smart:        false,
							SectionID:    1,
							Type:         "collection",
						},
					},
				},
			})
			return
		}
		
		// Second request: Get the collection items
		if requestCount == 2 {
			if r.URL.Path != "/library/collections/5/children" {
				t.Errorf("Expected second request to '/library/collections/5/children', got: %s", r.URL.Path)
			}

			if r.Method != "GET" {
				t.Errorf("Expected GET request, got: %s", r.Method)
			}

			// Return a mock response with collection items
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// Create a mock response body with collection items
			response := CollectionResponse{
				MediaContainer: CollectionMediaContainer{
					Size:      3,
					TotalSize: 3,
					Metadata: []Collection{
						{
							RatingKey:    "101",
							Title:        "Movie 1",
							Type:         "movie",
						},
						{
							RatingKey:    "102",
							Title:        "Movie 2",
							Type:         "movie",
						},
						{
							RatingKey:    "103",
							Title:        "Movie 3",
							Type:         "movie",
						},
					},
					AllowSync:  true,
					Identifier: "com.plexapp.plugins.library",
				},
			}

			// Encode the response
			json.NewEncoder(w).Encode(response)
			return
		}
		
		t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested
	items, err := client.Collections.GetCollectionItems(context.Background(), 5)
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Verify we made the expected requests
	if requestCount != 2 {
		t.Errorf("Expected 2 requests, got: %d", requestCount)
	}
	
	// Check the items returned
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got: %d", len(items))
	}
	
	// Check the specific items
	expectedItems := []string{"101", "102", "103"}
	for i, expected := range expectedItems {
		if items[i] != expected {
			t.Errorf("Expected item %d to be '%s', got: '%s'", i, expected, items[i])
		}
	}
}

func TestCreateSmartCollection(t *testing.T) {
	// For testing smart collections, we need to account for the smart filter test
	requestCount := 0
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		
		// First request: Test the smart filter
		if requestCount == 1 {
			// Accept any library sections path with all query parameter
			if strings.Contains(r.URL.Path, "/library/sections/") && 
			   strings.Contains(r.URL.Path, "/all") && 
			   r.Method == "GET" {
				// Return a mock response with some items to indicate the filter is valid
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(CollectionResponse{
					MediaContainer: CollectionMediaContainer{
						Size: 2,
						Metadata: []Collection{
							{
								RatingKey: "101",
								Title:    "Action Movie 1",
								Type:     "movie",
							},
							{
								RatingKey: "102",
								Title:    "Action Movie 2",
								Type:     "movie",
							},
						},
					},
				})
				return
			}
		}
		
		// Second request: Create the smart collection
		if requestCount == 2 {
			if r.URL.Path == "/library/collections" && r.Method == "POST" {
				// Check if this is a smart collection
				if !strings.Contains(r.URL.RawQuery, "smart=1") {
					t.Errorf("Expected smart=1 parameter for smart collection, got: %s", r.URL.RawQuery)
				}
				
				// Check for the URI parameter with the filter
				if !strings.Contains(r.URL.RawQuery, "uri=") {
					t.Errorf("Expected uri parameter for smart collection, got: %s", r.URL.RawQuery)
				}
				
				// Return a mock response with Location header
				w.Header().Set("Location", "/library/collections/7")
				w.WriteHeader(http.StatusCreated)
				return
			}
		}
		
		// Third request: Get the collection details
		if requestCount == 3 {
			if r.URL.Path == "/library/collections/7" && r.Method == "GET" {
				// Return a mock response with collection details
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				
				// Create a mock response body
				response := CollectionResponse{
					MediaContainer: CollectionMediaContainer{
						Size:      1,
						TotalSize: 1,
						Metadata: []Collection{
							{
								RatingKey:    "7",
								Title:        "Smart Action Movies",
								Smart:        true,
								ChildCount:   5,
								SectionID:    1,
								SectionTitle: "Movies",
								Type:         "collection",
							},
						},
						AllowSync:  true,
						Identifier: "com.plexapp.plugins.library",
					},
				}
				
				// Encode the response
				json.NewEncoder(w).Encode(response)
				return
			}
		}
		
		t.Errorf("Unexpected request #%d: %s %s", requestCount, r.Method, r.URL.Path)
	}))
	defer server.Close()
	
	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested
	collection, err := client.Collections.CreateSmartCollection(
		context.Background(),
		1,
		"Smart Action Movies",
		1, // type 1 for movies
		"?genre=action", // filter for action movies
	)
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Ensure we made the expected requests
	if requestCount != 3 {
		t.Errorf("Expected 3 requests, got: %d", requestCount)
	}
	
	// Check the collection details
	if collection.RatingKey != "7" {
		t.Errorf("Expected collection RatingKey '7', got: %s", collection.RatingKey)
	}
	
	if collection.Title != "Smart Action Movies" {
		t.Errorf("Expected collection Title 'Smart Action Movies', got: %s", collection.Title)
	}
	
	if !collection.IsSmartCollection() {
		t.Errorf("Expected collection to be smart, got: %v", collection.Smart)
	}
}

func TestDeleteCollection(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for the expected endpoint
		if r.URL.Path != "/library/collections/8" {
			t.Errorf("Expected request to '/library/collections/8', got: %s", r.URL.Path)
		}

		// Check if the request method is DELETE
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got: %s", r.Method)
		}

		// Return a successful response
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested
	err := client.Collections.DeleteCollection(context.Background(), 8)
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestUpdateCollectionMode(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for the expected endpoint
		if r.URL.Path != "/library/collections/9/prefs" {
			t.Errorf("Expected request to '/library/collections/9/prefs', got: %s", r.URL.Path)
		}

		// Check if the request method is PUT
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got: %s", r.Method)
		}
		
		// Check if the mode parameter is correct
		if !strings.Contains(r.URL.RawQuery, "collectionMode=") {
			t.Errorf("Expected collectionMode parameter, got: %s", r.URL.RawQuery)
		}

		// Return a successful response
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested
	err := client.Collections.UpdateCollectionMode(context.Background(), 9, CollectionModeShowItems)
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestUpdateCollectionSort(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for the expected endpoint
		if r.URL.Path != "/library/collections/10/prefs" {
			t.Errorf("Expected request to '/library/collections/10/prefs', got: %s", r.URL.Path)
		}

		// Check if the request method is PUT
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got: %s", r.Method)
		}
		
		// Check if the sort parameter is correct
		if !strings.Contains(r.URL.RawQuery, "collectionSort=") {
			t.Errorf("Expected collectionSort parameter, got: %s", r.URL.RawQuery)
		}

		// Return a successful response
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested
	err := client.Collections.UpdateCollectionSort(context.Background(), 10, CollectionSortAlpha)
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestGetCollectionVisibility(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for the expected endpoint
		if r.URL.Path != "/hubs/sections/1/manage" {
			t.Errorf("Expected request to '/hubs/sections/1/manage', got: %s", r.URL.Path)
		}

		// Check if the request method is GET
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got: %s", r.Method)
		}
		
		// Check if the metadataItemId parameter is correct
		if !strings.Contains(r.URL.RawQuery, "metadataItemId=11") {
			t.Errorf("Expected metadataItemId=11 parameter, got: %s", r.URL.RawQuery)
		}

		// Return a mock response for collection visibility
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// Format for collection visibility response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"MediaContainer": map[string]interface{}{
				"size": 1,
				"Directory": []map[string]interface{}{
					{
						"promotedToRecommended": "1",
						"promotedToOwnHome": "1",
						"promotedToSharedHome": "0",
					},
				},
			},
		})
	}))
	defer server.Close()

	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested
	visibility, err := client.Collections.GetCollectionVisibility(context.Background(), 1, 11)
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Check the visibility settings
	if !visibility.Library {
		t.Errorf("Expected Library visibility to be true, got: %v", visibility.Library)
	}
	
	if !visibility.Home {
		t.Errorf("Expected Home visibility to be true, got: %v", visibility.Home)
	}
	
	if visibility.Shared {
		t.Errorf("Expected Shared visibility to be false, got: %v", visibility.Shared)
	}
}

func TestUpdateCollectionVisibility(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for the expected endpoint
		if r.URL.Path != "/hubs/sections/1/manage" {
			t.Errorf("Expected request to '/hubs/sections/1/manage', got: %s", r.URL.Path)
		}

		// Check if the request method is POST
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got: %s", r.Method)
		}
		
		// Check if the parameters are correct
		query := r.URL.RawQuery
		if !strings.Contains(query, "metadataItemId=12") {
			t.Errorf("Expected metadataItemId=12 parameter, got: %s", query)
		}
		
		if !strings.Contains(query, "promotedToRecommended=1") {
			t.Errorf("Expected promotedToRecommended=1 parameter, got: %s", query)
		}
		
		if !strings.Contains(query, "promotedToOwnHome=1") {
			t.Errorf("Expected promotedToOwnHome=1 parameter, got: %s", query)
		}
		
		if !strings.Contains(query, "promotedToSharedHome=1") {
			t.Errorf("Expected promotedToSharedHome=1 parameter, got: %s", query)
		}

		// Return a successful response
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Create visibility settings
	visibility := &CollectionVisibility{
		Library: true,
		Home:    true,
		Shared:  true,
	}
	
	// Call the method being tested
	err := client.Collections.UpdateCollectionVisibility(context.Background(), 1, 12, visibility)
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestAddToCollection(t *testing.T) {
	// This is a multistep operation that requires mocking multiple requests
	requestCount := 0
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		
		// First request: Get collection
		if requestCount == 1 {
			if r.URL.Path != "/library/collections/13" || r.Method != "GET" {
				t.Errorf("Expected first request to GET /library/collections/13, got: %s %s", r.Method, r.URL.Path)
			}
			
			// Return a mock collection
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CollectionResponse{
				MediaContainer: CollectionMediaContainer{
					Size: 1,
					Metadata: []Collection{
						{
							RatingKey:    "13",
							Title:        "Test Collection",
							SectionID:    1,
							Type:         "collection",
						},
					},
				},
			})
			return
		}
		
		// Second request: Add items to collection (PUT to items endpoint)
		if requestCount == 2 {
			if r.URL.Path != "/library/collections/13/items" || r.Method != "PUT" {
				t.Errorf("Expected second request to PUT /library/collections/13/items, got: %s %s", r.Method, r.URL.Path)
			}
			
			// Check that uri parameter is present
			if !strings.Contains(r.URL.RawQuery, "uri=") {
				t.Errorf("Expected uri parameter in URL, got: %s", r.URL.RawQuery)
			}
			
			// Return success
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	
	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested
	err := client.Collections.AddToCollection(context.Background(), 13, []string{"103"})
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Check that all expected requests were made
	if requestCount != 2 {
		t.Errorf("Expected 2 requests, got: %d", requestCount)
	}
}

func TestRemoveFromCollection(t *testing.T) {
	// This is a multistep operation that requires mocking multiple requests
	requestCount := 0
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		
		// First request: Get collection
		if requestCount == 1 {
			if r.URL.Path != "/library/collections/14" || r.Method != "GET" {
				t.Errorf("Expected first request to GET /library/collections/14, got: %s %s", r.Method, r.URL.Path)
			}
			
			// Return a mock collection
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CollectionResponse{
				MediaContainer: CollectionMediaContainer{
					Size: 1,
					Metadata: []Collection{
						{
							RatingKey:    "14",
							Title:        "Test Collection",
							SectionID:    1,
							Type:         "collection",
						},
					},
				},
			})
			return
		}
		
		// Second request: Remove item from collection (DELETE to items/itemID endpoint)
		if requestCount == 2 {
			expectedPath := "/library/collections/14/items/102"
			if r.URL.Path != expectedPath || r.Method != "DELETE" {
				t.Errorf("Expected second request to DELETE %s, got: %s %s", expectedPath, r.Method, r.URL.Path)
			}
			
			// Return success
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	
	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested
	err := client.Collections.RemoveFromCollection(context.Background(), 14, []string{"102"})
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Check that all expected requests were made
	if requestCount != 2 {
		t.Errorf("Expected 2 requests, got: %d", requestCount)
	}
}

func TestUpdateSmartCollection(t *testing.T) {
	// This test needs to handle the URI parsing step and the filter test step
	requestCount := 0
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		
		// First request: Get collection to verify it's a smart collection
		if requestCount == 1 {
			if r.URL.Path != "/library/collections/15" || r.Method != "GET" {
				t.Errorf("Expected first request to GET /library/collections/15, got: %s %s", r.Method, r.URL.Path)
			}
			
			// Return a mock smart collection
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CollectionResponse{
				MediaContainer: CollectionMediaContainer{
					Size: 1,
					Metadata: []Collection{
						{
							RatingKey:    "15",
							Title:        "Smart Collection",
							Smart:        true,
							SectionID:    1,
							Type:         "collection",
						},
					},
				},
			})
			return
		}
		
		// Second request: Test the smart filter by checking for results
		if requestCount == 2 {
			// Accept any library sections path with all query parameter
			if strings.Contains(r.URL.Path, "/library/sections/") && 
			   strings.Contains(r.URL.Path, "/all") && 
			   r.Method == "GET" {
				// Return a mock response with some items to indicate the filter is valid
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(CollectionResponse{
					MediaContainer: CollectionMediaContainer{
						Size: 1,
						Metadata: []Collection{
							{
								RatingKey: "101",
								Title:    "New Action Movie",
								Type:     "movie",
							},
						},
					},
				})
				return
			}
		}
		
		// Third request: Update the smart collection filter
		if requestCount == 3 {
			// Check if the request is for the expected endpoint
			if r.URL.Path != "/library/collections/15/items" {
				t.Errorf("Expected request to '/library/collections/15/items', got: %s", r.URL.Path)
			}

			// Check if the request method is PUT
			if r.Method != "PUT" {
				t.Errorf("Expected PUT request, got: %s", r.Method)
			}
			
			// Check if the URI parameter is correct
			if !strings.Contains(r.URL.RawQuery, "uri=") {
				t.Errorf("Expected uri parameter, got: %s", r.URL.RawQuery)
			}

			// Return a successful response
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		t.Errorf("Unexpected request #%d: %s %s", requestCount, r.Method, r.URL.Path)
	}))
	defer server.Close()

	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested - use a URI that includes both server and query part
	filterURI := fmt.Sprintf("%s/library/sections/1/all?genre=action&year>=2020", server.URL)
	err := client.Collections.UpdateSmartCollection(context.Background(), 15, filterURI)
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Check that all expected requests were made
	if requestCount != 3 {
		t.Errorf("Expected 3 requests, got: %d", requestCount)
	}
}

func TestMoveCollectionItem(t *testing.T) {
	// This test needs to get the collection first to check if it's a smart collection,
	// then perform the move operation
	requestCount := 0
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		
		// First request: Get collection to verify it's not a smart collection
		if requestCount == 1 {
			if r.URL.Path != "/library/collections/16" || r.Method != "GET" {
				t.Errorf("Expected first request to GET /library/collections/16, got: %s %s", r.Method, r.URL.Path)
			}
			
			// Return a mock regular collection
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(CollectionResponse{
				MediaContainer: CollectionMediaContainer{
					Size: 1,
					Metadata: []Collection{
						{
							RatingKey:    "16",
							Title:        "Regular Collection",
							Smart:        false,
							SectionID:    1,
							Type:         "collection",
						},
					},
				},
			})
			return
		}
		
		// Second request: Move an item in the collection
		if requestCount == 2 {
			expectedPath := "/library/collections/16/items/102/move"
			if r.URL.Path != expectedPath || r.Method != "PUT" {
				t.Errorf("Expected second request to PUT %s, got: %s %s", expectedPath, r.Method, r.URL.Path)
			}
			
			// Check for the after parameter
			if !strings.Contains(r.URL.RawQuery, "after=101") {
				t.Errorf("Expected after=101 parameter, got: %s", r.URL.RawQuery)
			}
			
			// Return success
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()
	
	// Create a client with the mock server URL
	client := New(WithServerURL(server.URL))
	
	// Call the method being tested - move item 102 after item 101
	err := client.Collections.MoveCollectionItem(context.Background(), 16, "102", "101")
	
	// Check for errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Check that both expected requests were made
	if requestCount != 2 {
		t.Errorf("Expected 2 requests, got: %d", requestCount)
	}
}