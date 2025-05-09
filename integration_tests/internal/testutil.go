package internal

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/LukeHagar/plexgo"
	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from .env file
func LoadEnv() error {
	// Try different potential locations for the .env file
	locations := []string{
		".env",
		"../.env",
		"../../.env",
	}
	
	for _, location := range locations {
		err := godotenv.Load(location)
		if err == nil {
			return nil // Successfully loaded
		}
	}
	
	return fmt.Errorf("error loading .env file: could not find .env file in known locations")
}

// GetPlexClient creates a new Plex client using environment variables
func GetPlexClient() (*plexgo.PlexAPI, error) {
	token := os.Getenv("PLEX_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("PLEX_TOKEN environment variable is not set")
	}

	protocol := os.Getenv("PLEX_SERVER_PROTOCOL")
	if protocol == "" {
		protocol = "https" // Default to https
	}

	ip := os.Getenv("PLEX_SERVER_IP")
	if ip == "" {
		return nil, fmt.Errorf("PLEX_SERVER_IP environment variable is not set")
	}

	port := os.Getenv("PLEX_SERVER_PORT")
	if port == "" {
		return nil, fmt.Errorf("PLEX_SERVER_PORT environment variable is not set")
	}

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

	return client, nil
}

// GetSectionID returns the library section ID from environment variables
func GetSectionID() (int, error) {
	sectionID := os.Getenv("PLEX_SECTION_ID")
	if sectionID == "" {
		return 0, fmt.Errorf("PLEX_SECTION_ID environment variable is not set")
	}

	id, err := strconv.Atoi(sectionID)
	if err != nil {
		return 0, fmt.Errorf("invalid PLEX_SECTION_ID: %s - %w", sectionID, err)
	}

	return id, nil
}

// GetTestMediaIDs returns a slice of media IDs from environment variables
func GetTestMediaIDs() ([]string, error) {
	mediaIDs := os.Getenv("PLEX_TEST_MEDIA_IDS")
	if mediaIDs == "" {
		return nil, fmt.Errorf("PLEX_TEST_MEDIA_IDS environment variable is not set")
	}

	return strings.Split(mediaIDs, ","), nil
}