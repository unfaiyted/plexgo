# PlexGo Integration Tests

This directory contains integration tests for the PlexGo SDK that connect to a real Plex Media Server to verify functionality.

## Setup

1. Create a copy of the `.env.example` file in the project root directory and rename it to `.env`:

```bash
cp ../.env.example ../.env
```

2. Edit the `.env` file and fill in your Plex server details:

```
# Plex Server Configuration
PLEX_SERVER_PROTOCOL=https  # or http
PLEX_SERVER_IP=10.10.10.47  # Your Plex server IP
PLEX_SERVER_PORT=32400      # Usually 32400
PLEX_TOKEN=your_plex_token_here  # Your Plex authentication token

# Test Configuration
PLEX_SECTION_ID=1  # The ID of a library section to use for testing
PLEX_TEST_MEDIA_IDS=123,456,789  # Comma-separated list of media IDs to use for testing
```

3. Install dependencies:

```bash
go mod tidy
```

## Running the Tests

Run all integration tests:

```bash
go test -v ./...
```

Run a specific test:

```bash
go test -v -run TestCollections_Integration/GetCollection
```

## Finding Media IDs for Testing

To find media IDs for testing, you can:

1. Use the Plex Web UI and inspect network requests in your browser's developer tools
2. Use a tool like `curl` to query your Plex server:

```bash
curl -H "X-Plex-Token: YOUR_TOKEN" https://YOUR_SERVER:32400/library/sections/SECTION_ID/all
```

3. Use the PlexGo SDK to list media items:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/unfaiyted/plexgo"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	
	token := os.Getenv("PLEX_TOKEN")
	ip := os.Getenv("PLEX_SERVER_IP")
	port := os.Getenv("PLEX_SERVER_PORT")
	
	client := plexgo.New(
		plexgo.WithSecurity(token),
		plexgo.WithIP(ip),
		plexgo.WithPort(port),
	)
	
	// Get media items from a section
	// This is simplified - check the actual SDK for the correct method to use
	items, err := client.Library.GetLibraryItems(context.Background(), 1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	for _, item := range items {
		fmt.Printf("ID: %s, Title: %s\n", item.RatingKey, item.Title)
	}
}
```

## Test Coverage

The integration tests cover the following functionality:

- Creating collections
- Retrieving collections
- Updating collection settings
- Managing collection contents
- Creating and managing smart collections
- Managing collection visibility

## Notes

- These tests create real collections on your Plex server, but they clean up after themselves
- Tests are designed to be idempotent and to not interfere with each other
- Some operations may not be immediately reflected in the Plex API due to caching