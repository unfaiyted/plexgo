# Plex Collections API

This document describes the Collections API extension for the PlexGo SDK. The Collections API allows you to manage collections within your Plex Media Server.

## Table of Contents

- [Overview](#overview)
- [Collection Types](#collection-types)
- [Collection Modes](#collection-modes)
- [Collection Sorting](#collection-sorting)
- [Collection Visibility](#collection-visibility)
- [API Methods](#api-methods)
- [Examples](#examples)

## Overview

Collections are a way to group media items in Plex. They can be either regular collections (manually curated) or smart collections (dynamically populated based on filters).

The Collections API provides methods to:
- Retrieve collections from a library
- Create new collections (both regular and smart)
- Add and remove items from collections
- Update collection settings (display mode, sort order)
- Manage collection visibility
- Delete collections

## Collection Types

There are two types of collections in Plex:

1. **Regular Collections**: These are manually curated collections where you explicitly add and remove items.
2. **Smart Collections**: These are dynamically populated based on filters (similar to smart playlists).

## Collection Modes

Collections can have different display modes that control how they appear in the library:

- `default`: Use the library default setting
- `hide`: Hide the collection
- `hideItems`: Show the collection but hide its items from the library
- `showItems`: Show both the collection and its items in the library

These modes are available as constants in the SDK:
```go
plexgo.CollectionModeDefault  // "default"
plexgo.CollectionModeHide     // "hide"
plexgo.CollectionModeHideItems // "hideItems"
plexgo.CollectionModeShowItems // "showItems"
```

## Collection Sorting

Collections can be sorted in different ways:

- `release`: Sort by release date
- `alpha`: Sort alphabetically
- `custom`: Custom sort order

These sort options are available as constants in the SDK:
```go
plexgo.CollectionSortRelease // "release" 
plexgo.CollectionSortAlpha   // "alpha"
plexgo.CollectionSortCustom  // "custom"
```

## Collection Visibility

Collections have visibility settings that control where they appear:

- `Library`: Whether the collection appears in the library
- `Home`: Whether the collection appears in the owner's home screen
- `Shared`: Whether the collection appears in shared users' home screens

These settings are managed through the `CollectionVisibility` struct:
```go
type CollectionVisibility struct {
    Library bool
    Home    bool
    Shared  bool
}
```

## API Methods

The Collections API includes the following methods:

### GetAllCollections

```go
func (s *Collections) GetAllCollections(ctx context.Context, sectionID int, opts ...Option) ([]Collection, error)
```

Retrieves all collections in a library section.

### GetCollection

```go
func (s *Collections) GetCollection(ctx context.Context, collectionID int, opts ...Option) (*Collection, error)
```

Retrieves a specific collection by ID.

### GetCollectionItems

```go
func (s *Collections) GetCollectionItems(ctx context.Context, collectionID int, opts ...Option) ([]string, error)
```

Retrieves all items in a collection.

### CreateCollection

```go
func (s *Collections) CreateCollection(ctx context.Context, sectionID int, title string, itemIDs []string, opts ...Option) (*Collection, error)
```

Creates a new collection with the specified items.

### CreateSmartCollection

```go
func (s *Collections) CreateSmartCollection(ctx context.Context, sectionID int, title string, smartType int, filterArgs string, opts ...Option) (*Collection, error)
```

Creates a new smart collection with a filter.

### DeleteCollection

```go
func (s *Collections) DeleteCollection(ctx context.Context, collectionID int, opts ...Option) error
```

Deletes a collection.

### AddToCollection

```go
func (s *Collections) AddToCollection(ctx context.Context, collectionID int, itemIDs []string, opts ...Option) error
```

Adds items to an existing collection.

### RemoveFromCollection

```go
func (s *Collections) RemoveFromCollection(ctx context.Context, collectionID int, itemIDs []string, opts ...Option) error
```

Removes items from a collection.

### UpdateCollectionMode

```go
func (s *Collections) UpdateCollectionMode(ctx context.Context, collectionID int, mode string, opts ...Option) error
```

Updates the display mode of a collection.

### UpdateCollectionSort

```go
func (s *Collections) UpdateCollectionSort(ctx context.Context, collectionID int, sort string, opts ...Option) error
```

Updates the sort order of a collection.

### GetCollectionVisibility

```go
func (s *Collections) GetCollectionVisibility(ctx context.Context, sectionID int, collectionID int, opts ...Option) (*CollectionVisibility, error)
```

Gets the visibility settings for a collection.

### UpdateCollectionVisibility

```go
func (s *Collections) UpdateCollectionVisibility(ctx context.Context, sectionID int, collectionID int, visibility *CollectionVisibility, opts ...Option) error
```

Updates the visibility settings for a collection.

### UpdateSmartCollection

```go
func (s *Collections) UpdateSmartCollection(ctx context.Context, collectionID int, filterURI string, opts ...Option) error
```

Updates the smart filter for a collection.

## Examples

See [collections_example.go](../examples/collections_example.go) for complete examples.

### Basic Usage

```go
// Create a new Plex API client
client := plexgo.New(
    plexgo.WithSecurity("<YOUR-PLEX-TOKEN>"),
    plexgo.WithIP("<YOUR-PLEX-SERVER-IP>"),
    plexgo.WithPort("<YOUR-PLEX-SERVER-PORT>"),
)

// Get all collections from a library section
collections, err := client.Collections.GetAllCollections(context.Background(), 1)
if err != nil {
    log.Fatalf("Error getting collections: %v", err)
}

// Create a new collection
newCollection, err := client.Collections.CreateCollection(
    context.Background(),
    1,                       // Library section ID
    "My New Collection",     // Collection title
    []string{"1234", "5678"}, // Item IDs to add to the collection
)

// Update collection mode
err = client.Collections.UpdateCollectionMode(
    context.Background(),
    collectionID,
    plexgo.CollectionModeShowItems, // Use the predefined constant
)
```