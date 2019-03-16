package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/tympanix/supper/media/parse"
	"github.com/tympanix/supper/types"
)

var episodeRegexp = regexp.MustCompile(`^(.*?[\w)]+)[\W_]+?[Ss]?(\d{1,2})[Eex](\d{1,2})(?:[Ee]\d{1,2})?[\W_]*(.*)$`)

// Episode represents an episode from a TV show
type Episode struct {
	Metadata
	TypeNone
	NameX        string
	EpisodeNameX string
	EpisodeX     int
	SeasonX      int
}

// MarshalJSON returns the JSON representation of an episode
func (e *Episode) MarshalJSON() (b []byte, err error) {
	type jsonEpisode struct {
		Meta    Metadata `json:"metadata"`
		Name    string   `json:"name"`
		Episode int      `json:"episode"`
		Seasion int      `json:"season"`
		ID      string   `json:"id"`
	}

	return json.Marshal(jsonEpisode{
		e.Metadata,
		e.TVShow(),
		e.Episode(),
		e.Season(),
		e.Identity(),
	})
}

// NewEpisode parses media info from a filename (without extension). The
// filename must describe the episode adequately (e.g. must contain season
// and episode numbers)
func NewEpisode(filename string) (*Episode, error) {
	groups := episodeRegexp.FindStringSubmatch(filename)

	if groups == nil {
		return nil, errors.New("could not parse episode")
	}

	name := groups[1]
	season, err := strconv.Atoi(groups[2])

	if err != nil {
		return nil, err
	}

	episode, err := strconv.Atoi(groups[3])

	if err != nil {
		return nil, err
	}

	tags := groups[4]
	end, metadata := ParseMetadataIndex(tags)

	return &Episode{
		Metadata:     metadata,
		NameX:        parse.CleanName(name),
		EpisodeX:     episode,
		SeasonX:      season,
		EpisodeNameX: parse.CleanName(tags[:end]),
	}, nil
}

func (e *Episode) String() string {
	return fmt.Sprintf("%s S%02dE%02d", e.TVShow(), e.Season(), e.Episode())
}

// Identity returns the identity string of the episode which can be used for
// hashing, caching ect.
func (e *Episode) Identity() string {
	return fmt.Sprintf("%s:%v:%v", parse.Identity(e.TVShow()), e.Season(), e.Episode())
}

// TVShow is the name of the TV show
func (e *Episode) TVShow() string {
	return e.NameX
}

// IsVideo returns true since an TV episode is also a video
func (e *Episode) IsVideo() bool {
	return true
}

// Merge merges metadata from another episode
func (e *Episode) Merge(other types.Media) error {
	if !e.Similar(other) {
		return errors.New("invalid merge episodes not similar")
	}
	if episode, ok := other.TypeEpisode(); ok {
		e.NameX = episode.TVShow()
		e.EpisodeNameX = episode.EpisodeName()
		return nil
	}
	return errors.New("invalid media merge of different media")
}

// Similar returns true if other media is also an episode from the same season
// and with the same episode number
func (e *Episode) Similar(other types.Media) bool {
	if o, ok := other.TypeEpisode(); ok {
		if e.Season() != o.Season() || e.Episode() != o.Episode() {
			return false
		}
		return true
	}
	return false
}

// Meta returns the metadata interface for the episode
func (e *Episode) Meta() types.Metadata {
	return e.Metadata
}

// TypeEpisode returns true since an episode is an episode
func (e *Episode) TypeEpisode() (types.Episode, bool) {
	return e, true
}

// EpisodeName is the name of the episode
func (e *Episode) EpisodeName() string {
	return e.EpisodeNameX
}

// Episode is the episode number in the season
func (e *Episode) Episode() int {
	return e.EpisodeX
}

// Season is the season number of the show
func (e *Episode) Season() int {
	return e.SeasonX
}
