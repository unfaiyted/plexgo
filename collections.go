// Package plexgo provides a Plex Media Server API client.
package plexgo

import (
	"bytes"
	"context"
	"fmt"
	"github.com/unfaiyted/plexgo/internal/hooks"
	"github.com/unfaiyted/plexgo/internal/utils"
	"github.com/unfaiyted/plexgo/models/operations"
	"github.com/unfaiyted/plexgo/models/sdkerrors"
	// "log"

	"github.com/unfaiyted/plexgo/retry"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Collection represents a Plex collection
type Collection struct {
	RatingKey       string      `json:"ratingKey"`
	Key             string      `json:"key"`
	GUID            string      `json:"guid"`
	Title           string      `json:"title"`
	TitleSort       string      `json:"titleSort,omitempty"`
	Summary         string      `json:"summary,omitempty"`
	Smart           interface{} `json:"smart,omitempty"` // Can be bool or string
	AddedAt         int64       `json:"addedAt"`
	UpdatedAt       int64       `json:"updatedAt,omitempty"`
	ContentRating   string      `json:"contentRating,omitempty"`
	Thumb           string      `json:"thumb,omitempty"`
	Art             string      `json:"art,omitempty"`
	ChildCount      int         `json:"childCount,omitempty"`
	CollectionMode  string      `json:"collectionMode,omitempty"`
	CollectionSort  string      `json:"collectionSort,omitempty"`
	SectionID       int         `json:"librarySectionID"`
	SectionTitle    string      `json:"librarySectionTitle"`
	SectionUUID     string      `json:"librarySectionUUID,omitempty"`
	Type            string      `json:"type"`
	SubType         string      `json:"subtype,omitempty"`
	CollectionItems []string    `json:"-"` // Slice of rating keys for items in the collection
}

// IsSmartCollection returns true if the collection is a smart collection
func (c *Collection) IsSmartCollection() bool {
	switch v := c.Smart.(type) {
	case bool:
		return v
	case string:
		return v == "1" || v == "true"
	case float64:
		return v > 0
	case int:
		return v > 0
	default:
		return false
	}
}

// CollectionVisibility represents collection visibility settings
type CollectionVisibility struct {
	Library bool
	Home    bool
	Shared  bool
}

// SmartFilterConfig represents smart filter configuration
type SmartFilterConfig struct {
	Type   int    // 1=movie, 2=show, etc.
	Filter string // filter string
	URI    string // full smart filter URI
}

// CollectionMode constants
const (
	CollectionModeDefault   = "default"
	CollectionModeHide      = "hide"
	CollectionModeHideItems = "hideItems"
	CollectionModeShowItems = "showItems"
)

// CollectionModeKeys maps between Plex numeric mode values and string constants
var CollectionModeKeys = map[int]string{
	-1: CollectionModeDefault,
	0:  CollectionModeHide,
	1:  CollectionModeHideItems,
	2:  CollectionModeShowItems,
}

// CollectionSort constants
const (
	CollectionSortRelease = "release"
	CollectionSortAlpha   = "alpha"
	CollectionSortCustom  = "custom"
)

// CollectionSortKeys maps between Plex numeric sort values and string constants
var CollectionSortKeys = map[int]string{
	0: CollectionSortRelease,
	1: CollectionSortAlpha,
	2: CollectionSortCustom,
}

// Collections provides operations for working with collections
type Collections struct {
	sdkConfiguration sdkConfiguration
}

func newCollections(sdkConfig sdkConfiguration) *Collections {
	return &Collections{
		sdkConfiguration: sdkConfig,
	}
}

// CollectionMediaContainer represents the response from the collections API
type CollectionMediaContainer struct {
	Size       int          `json:"size"`
	TotalSize  int          `json:"totalSize"`
	Metadata   []Collection `json:"Metadata,omitempty"`
	AllowSync  bool         `json:"allowSync"`
	Identifier string       `json:"identifier"`
	Content    string       `json:"content,omitempty"` // Used for smart collection filter URI
}

// CollectionResponse represents the response from the collections API
type CollectionResponse struct {
	MediaContainer CollectionMediaContainer `json:"MediaContainer"`
}

// GetAllCollections gets all collections, optionally filtered by label
func (s *Collections) GetAllCollections(ctx context.Context, sectionID int, opts ...operations.Option) ([]Collection, error) {
	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/library/sections/%d/collections", sectionID))
	if err != nil {
		return nil, fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "getAllCollections",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", opURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return nil, err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return nil, err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return nil, err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return nil, err
		}
		return nil, sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return nil, err
		}
	}

	rawBody, err := utils.ConsumeRawBody(httpRes)
	if err != nil {
		return nil, err
	}

	var out CollectionResponse
	if err := utils.UnmarshalJsonFromResponseBody(bytes.NewBuffer(rawBody), &out, ""); err != nil {
		return nil, err
	}

	return out.MediaContainer.Metadata, nil
}

// GetCollection gets a collection by ID
func (s *Collections) GetCollection(ctx context.Context, collectionID int, opts ...operations.Option) (*Collection, error) {
	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/library/collections/%d", collectionID))
	if err != nil {
		return nil, fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "getCollection",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", opURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return nil, err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return nil, err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return nil, err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return nil, err
		}
		return nil, sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return nil, err
		}
	}

	rawBody, err := utils.ConsumeRawBody(httpRes)
	if err != nil {
		return nil, err
	}

	var out CollectionResponse
	if err := utils.UnmarshalJsonFromResponseBody(bytes.NewBuffer(rawBody), &out, ""); err != nil {
		return nil, err
	}

	if len(out.MediaContainer.Metadata) == 0 {
		return nil, fmt.Errorf("collection not found")
	}

	return &out.MediaContainer.Metadata[0], nil
}

// GetCollectionItems gets all items in a collection
func (s *Collections) GetCollectionItems(ctx context.Context, collectionID int, opts ...operations.Option) ([]string, error) {
	// First get the collection to check if it's a smart collection
	collection, err := s.GetCollection(ctx, collectionID, opts...)
	if err != nil {
		return nil, fmt.Errorf("error getting collection: %w", err)
	}

	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	var opURL string

	// Handle differently based on collection type
	if collection.IsSmartCollection() {
		// For smart collections, try to get the filter from the collection
		smartFilter, err := s.GetSmartFilter(ctx, collection, opts...)
		if err == nil {
			// Smart collections use the filter applied at the library level
			opURL, err = url.JoinPath(baseURL, fmt.Sprintf("/library/sections/%d/all%s", collection.SectionID, smartFilter))
			if err != nil {
				return nil, fmt.Errorf("error generating URL: %w", err)
			}
		} else {
			// If we can't get the smart filter (sometimes it's not accessible via API),
			// fall back to the regular method
			opURL, err = url.JoinPath(baseURL, fmt.Sprintf("/library/collections/%d/children", collectionID))
			if err != nil {
				return nil, fmt.Errorf("error generating URL: %w", err)
			}
		}
	} else {
		// For standard collections, just use the children endpoint
		opURL, err = url.JoinPath(baseURL, fmt.Sprintf("/library/collections/%d/children", collectionID))
		if err != nil {
			return nil, fmt.Errorf("error generating URL: %w", err)
		}
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "getCollectionItems",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", opURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return nil, err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return nil, err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return nil, err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return nil, err
		}
		return nil, sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return nil, err
		}
	}

	rawBody, err := utils.ConsumeRawBody(httpRes)
	if err != nil {
		return nil, err
	}

	var out CollectionResponse
	if err := utils.UnmarshalJsonFromResponseBody(bytes.NewBuffer(rawBody), &out, ""); err != nil {
		return nil, err
	}

	items := make([]string, 0, len(out.MediaContainer.Metadata))
	for _, item := range out.MediaContainer.Metadata {
		items = append(items, item.RatingKey)
	}

	return items, nil
}

// CreateCollection creates a new collection with the given items
func (s *Collections) CreateCollection(ctx context.Context, sectionID int, title string, itemIDs []string, opts ...operations.Option) (*Collection, error) {
	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	opURL, err := url.JoinPath(baseURL, "/library/collections")
	if err != nil {
		return nil, fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "createCollection",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	queryParams := url.Values{}
	queryParams.Add("type", "1") // Default to movie type
	queryParams.Add("title", title)
	queryParams.Add("smart", "0")
	queryParams.Add("sectionId", strconv.Itoa(sectionID))

	// Add item IDs as a comma-separated list
	if len(itemIDs) > 0 {
		itemList := ""
		for i, id := range itemIDs {
			if i > 0 {
				itemList += ","
			}
			itemList += id
		}
		// queryParams.Add("uri", fmt.Sprintf("server://%s/com.plexapp.plugins.library/library/metadata/%s", s.sdkConfiguration.GetServerMachineID(), itemList))
	} else {
		// Empty collection
		queryParams.Add("uri", fmt.Sprintf("%s/library/metadata", baseURL))
	}

	opURL = fmt.Sprintf("%s?%s", opURL, queryParams.Encode())

	req, err := http.NewRequestWithContext(ctx, "POST", opURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return nil, err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return nil, err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return nil, err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return nil, err
		}
		return nil, sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return nil, err
		}
	}

	var collectionID int

	// Try to get the collection ID from the Location header first
	location := httpRes.Header.Get("Location")
	if location != "" {
		// Location header should be something like /library/collections/12345
		// Extract the ID
		collectionIDStr := ""
		_, err = fmt.Sscanf(location, "/library/collections/%s", &collectionIDStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing collection ID from location header: %w", err)
		}

		collectionID, err = strconv.Atoi(collectionIDStr)
		if err != nil {
			return nil, fmt.Errorf("error converting collection ID to int: %w", err)
		}
	} else {
		// If no Location header, try to parse the response body
		rawBody, err := utils.ConsumeRawBody(httpRes)
		if err != nil {
			return nil, err
		}

		var resp CollectionResponse
		if err := utils.UnmarshalJsonFromResponseBody(bytes.NewBuffer(rawBody), &resp, ""); err != nil {
			return nil, err
		}

		if len(resp.MediaContainer.Metadata) == 0 {
			return nil, fmt.Errorf("no collection created or returned in response")
		}

		collectionIDStr := resp.MediaContainer.Metadata[0].RatingKey
		collectionID, err = strconv.Atoi(collectionIDStr)
		if err != nil {
			return nil, fmt.Errorf("error converting collection ID to int: %w", err)
		}
	}

	// Add a delay to allow Plex to process the changes
	// This improves reliability when immediately checking collection contents after creation/modification
	time.Sleep(2 * time.Second)

	// Get the created collection
	return s.GetCollection(ctx, collectionID, opts...)
}

// CreateSmartCollection creates a new smart collection with the given filter
func (s *Collections) CreateSmartCollection(ctx context.Context, sectionID int, title string, smartType int, filterArgs string, opts ...operations.Option) (*Collection, error) {
	options := processOptions(opts)

	// Ensure filterArgs has a leading ? if not already present
	if !strings.HasPrefix(filterArgs, "?") {
		filterArgs = "?" + filterArgs
	}

	// Test the smart filter first to ensure it returns results
	hasResults, err := s.TestSmartFilter(ctx, sectionID, filterArgs, opts...)
	if err != nil {
		return nil, fmt.Errorf("error testing smart filter: %w", err)
	}

	// If the filter returns no results, return an error
	// Note: We could add support for custom options in the future to ignore blank results
	if !hasResults {
		return nil, fmt.Errorf("smart filter returned no results: %s", filterArgs)
	}

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	opURL, err := url.JoinPath(baseURL, "/library/collections")
	if err != nil {
		return nil, fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "createSmartCollection",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	queryParams := url.Values{}
	queryParams.Add("type", strconv.Itoa(smartType))
	queryParams.Add("title", title)
	queryParams.Add("smart", "1")
	queryParams.Add("sectionId", strconv.Itoa(sectionID))

	// Build the smart filter URI using our helper method
	uri := s.BuildSmartFilterURI(sectionID, filterArgs, opts...)
	queryParams.Add("uri", uri)

	opURL = fmt.Sprintf("%s?%s", opURL, queryParams.Encode())

	req, err := http.NewRequestWithContext(ctx, "POST", opURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return nil, err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return nil, err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return nil, err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return nil, err
		}
		return nil, sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return nil, err
		}
	}

	var collectionID int

	// Try to get the collection ID from the Location header first
	location := httpRes.Header.Get("Location")
	if location != "" {
		// Location header should be something like /library/collections/12345
		// Extract the ID
		collectionIDStr := ""
		_, err = fmt.Sscanf(location, "/library/collections/%s", &collectionIDStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing collection ID from location header: %w", err)
		}

		collectionID, err = strconv.Atoi(collectionIDStr)
		if err != nil {
			return nil, fmt.Errorf("error converting collection ID to int: %w", err)
		}
	} else {
		// If no Location header, try to parse the response body
		rawBody, err := utils.ConsumeRawBody(httpRes)
		if err != nil {
			return nil, err
		}

		var resp CollectionResponse
		if err := utils.UnmarshalJsonFromResponseBody(bytes.NewBuffer(rawBody), &resp, ""); err != nil {
			return nil, err
		}

		if len(resp.MediaContainer.Metadata) == 0 {
			return nil, fmt.Errorf("no collection created or returned in response")
		}

		collectionIDStr := resp.MediaContainer.Metadata[0].RatingKey
		collectionID, err = strconv.Atoi(collectionIDStr)
		if err != nil {
			return nil, fmt.Errorf("error converting collection ID to int: %w", err)
		}
	}

	// Add a delay to allow Plex to process the changes
	// This improves reliability when immediately checking collection contents after creation/modification
	time.Sleep(2 * time.Second)

	// Get the created collection
	return s.GetCollection(ctx, collectionID, opts...)
}

// DeleteCollection deletes a collection
func (s *Collections) DeleteCollection(ctx context.Context, collectionID int, opts ...operations.Option) error {
	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/library/collections/%d", collectionID))
	if err != nil {
		return fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "deleteCollection",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", opURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return err
		}
		return sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return err
		}
	}

	// Add a delay to allow Plex to process the deletion
	// This improves reliability when immediately checking collection status after deletion
	time.Sleep(2 * time.Second)

	return nil
}

// AddToCollection adds items to a collection
func (s *Collections) AddToCollection(ctx context.Context, collectionID int, itemIDs []string, opts ...operations.Option) error {
	// First, get the collection to check if it's a smart collection
	collection, err := s.GetCollection(ctx, collectionID, opts...)
	if err != nil {
		return fmt.Errorf("error getting collection: %w", err)
	}

	// Check if it's a smart collection - cannot manually add items to smart collections
	if collection.IsSmartCollection() {
		return fmt.Errorf("cannot manually add items to a smart collection")
	}

	// If no items to add, return early
	if len(itemIDs) == 0 {
		return nil
	}

	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	// Join rating keys into comma-separated string
	ratingKeys := strings.Join(itemIDs, ",")

	// Build the metadata URI - first get the server machine ID
	serverIdentity, err := s.getServerIdentity(ctx)
	if err != nil {
		return fmt.Errorf("error getting server identity: %w", err)
	}

	if serverIdentity.Object == nil || serverIdentity.Object.MediaContainer == nil || serverIdentity.Object.MediaContainer.MachineIdentifier == nil {
		return fmt.Errorf("could not get server machine identifier")
	}

	machineID := *serverIdentity.Object.MediaContainer.MachineIdentifier

	// Build the complete URL
	opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/library/collections/%d/items", collectionID))
	if err != nil {
		return fmt.Errorf("error generating URL: %w", err)
	}

	// Create the URI using the server://{machineId}/com.plexapp.plugins.library format
	uri := fmt.Sprintf("server://%s/com.plexapp.plugins.library/library/metadata/%s", machineID, ratingKeys)

	// Add the URI as a query parameter
	queryParams := url.Values{}
	queryParams.Add("uri", uri)
	opURL = fmt.Sprintf("%s?%s", opURL, queryParams.Encode())

	// Set up the request context
	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "addToCollection",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "PUT", opURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return err
	}

	// Send the request
	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return err
		}
		return sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return err
		}
	}

	// Add a delay to allow Plex to process the changes
	// This improves reliability when immediately checking collection contents after modification
	time.Sleep(2 * time.Second)

	return nil
}

// RemoveFromCollection removes items from a collection
func (s *Collections) RemoveFromCollection(ctx context.Context, collectionID int, itemIDs []string, opts ...operations.Option) error {
	// First, get the collection to check if it's a smart collection
	collection, err := s.GetCollection(ctx, collectionID, opts...)
	if err != nil {
		return fmt.Errorf("error getting collection: %w", err)
	}

	// Check if it's a smart collection - cannot manually remove items from smart collections
	if collection.IsSmartCollection() {
		return fmt.Errorf("cannot manually remove items from a smart collection")
	}

	// If no items to remove, return early
	if len(itemIDs) == 0 {
		return nil
	}

	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "removeFromCollection",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	// Process each item to remove separately with a DELETE request
	for _, itemID := range itemIDs {
		// Build the endpoint URL for removing this specific item
		opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/library/collections/%d/items/%s", collectionID, itemID))
		if err != nil {
			return fmt.Errorf("error generating URL: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "DELETE", opURL, nil)
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
		}

		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

		if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
			return err
		}

		req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
		if err != nil {
			return err
		}

		httpRes, err := s.sdkConfiguration.Client.Do(req)
		if err != nil || httpRes == nil {
			if err != nil {
				err = fmt.Errorf("error sending request: %w", err)
			} else {
				err = fmt.Errorf("error sending request: no response")
			}

			_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
			return err
		} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
			httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
			if err != nil {
				return err
			}
			// Don't return an error for 404, it just means the item wasn't in the collection
			if httpRes.StatusCode != 404 {
				return sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
			}
		} else {
			httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
			if err != nil {
				return err
			}
		}
	}

	// Add a delay to allow Plex to process the changes
	// This improves reliability when immediately checking collection contents after modification
	time.Sleep(2 * time.Second)

	return nil
}

// MoveCollectionItem moves an item to a new position in the collection
func (s *Collections) MoveCollectionItem(ctx context.Context, collectionID int, itemID string, afterItemID string, opts ...operations.Option) error {
	// First, get the collection to check if it's a smart collection
	collection, err := s.GetCollection(ctx, collectionID, opts...)
	if err != nil {
		return fmt.Errorf("error getting collection: %w", err)
	}

	// Check if it's a smart collection - cannot manually move items in smart collections
	if collection.IsSmartCollection() {
		return fmt.Errorf("cannot manually move items in a smart collection")
	}

	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	// Build the base URL for the move operation
	opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/library/collections/%d/items/%s/move", collectionID, itemID))
	if err != nil {
		return fmt.Errorf("error generating URL: %w", err)
	}

	// Add the after parameter if specified
	if afterItemID != "" {
		queryParams := url.Values{}
		queryParams.Add("after", afterItemID)
		opURL = fmt.Sprintf("%s?%s", opURL, queryParams.Encode())
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "moveCollectionItem",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", opURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return err
		}
		return sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return err
		}
	}

	// Add a delay to allow Plex to process the changes
	time.Sleep(2 * time.Second)

	return nil
}

// UpdateCollectionMode updates the mode of a collection
func (s *Collections) UpdateCollectionMode(ctx context.Context, collectionID int, mode string, opts ...operations.Option) error {
	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/library/collections/%d/prefs", collectionID))
	if err != nil {
		return fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "updateCollectionMode",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	// Translate string mode to numeric mode
	modeValue := "-1" // default
	for k, v := range CollectionModeKeys {
		if v == mode {
			modeValue = strconv.Itoa(k)
			break
		}
	}

	queryParams := url.Values{}
	queryParams.Add("collectionMode", modeValue)
	opURL = fmt.Sprintf("%s?%s", opURL, queryParams.Encode())

	req, err := http.NewRequestWithContext(ctx, "PUT", opURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return err
		}
		return sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateCollectionSort updates the sort order of a collection
func (s *Collections) UpdateCollectionSort(ctx context.Context, collectionID int, sort string, opts ...operations.Option) error {
	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/library/collections/%d/prefs", collectionID))
	if err != nil {
		return fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "updateCollectionSort",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	// Translate string sort to numeric sort
	sortValue := "0" // default release
	for k, v := range CollectionSortKeys {
		if v == sort {
			sortValue = strconv.Itoa(k)
			break
		}
	}

	queryParams := url.Values{}
	queryParams.Add("collectionSort", sortValue)
	opURL = fmt.Sprintf("%s?%s", opURL, queryParams.Encode())

	req, err := http.NewRequestWithContext(ctx, "PUT", opURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return err
		}
		return sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetCollectionVisibility gets the visibility of a collection
func (s *Collections) GetCollectionVisibility(ctx context.Context, sectionID int, collectionID int, opts ...operations.Option) (*CollectionVisibility, error) {
	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/hubs/sections/%d/manage", sectionID))
	if err != nil {
		return nil, fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "getCollectionVisibility",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	queryParams := url.Values{}
	queryParams.Add("metadataItemId", strconv.Itoa(collectionID))
	opURL = fmt.Sprintf("%s?%s", opURL, queryParams.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", opURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return nil, err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return nil, err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return nil, err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return nil, err
		}
		return nil, sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return nil, err
		}
	}

	rawBody, err := utils.ConsumeRawBody(httpRes)
	if err != nil {
		return nil, err
	}

	// This response has a complex structure, so we'll extract the fields we need
	type Item struct {
		PromotedToRecommended string `json:"promotedToRecommended"`
		PromotedToOwnHome     string `json:"promotedToOwnHome"`
		PromotedToSharedHome  string `json:"promotedToSharedHome"`
	}

	type Container struct {
		Size     int    `json:"size"`
		Elements []Item `json:"Directory"`
	}

	type Response struct {
		MediaContainer Container `json:"MediaContainer"`
	}

	var resp Response
	if err := utils.UnmarshalJsonFromResponseBody(bytes.NewBuffer(rawBody), &resp, ""); err != nil {
		return nil, err
	}

	if len(resp.MediaContainer.Elements) == 0 {
		return nil, fmt.Errorf("no visibility information found for collection")
	}

	item := resp.MediaContainer.Elements[0]

	visibility := &CollectionVisibility{
		Library: item.PromotedToRecommended == "1",
		Home:    item.PromotedToOwnHome == "1",
		Shared:  item.PromotedToSharedHome == "1",
	}

	return visibility, nil
}

// UpdateCollectionVisibility updates the visibility of a collection
func (s *Collections) UpdateCollectionVisibility(ctx context.Context, sectionID int, collectionID int, visibility *CollectionVisibility, opts ...operations.Option) error {
	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/hubs/sections/%d/manage", sectionID))
	if err != nil {
		return fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "updateCollectionVisibility",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	queryParams := url.Values{}
	queryParams.Add("metadataItemId", strconv.Itoa(collectionID))
	queryParams.Add("promotedToRecommended", boolToString(visibility.Library))
	queryParams.Add("promotedToOwnHome", boolToString(visibility.Home))
	queryParams.Add("promotedToSharedHome", boolToString(visibility.Shared))
	opURL = fmt.Sprintf("%s?%s", opURL, queryParams.Encode())

	req, err := http.NewRequestWithContext(ctx, "POST", opURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return err
		}
		return sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateSmartCollection updates the smart filter for a collection
func (s *Collections) UpdateSmartCollection(ctx context.Context, collectionID int, filterURI string, opts ...operations.Option) error {
	options := processOptions(opts)

	// First, get the collection to verify it's a smart collection
	collection, err := s.GetCollection(ctx, collectionID, opts...)
	if err != nil {
		return fmt.Errorf("error getting collection: %w", err)
	}

	if !collection.IsSmartCollection() {
		return fmt.Errorf("cannot update smart filter for a non-smart collection")
	}

	// Parse filter URI to extract the query part and section ID
	parsedURI, err := url.Parse(filterURI)
	if err == nil && parsedURI.Path != "" {
		// Try to extract sectionID from the URL path
		pathParts := strings.Split(parsedURI.Path, "/")
		for i, part := range pathParts {
			if part == "sections" && i+1 < len(pathParts) {
				sectionIDStr := pathParts[i+1]
				sectionID, err := strconv.Atoi(sectionIDStr)
				if err == nil {
					// We successfully extracted the section ID, now test the filter
					query := "?" + parsedURI.RawQuery
					hasResults, err := s.TestSmartFilter(ctx, sectionID, query, opts...)
					if err != nil {
						return fmt.Errorf("error testing smart filter: %w", err)
					}

					// If the filter returns no results, return an error
					if !hasResults {
						return fmt.Errorf("smart filter returned no results: %s", query)
					}
					break
				}
			}
		}
	}

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/library/collections/%d/items", collectionID))
	if err != nil {
		return fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "updateSmartCollection",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	queryParams := url.Values{}
	queryParams.Add("uri", filterURI)
	opURL = fmt.Sprintf("%s?%s", opURL, queryParams.Encode())

	req, err := http.NewRequestWithContext(ctx, "PUT", opURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return err
		}
		return sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return err
		}
	}

	return nil
}

// Helper function to convert bool to "0" or "1"
func boolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// Helper function to process options
func processOptions(opts []operations.Option) *operations.Options {
	o := &operations.Options{}
	for _, opt := range opts {
		// Note: We ignore errors here as we're not checking for supported options
		_ = opt(o)
	}
	return o
}

// joinArgs returns a query string where only the value is URL encoded.
// Example return value: '?genre=action&type=1337'.
func (s *Collections) joinArgs(args map[string]string) string {
	if len(args) == 0 {
		return ""
	}

	// Create a list of keys for sorted output
	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}

	// Sort keys by lowercase comparison
	// This uses a simple bubble sort since we typically have few parameters
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if strings.ToLower(keys[i]) > strings.ToLower(keys[j]) {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	// Build query string with sorted keys
	var argList []string
	for _, key := range keys {
		value := args[key]
		// URL encode only the value, not the key
		encoded := url.QueryEscape(value)
		argList = append(argList, fmt.Sprintf("%s=%s", key, encoded))
	}

	return "?" + strings.Join(argList, "&")
}

// lowerFirst returns the string with the first character converted to lowercase
func (s *Collections) lowerFirst(str string) string {
	if len(str) == 0 {
		return ""
	}
	return strings.ToLower(str[0:1]) + str[1:]
}

// GetSmartFilter retrieves the smart filter URI for a smart collection
func (s *Collections) GetSmartFilter(ctx context.Context, collection *Collection, opts ...operations.Option) (string, error) {
	if !collection.IsSmartCollection() {
		return "", fmt.Errorf("collection is not a smart collection")
	}

	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	// Get the collection content which contains the smart filter URI
	opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/library/collections/%s", collection.RatingKey))
	if err != nil {
		return "", fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "getSmartFilter",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", opURL, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return "", err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return "", err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return "", err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return "", err
		}
		return "", sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return "", err
		}
	}

	rawBody, err := utils.ConsumeRawBody(httpRes)
	if err != nil {
		return "", err
	}

	var out CollectionResponse
	if err := utils.UnmarshalJsonFromResponseBody(bytes.NewBuffer(rawBody), &out, ""); err != nil {
		return "", err
	}

	// Extract filter from the content field
	if out.MediaContainer.Content == "" {
		return "", fmt.Errorf("smart filter not found in collection response")
	}

	// The smart filter is usually in the format of a URL, we want to extract just the query part
	parsedURL, err := url.Parse(out.MediaContainer.Content)
	if err != nil {
		return "", fmt.Errorf("error parsing smart filter URL: %w", err)
	}

	// Return just the query string part (with the '?' prefix)
	return "?" + parsedURL.RawQuery, nil
}

// BuildSmartFilterURI creates a full URI for a smart filter
func (s *Collections) BuildSmartFilterURI(sectionID int, filterQuery string, opts ...operations.Option) string {
	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	// Ensure filterQuery has a leading ? if not already present
	if !strings.HasPrefix(filterQuery, "?") {
		filterQuery = "?" + filterQuery
	}

	return fmt.Sprintf("%s/library/sections/%d/all%s", baseURL, sectionID, filterQuery)
}

// TestSmartFilter tests a smart filter to verify it returns results
func (s *Collections) TestSmartFilter(ctx context.Context, sectionID int, filterQuery string, opts ...operations.Option) (bool, error) {
	options := processOptions(opts)

	var baseURL string
	if options.ServerURL == nil {
		serverURL, params := s.sdkConfiguration.GetServerDetails()
		baseURL = utils.ReplaceParameters(serverURL, params)
	} else {
		baseURL = *options.ServerURL
	}

	// Ensure filterQuery has a leading ? if not already present
	if !strings.HasPrefix(filterQuery, "?") {
		filterQuery = "?" + filterQuery
	}

	opURL, err := url.JoinPath(baseURL, fmt.Sprintf("/library/sections/%d/all%s", sectionID, filterQuery))
	if err != nil {
		return false, fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "testSmartFilter",
		OAuth2Scopes:   []string{},
		SecuritySource: s.sdkConfiguration.Security,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", opURL, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	if err := utils.PopulateSecurity(ctx, req, s.sdkConfiguration.Security); err != nil {
		return false, err
	}

	req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
	if err != nil {
		return false, err
	}

	httpRes, err := s.sdkConfiguration.Client.Do(req)
	if err != nil || httpRes == nil {
		if err != nil {
			err = fmt.Errorf("error sending request: %w", err)
		} else {
			err = fmt.Errorf("error sending request: no response")
		}

		_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
		return false, err
	} else if utils.MatchStatusCodes([]string{"400", "401", "404", "4XX", "5XX"}, httpRes.StatusCode) {
		httpRes, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
		if err != nil {
			return false, err
		}
		return false, sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, "", httpRes)
	} else {
		httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
		if err != nil {
			return false, err
		}
	}

	rawBody, err := utils.ConsumeRawBody(httpRes)
	if err != nil {
		return false, err
	}

	var out CollectionResponse
	if err := utils.UnmarshalJsonFromResponseBody(bytes.NewBuffer(rawBody), &out, ""); err != nil {
		return false, err
	}

	// Return whether the filter returned any results
	return len(out.MediaContainer.Metadata) > 0, nil
}

func (s *Collections) getServerIdentity(ctx context.Context, opts ...operations.Option) (*operations.GetServerIdentityResponse, error) {
	o := operations.Options{}

	supportedOptions := []string{
		operations.SupportedOptionRetries,
		operations.SupportedOptionTimeout,
	}

	for _, opt := range opts {
		if err := opt(&o, supportedOptions...); err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	var baseURL string
	if o.ServerURL == nil {
		baseURL = utils.ReplaceParameters(s.sdkConfiguration.GetServerDetails())
	} else {
		baseURL = *o.ServerURL
	}
	opURL, err := url.JoinPath(baseURL, "/identity")
	if err != nil {
		return nil, fmt.Errorf("error generating URL: %w", err)
	}

	hookCtx := hooks.HookContext{
		BaseURL:        baseURL,
		Context:        ctx,
		OperationID:    "get-server-identity",
		OAuth2Scopes:   []string{},
		SecuritySource: nil,
	}

	timeout := o.Timeout
	if timeout == nil {
		timeout = s.sdkConfiguration.Timeout
	}

	if timeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", opURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.sdkConfiguration.UserAgent)

	for k, v := range o.SetHeaders {
		req.Header.Set(k, v)
	}

	globalRetryConfig := s.sdkConfiguration.RetryConfig
	retryConfig := o.Retries
	if retryConfig == nil {
		if globalRetryConfig != nil {
			retryConfig = globalRetryConfig
		}
	}

	var httpRes *http.Response
	if retryConfig != nil {
		httpRes, err = utils.Retry(ctx, utils.Retries{
			Config: retryConfig,
			StatusCodes: []string{
				"429",
				"500",
				"502",
				"503",
				"504",
			},
		}, func() (*http.Response, error) {
			if req.Body != nil {
				copyBody, err := req.GetBody()
				if err != nil {
					return nil, err
				}
				req.Body = copyBody
			}

			req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
			if err != nil {
				if retry.IsPermanentError(err) || retry.IsTemporaryError(err) {
					return nil, err
				}

				return nil, retry.Permanent(err)
			}

			httpRes, err := s.sdkConfiguration.Client.Do(req)
			if err != nil || httpRes == nil {
				if err != nil {
					err = fmt.Errorf("error sending request: %w", err)
				} else {
					err = fmt.Errorf("error sending request: no response")
				}

				_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
			}
			return httpRes, err
		})

		if err != nil {
			return nil, err
		} else {
			httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
			if err != nil {
				return nil, err
			}
		}
	} else {
		req, err = s.sdkConfiguration.Hooks.BeforeRequest(hooks.BeforeRequestContext{HookContext: hookCtx}, req)
		if err != nil {
			return nil, err
		}

		httpRes, err = s.sdkConfiguration.Client.Do(req)
		if err != nil || httpRes == nil {
			if err != nil {
				err = fmt.Errorf("error sending request: %w", err)
			} else {
				err = fmt.Errorf("error sending request: no response")
			}

			_, err = s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, nil, err)
			return nil, err
		} else if utils.MatchStatusCodes([]string{"408", "4XX", "5XX"}, httpRes.StatusCode) {
			_httpRes, err := s.sdkConfiguration.Hooks.AfterError(hooks.AfterErrorContext{HookContext: hookCtx}, httpRes, nil)
			if err != nil {
				return nil, err
			} else if _httpRes != nil {
				httpRes = _httpRes
			}
		} else {
			httpRes, err = s.sdkConfiguration.Hooks.AfterSuccess(hooks.AfterSuccessContext{HookContext: hookCtx}, httpRes)
			if err != nil {
				return nil, err
			}
		}
	}

	res := &operations.GetServerIdentityResponse{
		StatusCode:  httpRes.StatusCode,
		ContentType: httpRes.Header.Get("Content-Type"),
		RawResponse: httpRes,
	}

	switch {
	case httpRes.StatusCode == 200:
		switch {
		case utils.MatchContentType(httpRes.Header.Get("Content-Type"), `application/json`):
			rawBody, err := utils.ConsumeRawBody(httpRes)
			if err != nil {
				return nil, err
			}

			var out operations.GetServerIdentityResponseBody
			if err := utils.UnmarshalJsonFromResponseBody(bytes.NewBuffer(rawBody), &out, ""); err != nil {
				return nil, err
			}

			res.Object = &out
		default:
			rawBody, err := utils.ConsumeRawBody(httpRes)
			if err != nil {
				return nil, err
			}
			return nil, sdkerrors.NewSDKError(fmt.Sprintf("unknown content-type received: %s", httpRes.Header.Get("Content-Type")), httpRes.StatusCode, string(rawBody), httpRes)
		}
	case httpRes.StatusCode == 408:
		switch {
		case utils.MatchContentType(httpRes.Header.Get("Content-Type"), `application/json`):
			rawBody, err := utils.ConsumeRawBody(httpRes)
			if err != nil {
				return nil, err
			}

			var out sdkerrors.GetServerIdentityRequestTimeout
			if err := utils.UnmarshalJsonFromResponseBody(bytes.NewBuffer(rawBody), &out, ""); err != nil {
				return nil, err
			}

			out.RawResponse = httpRes
			return nil, &out
		default:
			rawBody, err := utils.ConsumeRawBody(httpRes)
			if err != nil {
				return nil, err
			}
			return nil, sdkerrors.NewSDKError(fmt.Sprintf("unknown content-type received: %s", httpRes.Header.Get("Content-Type")), httpRes.StatusCode, string(rawBody), httpRes)
		}
	case httpRes.StatusCode >= 400 && httpRes.StatusCode < 500:
		rawBody, err := utils.ConsumeRawBody(httpRes)
		if err != nil {
			return nil, err
		}
		return nil, sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, string(rawBody), httpRes)
	case httpRes.StatusCode >= 500 && httpRes.StatusCode < 600:
		rawBody, err := utils.ConsumeRawBody(httpRes)
		if err != nil {
			return nil, err
		}
		return nil, sdkerrors.NewSDKError("API error occurred", httpRes.StatusCode, string(rawBody), httpRes)
	default:
		rawBody, err := utils.ConsumeRawBody(httpRes)
		if err != nil {
			return nil, err
		}
		return nil, sdkerrors.NewSDKError("unknown status code returned", httpRes.StatusCode, string(rawBody), httpRes)
	}

	return res, nil
}
