package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/tympanix/supper/media"
	"github.com/tympanix/supper/types"
)

const tmdbHost = "https://api.themoviedb.org/3/"

const tmdbTimeFormat = "2006-01-02"

var tmdbClient = NewAPIClient("TheMovieDB", 35)

var tmdbToken string

var tmdbCache = make(map[string]types.Media)

// TheMovieDB is a scraper for themoviedb.org
func TheMovieDB(token string) types.Scraper {
	if token == "" {
		token = tmdbToken
	}

	return &tmdb{
		client: tmdbClient,
		token:  token,
	}
}

type tmdb struct {
	client *APIClient
	token  string
}

func (t *tmdb) Scrape(m types.Media) (src types.Media, err error) {
	if t.token == "" {
		return nil, errors.New("tmdb: missing API token")
	}
	if m == nil {
		return nil, errors.New("tmdb: can't scrape nil media")
	}
	if cache, ok := tmdbCache[m.Identity()]; ok {
		return cache, nil
	}
	if movie, ok := m.TypeMovie(); ok {
		src, err = t.searchMovie(movie)
	} else if sub, ok := m.TypeSubtitle(); ok {
		src, err = t.Scrape(sub.ForMedia())
	} else {
		return nil, mediaNotSupported("tmdb")
	}
	if err == nil {
		tmdbCache[m.Identity()] = src
	}
	return
}

func (t *tmdb) url(p string) (*url.URL, error) {
	url, err := url.Parse(tmdbHost)

	url.Path = path.Join(url.Path, p)

	if err != nil {
		return nil, err
	}

	q := url.Query()
	q.Set("api_key", t.token)

	url.RawQuery = q.Encode()
	return url, nil
}

func (t *tmdb) searchMovie(m types.Movie) (types.Media, error) {
	url, err := t.url("/search/movie")

	if err != nil {
		return nil, err
	}

	q := url.Query()
	q.Set("query", m.MovieName())
	q.Set("year", strconv.Itoa(m.Year()))
	url.RawQuery = q.Encode()

	resp, err := t.client.Get(url.String())

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("tmdb returned status code %v", resp.StatusCode)
	}

	defer resp.Body.Close()

	type result struct {
		Title         string `json:"title"`
		OriginalTitle string `json:"original_title"`
		ReleaseDate   string `json:"release_date"`
	}

	type response struct {
		Page         int      `json:"page"`
		Results      []result `json:"results"`
		TotalResults int      `json:"total_results"`
		TotalPages   int      `json:"total_pages"`
	}

	var res response
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	if len(res.Results) == 0 {
		return nil, errors.New("could not find media on tmdb")
	}

	d, err := time.Parse(tmdbTimeFormat, res.Results[0].ReleaseDate)

	if err != nil {
		return nil, err
	}

	movie := media.Movie{
		NameX: res.Results[0].OriginalTitle,
		YearX: d.Year(),
	}

	return &movie, nil
}
