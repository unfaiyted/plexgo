// Code generated by Speakeasy (https://speakeasy.com). DO NOT EDIT.

package plexgo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/unfaiyted/plexgo/internal/hooks"
	"github.com/unfaiyted/plexgo/internal/utils"
	"github.com/unfaiyted/plexgo/models/components"
	"github.com/unfaiyted/plexgo/retry"
	"net/http"
	"time"
)

// ServerList contains the list of servers available to the SDK
var ServerList = []string{
	// The full address of your Plex Server
	"{protocol}://{ip}:{port}",
}

// HTTPClient provides an interface for suplying the SDK with a custom HTTP client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// String provides a helper function to return a pointer to a string
func String(s string) *string { return &s }

// Bool provides a helper function to return a pointer to a bool
func Bool(b bool) *bool { return &b }

// Int provides a helper function to return a pointer to an int
func Int(i int) *int { return &i }

// Int64 provides a helper function to return a pointer to an int64
func Int64(i int64) *int64 { return &i }

// Float32 provides a helper function to return a pointer to a float32
func Float32(f float32) *float32 { return &f }

// Float64 provides a helper function to return a pointer to a float64
func Float64(f float64) *float64 { return &f }

// Pointer provides a helper function to return a pointer to a type
func Pointer[T any](v T) *T { return &v }

type sdkConfiguration struct {
	Client            HTTPClient
	Security          func(context.Context) (interface{}, error)
	ServerURL         string
	ServerIndex       int
	ServerDefaults    []map[string]string
	Language          string
	OpenAPIDocVersion string
	SDKVersion        string
	GenVersion        string
	UserAgent         string
	RetryConfig       *retry.Config
	Hooks             *hooks.Hooks
	Timeout           *time.Duration
}

func (c *sdkConfiguration) GetServerDetails() (string, map[string]string) {
	if c.ServerURL != "" {
		return c.ServerURL, nil
	}

	return ServerList[c.ServerIndex], c.ServerDefaults[c.ServerIndex]
}

func (c *sdkConfiguration) GetURIRoot(machineIdentifer string) (string) {
	return fmt.Sprintf("server://%s/com.plexapp.plugins.library", machineIdentifer)
}

// PlexAPI - Plex-API: An Open API Spec for interacting with Plex.tv and Plex Media Server
// # Plex Media Server OpenAPI Specification
//
// # An Open Source OpenAPI Specification for Plex Media Server
//
// Automation and SDKs provided by [Speakeasy](https://speakeasyapi.dev/)
//
// ## Documentation
//
// [API Documentation](https://plexapi.dev)
//
// ## SDKs
//
// The following SDKs are generated from the OpenAPI Specification. They are automatically generated and may not be fully tested. If you find any issues, please open an issue on the [main specification Repository](https://github.com/unfaiyted/plex-api-spec).
//
// | Language              | Repository                                        | Releases                                                                                         | Other                                                   |
// | --------------------- | ------------------------------------------------- | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------- |
// | Python                | [GitHub](https://github.com/unfaiyted/plexpy)     | [PyPI](https://pypi.org/project/plex-api-client/)                                                | -                                                       |
// | JavaScript/TypeScript | [GitHub](https://github.com/unfaiyted/plexjs)     | [NPM](https://www.npmjs.com/package/@lukehagar/plexjs) \ [JSR](https://jsr.io/@lukehagar/plexjs) | -                                                       |
// | Go                    | [GitHub](https://github.com/unfaiyted/plexgo)     | [Releases](https://github.com/unfaiyted/plexgo/releases)                                         | [GoDoc](https://pkg.go.dev/github.com/unfaiyted/plexgo) |
// | Ruby                  | [GitHub](https://github.com/unfaiyted/plexruby)   | [Releases](https://github.com/unfaiyted/plexruby/releases)                                       | -                                                       |
// | Swift                 | [GitHub](https://github.com/unfaiyted/plexswift)  | [Releases](https://github.com/unfaiyted/plexswift/releases)                                      | -                                                       |
// | PHP                   | [GitHub](https://github.com/unfaiyted/plexphp)    | [Releases](https://github.com/unfaiyted/plexphp/releases)                                        | -                                                       |
// | Java                  | [GitHub](https://github.com/unfaiyted/plexjava)   | [Releases](https://github.com/unfaiyted/plexjava/releases)                                       | -                                                       |
// | C#                    | [GitHub](https://github.com/unfaiyted/plexcsharp) | [Releases](https://github.com/unfaiyted/plexcsharp/releases)                                     | -
type PlexAPI struct {
	// Operations against the Plex Media Server System.
	//
	Server *Server
	// API Calls interacting with Plex Media Server Media
	//
	Media *Media
	// API Calls that perform operations with Plex Media Server Videos
	//
	Video *Video
	// Activities are awesome. They provide a way to monitor and control asynchronous operations on the server. In order to receive real-time updates for activities, a client would normally subscribe via either EventSource or Websocket endpoints.
	// Activities are associated with HTTP replies via a special `X-Plex-Activity` header which contains the UUID of the activity.
	// Activities are optional cancellable. If cancellable, they may be cancelled via the `DELETE` endpoint. Other details:
	// - They can contain a `progress` (from 0 to 100) marking the percent completion of the activity.
	// - They must contain an `type` which is used by clients to distinguish the specific activity.
	// - They may contain a `Context` object with attributes which associate the activity with various specific entities (items, libraries, etc.)
	// - The may contain a `Response` object which attributes which represent the result of the asynchronous operation.
	//
	Activities *Activities
	// Butler is the task manager of the Plex Media Server Ecosystem.
	//
	Butler *Butler
	// API Calls that perform operations directly against https://Plex.tv
	//
	Plex *Plex
	// Hubs are a structured two-dimensional container for media, generally represented by multiple horizontal rows.
	//
	Hubs *Hubs
	// API Calls that perform search operations with Plex Media Server
	//
	Search *Search
	// API Calls interacting with Plex Media Server Libraries
	//
	Library *Library
	// API Calls that perform operations with Plex Media Server Watchlists
	//
	Watchlist *Watchlist
	// Submit logs to the Log Handler for Plex Media Server
	//
	Log *Log
	// Playlists are ordered collections of media. They can be dumb (just a list of media) or smart (based on a media query, such as "all albums from 2017").
	// They can be organized in (optionally nesting) folders.
	// Retrieving a playlist, or its items, will trigger a refresh of its metadata.
	// This may cause the duration and number of items to change.
	//
	Playlists *Playlists
	// API Calls regarding authentication for Plex Media Server
	//
	Authentication *Authentication
	// API Calls that perform operations with Plex Media Server Statistics
	//
	Statistics *Statistics
	// API Calls that perform search operations with Plex Media Server Sessions
	//
	Sessions *Sessions
	// This describes the API for searching and applying updates to the Plex Media Server.
	// Updates to the status can be observed via the Event API.
	//
	Updater *Updater
	// API Calls that perform operations with Plex Media Server Users
	//
	Users *Users
	// Collections are groupings of media items in a library. They can be regular (manually curated) or smart (based on filters).
	// Collections support different view modes and sort orders, and can have visibility settings for library, home, and shared users.
	//
	Collections *Collections

	sdkConfiguration sdkConfiguration
}

type SDKOption func(*PlexAPI)

// WithServerURL allows the overriding of the default server URL
func WithServerURL(serverURL string) SDKOption {
	return func(sdk *PlexAPI) {
		sdk.sdkConfiguration.ServerURL = serverURL
	}
}

// WithTemplatedServerURL allows the overriding of the default server URL with a templated URL populated with the provided parameters
func WithTemplatedServerURL(serverURL string, params map[string]string) SDKOption {
	return func(sdk *PlexAPI) {
		if params != nil {
			serverURL = utils.ReplaceParameters(serverURL, params)
		}

		sdk.sdkConfiguration.ServerURL = serverURL
	}
}

// WithServerIndex allows the overriding of the default server by index
func WithServerIndex(serverIndex int) SDKOption {
	return func(sdk *PlexAPI) {
		if serverIndex < 0 || serverIndex >= len(ServerList) {
			panic(fmt.Errorf("server index %d out of range", serverIndex))
		}

		sdk.sdkConfiguration.ServerIndex = serverIndex
	}
}

// ServerProtocol - The protocol to use for the server connection
type ServerProtocol string

const (
	ServerProtocolHTTP  ServerProtocol = "http"
	ServerProtocolHTTPS ServerProtocol = "https"
)

func (e ServerProtocol) ToPointer() *ServerProtocol {
	return &e
}
func (e *ServerProtocol) UnmarshalJSON(data []byte) error {
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch v {
	case "http":
		fallthrough
	case "https":
		*e = ServerProtocol(v)
		return nil
	default:
		return fmt.Errorf("invalid value for ServerProtocol: %v", v)
	}
}

// WithProtocol allows setting the protocol variable for url substitution
func WithProtocol(protocol ServerProtocol) SDKOption {
	return func(sdk *PlexAPI) {
		for idx := range sdk.sdkConfiguration.ServerDefaults {
			if _, ok := sdk.sdkConfiguration.ServerDefaults[idx]["protocol"]; !ok {
				continue
			}

			sdk.sdkConfiguration.ServerDefaults[idx]["protocol"] = fmt.Sprintf("%v", protocol)
		}
	}
}

// WithIP allows setting the ip variable for url substitution
func WithIP(ip string) SDKOption {
	return func(sdk *PlexAPI) {
		for idx := range sdk.sdkConfiguration.ServerDefaults {
			if _, ok := sdk.sdkConfiguration.ServerDefaults[idx]["ip"]; !ok {
				continue
			}

			sdk.sdkConfiguration.ServerDefaults[idx]["ip"] = fmt.Sprintf("%v", ip)
		}
	}
}

// WithPort allows setting the port variable for url substitution
func WithPort(port string) SDKOption {
	return func(sdk *PlexAPI) {
		for idx := range sdk.sdkConfiguration.ServerDefaults {
			if _, ok := sdk.sdkConfiguration.ServerDefaults[idx]["port"]; !ok {
				continue
			}

			sdk.sdkConfiguration.ServerDefaults[idx]["port"] = fmt.Sprintf("%v", port)
		}
	}
}

// WithClient allows the overriding of the default HTTP client used by the SDK
func WithClient(client HTTPClient) SDKOption {
	return func(sdk *PlexAPI) {
		sdk.sdkConfiguration.Client = client
	}
}

// WithSecurity configures the SDK to use the provided security details
func WithSecurity(accessToken string) SDKOption {
	return func(sdk *PlexAPI) {
		security := components.Security{AccessToken: &accessToken}
		sdk.sdkConfiguration.Security = utils.AsSecuritySource(&security)
	}
}

// WithSecuritySource configures the SDK to invoke the Security Source function on each method call to determine authentication
func WithSecuritySource(security func(context.Context) (components.Security, error)) SDKOption {
	return func(sdk *PlexAPI) {
		sdk.sdkConfiguration.Security = func(ctx context.Context) (interface{}, error) {
			return security(ctx)
		}
	}
}

func WithRetryConfig(retryConfig retry.Config) SDKOption {
	return func(sdk *PlexAPI) {
		sdk.sdkConfiguration.RetryConfig = &retryConfig
	}
}

// WithTimeout Optional request timeout applied to each operation
func WithTimeout(timeout time.Duration) SDKOption {
	return func(sdk *PlexAPI) {
		sdk.sdkConfiguration.Timeout = &timeout
	}
}

// New creates a new instance of the SDK with the provided options
func New(opts ...SDKOption) *PlexAPI {
	sdk := &PlexAPI{
		sdkConfiguration: sdkConfiguration{
			Language:          "go",
			OpenAPIDocVersion: "0.0.3",
			SDKVersion:        "0.21.2",
			GenVersion:        "2.597.9",
			UserAgent:         "speakeasy-sdk/go 0.21.2 2.597.9 0.0.3 github.com/unfaiyted/plexgo",
			ServerDefaults: []map[string]string{
				{
					"protocol": "https",
					"ip":       "10.10.10.47",
					"port":     "32400",
				},
			},
			Hooks: hooks.New(),
		},
	}
	for _, opt := range opts {
		opt(sdk)
	}

	// Use WithClient to override the default client if you would like to customize the timeout
	if sdk.sdkConfiguration.Client == nil {
		sdk.sdkConfiguration.Client = &http.Client{Timeout: 60 * time.Second}
	}

	currentServerURL, _ := sdk.sdkConfiguration.GetServerDetails()
	serverURL := currentServerURL
	serverURL, sdk.sdkConfiguration.Client = sdk.sdkConfiguration.Hooks.SDKInit(currentServerURL, sdk.sdkConfiguration.Client)
	if serverURL != currentServerURL {
		sdk.sdkConfiguration.ServerURL = serverURL
	}

	sdk.Server = newServer(sdk.sdkConfiguration)

	sdk.Media = newMedia(sdk.sdkConfiguration)

	sdk.Video = newVideo(sdk.sdkConfiguration)

	sdk.Activities = newActivities(sdk.sdkConfiguration)

	sdk.Butler = newButler(sdk.sdkConfiguration)

	sdk.Plex = newPlex(sdk.sdkConfiguration)

	sdk.Hubs = newHubs(sdk.sdkConfiguration)

	sdk.Search = newSearch(sdk.sdkConfiguration)

	sdk.Library = newLibrary(sdk.sdkConfiguration)

	sdk.Watchlist = newWatchlist(sdk.sdkConfiguration)

	sdk.Log = newLog(sdk.sdkConfiguration)

	sdk.Playlists = newPlaylists(sdk.sdkConfiguration)

	sdk.Authentication = newAuthentication(sdk.sdkConfiguration)

	sdk.Statistics = newStatistics(sdk.sdkConfiguration)

	sdk.Sessions = newSessions(sdk.sdkConfiguration)

	sdk.Updater = newUpdater(sdk.sdkConfiguration)

	sdk.Users = newUsers(sdk.sdkConfiguration)

	sdk.Collections = newCollections(sdk.sdkConfiguration)

	return sdk
}
