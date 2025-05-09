# PlexGo Collections Feature Implementation Plan

This document outlines the plan for extending the PlexGo SDK to support comprehensive collection management functionality, based on the Python implementation in the provided code.

## Current State

The current PlexGo SDK has limited support for collections:

- Can retrieve collections through library items via `GetLibraryItems` with `TagCollection` 
- No direct support for creating, updating, or managing collections
- No smart collection support
- No ability to add/remove items from collections

## Implementation Plan

### 1. New Types

We'll need to implement the following types:

```go
// Collection represents a Plex collection
type Collection struct {
    ID              int64
    RatingKey       string
    Key             string
    GUID            string
    Title           string
    TitleSort       string
    Summary         string
    Smart           bool
    AddedAt         int64
    UpdatedAt       int64
    ContentRating   string
    Thumb           string
    Art             string
    ChildCount      int
    Items           []CollectionItem // lazily loaded
    CollectionMode  string // "default", "hide", "hideItems", "showItems"
    CollectionSort  string // "release", "alpha", "custom"
}

// CollectionItem represents an item in a collection
type CollectionItem struct {
    ID        int64
    RatingKey string
    Title     string
    Type      string
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
}
```

### 2. Core Collection Functions

We'll implement the following functions in a new `collections.go` file:

```go
// GetCollection gets a collection by ID or title
func (s *PlexGo) GetCollection(ctx context.Context, identifier interface{}) (*Collection, error)

// GetAllCollections gets all collections, optionally filtered by label
func (s *PlexGo) GetAllCollections(ctx context.Context, label string) ([]*Collection, error)

// CreateCollection creates a new collection with the given items
func (s *PlexGo) CreateCollection(ctx context.Context, title string, items []int64) (*Collection, error)

// CreateSmartCollection creates a new smart collection with the given filter
func (s *PlexGo) CreateSmartCollection(ctx context.Context, title string, filter SmartFilterConfig) (*Collection, error)

// DeleteCollection deletes a collection
func (s *PlexGo) DeleteCollection(ctx context.Context, identifier interface{}) error

// AddToCollection adds items to a collection
func (s *PlexGo) AddToCollection(ctx context.Context, collection interface{}, items []int64) error

// RemoveFromCollection removes items from a collection
func (s *PlexGo) RemoveFromCollection(ctx context.Context, collection interface{}, items []int64) error

// MoveItem moves an item in a collection to a new position
func (s *PlexGo) MoveItem(ctx context.Context, collection interface{}, item int64, after int64) error
```

### 3. Collection Settings Functions

```go
// GetCollectionMode gets the mode of a collection
func (s *Collection) GetCollectionMode() (string, error)

// SetCollectionMode sets the mode of a collection
func (s *Collection) SetCollectionMode(mode string) error

// GetCollectionSort gets the sort order of a collection
func (s *Collection) GetCollectionSort() (string, error)

// SetCollectionSort sets the sort order of a collection
func (s *Collection) SetCollectionSort(sort string) error

// GetCollectionVisibility gets the visibility of a collection
func (s *Collection) GetCollectionVisibility() (*CollectionVisibility, error)

// SetCollectionVisibility sets the visibility of a collection
func (s *Collection) SetCollectionVisibility(visibility *CollectionVisibility) error
```

### 4. Smart Filter Functions

```go
// GetSmartFilter gets the smart filter for a collection
func (s *Collection) GetSmartFilter() (string, error)

// UpdateSmartFilter updates the smart filter for a collection
func (s *Collection) UpdateSmartFilter(filter SmartFilterConfig) error

// BuildSmartFilter builds a smart filter URI from parameters
func (s *PlexGo) BuildSmartFilter(paramString string) string
```

### 5. Collection Item Management

```go
// GetItems gets all items in a collection
func (s *Collection) GetItems() ([]CollectionItem, error)

// AddItems adds items to a collection
func (s *Collection) AddItems(items []int64) error

// RemoveItems removes items from a collection
func (s *Collection) RemoveItems(items []int64) error

// MoveItem moves an item in a collection
func (s *Collection) MoveItem(item int64, after int64) error
```

## API Endpoints to Implement

Based on the Python code, we need to implement the following Plex API endpoints:

1. `/library/collections` - GET to list collections, POST to create collections
2. `/library/collections/{id}` - GET to get collection details, DELETE to delete collection
3. `/library/collections/{id}/items` - GET to list items, PUT to update items (for smart collections)
4. `/library/collections/{id}/items/{itemID}/move` - PUT to move items
5. `/hubs/sections/{sectionId}/manage` - GET for visibility, POST to update visibility

## Implementation Approach

1. Start by implementing the basic Collection type and GET methods
2. Implement collection creation and deletion
3. Implement item manipulation (add/remove/move)
4. Implement smart collections
5. Implement collection settings (mode/sort/visibility)

## Priority Features

1. `GetAllCollections` - List all collections
2. `GetCollection` - Get a specific collection
3. `CreateCollection` - Create a collection
4. `AddToCollection` - Add items to a collection
5. `RemoveFromCollection` - Remove items from a collection

## Testing

For each implemented function, we should:

1. Create unit tests with mock HTTP responses
2. Document example usage in README
3. Verify functionality against a real Plex server