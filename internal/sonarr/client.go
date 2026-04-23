package sonarr

import (
	"encoding/json"
	"fmt"
	"net/url"

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

type Series struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	TvdbID int    `json:"tvdbId"`
	ImdbID string `json:"imdbId"`
	Path   string `json:"path"`
}

type Episode struct {
	ID                    int    `json:"id"`
	SeriesID              int    `json:"seriesId"`
	Title                 string `json:"title"`
	SeasonNumber          int    `json:"seasonNumber"`
	EpisodeNumber         int    `json:"episodeNumber"`
	TvdbID                int    `json:"tvdbId"`
	EpisodeFileID         int    `json:"episodeFileId"` // 0 when HasFile=false
	HasFile               bool   `json:"hasFile"`
	Monitored             bool   `json:"monitored"`
	AbsoluteEpisodeNumber int    `json:"absoluteEpisodeNumber"` // anime only
}

func (c *Client) GetAllSeries() ([]Series, error) {
	var series []Series
	if err := c.http.Get("/api/v3/series", nil, &series); err != nil {
		return nil, fmt.Errorf("sonarr GetAllSeries: %w", err)
	}
	return series, nil
}

func (c *Client) GetEpisodes(seriesID int) ([]Episode, error) {
	params := url.Values{"seriesId": {fmt.Sprint(seriesID)}}
	var episodes []Episode
	if err := c.http.Get("/api/v3/episode", params, &episodes); err != nil {
		return nil, fmt.Errorf("sonarr GetEpisodes (seriesId=%d): %w", seriesID, err)
	}
	return episodes, nil
}

// DeleteEpisodeFile does not change the episode's monitored status.
func (c *Client) DeleteEpisodeFile(episodeFileID int) error {
	if err := c.http.Delete(fmt.Sprintf("/api/v3/episodefile/%d", episodeFileID)); err != nil {
		return fmt.Errorf("sonarr DeleteEpisodeFile (fileId=%d): %w", episodeFileID, err)
	}
	return nil
}

func (c *Client) UnmonitorEpisode(episode Episode) error {
	episode.Monitored = false
	body, err := json.Marshal(episode)
	if err != nil {
		return err
	}
	if err := c.http.Put(fmt.Sprintf("/api/v3/episode/%d", episode.ID), body); err != nil {
		return fmt.Errorf("sonarr UnmonitorEpisode (id=%d): %w", episode.ID, err)
	}
	return nil
}
