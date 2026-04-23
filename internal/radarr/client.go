package radarr

import (
	"fmt"

	"github.com/bvdwalt/cullarr/internal/httpclient"
)

type Client struct {
	http *httpclient.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		http: httpclient.New(baseURL, "X-Api-Key", apiKey),
	}
}

type Movie struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	ImdbID           string `json:"imdbId"`
	TmdbID           int    `json:"tmdbId"`
	HasFile          bool   `json:"hasFile"`
	MovieFileID      int    `json:"movieFileId"`
	Monitored        bool   `json:"monitored"`
	Path             string `json:"path"`
	QualityProfileID int    `json:"qualityProfileId"`
}

func (c *Client) GetAllMovies() ([]Movie, error) {
	var movies []Movie
	if err := c.http.Get("/api/v3/movie", nil, &movies); err != nil {
		return nil, fmt.Errorf("radarr GetAllMovies: %w", err)
	}
	return movies, nil
}

// DeleteMovieFile does not change the movie's monitored status.
func (c *Client) DeleteMovieFile(movieFileID int) error {
	if err := c.http.Delete(fmt.Sprintf("/api/v3/moviefile/%d", movieFileID)); err != nil {
		return fmt.Errorf("radarr DeleteMovieFile (fileId=%d): %w", movieFileID, err)
	}
	return nil
}

func (c *Client) DeleteMovie(id int) error {
	if err := c.http.Delete(fmt.Sprintf("/api/v3/movie/%d", id)); err != nil {
		return fmt.Errorf("radarr DeleteMovie (id=%d): %w", id, err)
	}
	return nil
}
