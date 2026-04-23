package jellyfin

import (
	"fmt"
	"net/url"

	"github.com/bvdwalt/cullarr/internal/httpclient"
)

type Client struct {
	http *httpclient.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		http: httpclient.New(baseURL, "X-Emby-Token", apiKey),
	}
}

type User struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
}

type Item struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
	Type string `json:"Type"` // "Episode" or "Movie"

	SeriesName        string `json:"SeriesName"`
	ParentIndexNumber int    `json:"ParentIndexNumber"` // season number
	IndexNumber       int    `json:"IndexNumber"`       // episode number

	// ProviderIds holds external IDs: "Tvdb", "Imdb", "Tmdb" etc.
	ProviderIds map[string]string `json:"ProviderIds"`

	UserData struct {
		Played bool `json:"Played"`
	} `json:"UserData"`
}

type itemsResponse struct {
	Items []Item `json:"Items"`
}

// GetUsers requires an admin API key.
func (c *Client) GetUsers() ([]User, error) {
	var users []User
	if err := c.http.Get("/Users", nil, &users); err != nil {
		return nil, fmt.Errorf("jellyfin GetUsers: %w", err)
	}
	return users, nil
}

func (c *Client) GetWatchedEpisodes(userID string) ([]Item, error) {
	params := url.Values{
		"IncludeItemTypes": {"Episode"},
		"Filters":          {"IsPlayed"},
		"Recursive":        {"true"},
		"Fields":           {"ProviderIds,ParentIndexNumber,IndexNumber,SeriesName"},
	}
	var resp itemsResponse
	if err := c.http.Get(fmt.Sprintf("/Users/%s/Items", userID), params, &resp); err != nil {
		return nil, fmt.Errorf("jellyfin GetWatchedEpisodes (user %s): %w", userID, err)
	}
	return resp.Items, nil
}

func (c *Client) GetWatchedMovies(userID string) ([]Item, error) {
	params := url.Values{
		"IncludeItemTypes": {"Movie"},
		"Filters":          {"IsPlayed"},
		"Recursive":        {"true"},
		"Fields":           {"ProviderIds"},
	}
	var resp itemsResponse
	if err := c.http.Get(fmt.Sprintf("/Users/%s/Items", userID), params, &resp); err != nil {
		return nil, fmt.Errorf("jellyfin GetWatchedMovies (user %s): %w", userID, err)
	}
	return resp.Items, nil
}
