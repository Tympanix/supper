package list

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tympanix/supper/media/meta/codec"
	"github.com/tympanix/supper/media/meta/misc"
	"github.com/tympanix/supper/media/meta/quality"
	"github.com/tympanix/supper/media/meta/source"
	"github.com/tympanix/supper/types"
	"golang.org/x/text/language"
)

var errMock = errors.New("unsupported for mocked object")

func quote(s string) []byte {
	return []byte("\"" + s + "\"")
}

type fakemedia struct{}

func (fakemedia) Identity() string                     { return "" }
func (fakemedia) Merge(types.Media) error              { return nil }
func (fakemedia) TypeEpisode() (types.Episode, bool)   { return nil, false }
func (fakemedia) TypeMovie() (types.Movie, bool)       { return nil, false }
func (fakemedia) TypeSubtitle() (types.Subtitle, bool) { return nil, false }
func (fakemedia) Similar(other types.Media) bool       { return false }

type fakelocal struct{ types.Media }

func (fakelocal) IsDir() bool                    { return false }
func (fakelocal) ModTime() time.Time             { return time.Now() }
func (fakelocal) Mode() os.FileMode              { return 0 }
func (fakelocal) Size() int64                    { return 0 }
func (fakelocal) Sys() interface{}               { return nil }
func (fakelocal) Name() string                   { return "" }
func (fakelocal) Path() string                   { return "" }
func (l fakelocal) MarshalJSON() ([]byte, error) { return json.Marshal(l.Media) }

type metadata struct{}

func (m metadata) Meta() types.Metadata { return m }
func (m metadata) Source() source.Tag   { return source.None }
func (m metadata) Quality() quality.Tag { return quality.None }
func (m metadata) Codec() codec.Tag     { return codec.None }
func (m metadata) Group() string        { return "" }
func (m metadata) AllTags() []string    { return nil }
func (m metadata) String() string       { return "" }
func (m metadata) Misc() misc.List      { return nil }

type fakevideo struct {
	fakelocal
}

func (fakevideo) ExistingSubtitles() (types.SubtitleList, error) {
	return nil, errMock
}
func (fakevideo) SaveSubtitle(io.Reader, language.Tag) (types.LocalSubtitle, error) {
	return nil, errMock
}
func (v fakevideo) MarshalJSON() ([]byte, error) { return json.Marshal(v.fakelocal) }

type movie struct {
	name string
	year int
	metadata
	fakemedia
}

func (m movie) MovieName() string              { return m.name }
func (m movie) Year() int                      { return m.year }
func (m movie) TypeMovie() (types.Movie, bool) { return m, true }
func (m movie) String() string                 { return m.name }
func (m movie) MarshalJSON() ([]byte, error)   { return quote(m.name), nil }

type episode struct {
	show    string
	episode int
	season  int
	metadata
	fakemedia
}

func (e episode) TVShow() string                     { return e.show }
func (e episode) Season() int                        { return e.season }
func (e episode) Episode() int                       { return e.episode }
func (e episode) EpisodeName() string                { return "" }
func (e episode) TypeEpisode() (types.Episode, bool) { return e, true }
func (e episode) String() string                     { return e.show }
func (e episode) MarshalJSON() ([]byte, error)       { return quote(e.show), nil }

type subtitle struct {
	types.Media
	lang language.Tag
	hi   bool
}

func (s subtitle) String() string                       { return "Subtitle: " + s.Media.String() }
func (s subtitle) ForMedia() types.Media                { return s.Media }
func (s subtitle) HearingImpaired() bool                { return s.hi }
func (s subtitle) Language() language.Tag               { return s.lang }
func (s subtitle) TypeSubtitle() (types.Subtitle, bool) { return s, true }
func (s subtitle) TypeEpisode() (types.Episode, bool)   { return nil, false }
func (s subtitle) TypeMovie() (types.Movie, bool)       { return nil, false }
func (s subtitle) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Subtitle types.Media
	}{
		s.Media,
	})
}

var inception = movie{name: "Inception", year: 2010}
var fightclub = movie{name: "Fight Club", year: 1999}
var batmanbegins = movie{name: "Batman Begins", year: 2005}

var theoffice = episode{show: "The Office", season: 1, episode: 2}
var arrow = episode{show: "Arrow", season: 2, episode: 7}
var westworld = episode{show: "Westworld", season: 1, episode: 5}

var movies = []types.LocalMedia{
	fakevideo{fakelocal{inception}},
	fakevideo{fakelocal{fightclub}},
	fakevideo{fakelocal{batmanbegins}},
}

var episodes = []types.LocalMedia{
	fakevideo{fakelocal{theoffice}},
	fakevideo{fakelocal{arrow}},
	fakevideo{fakelocal{westworld}},
}

var videos = append(movies, episodes...)

var subtitles = []types.LocalMedia{
	fakelocal{subtitle{inception, language.English, false}},
	fakelocal{subtitle{fightclub, language.German, false}},
	fakelocal{subtitle{batmanbegins, language.French, false}},

	fakelocal{subtitle{theoffice, language.Spanish, false}},
	fakelocal{subtitle{arrow, language.Italian, false}},
	fakelocal{subtitle{westworld, language.Chinese, false}},
}

var list = NewLocalMedia(
	append(append(movies, episodes...), subtitles...)...,
)

func TestLocalMedia(t *testing.T) {
	assert.Equal(t, 12, list.Len())

	// test video media
	vid := list.FilterVideo()
	assert.Equal(t, 6, vid.Len())
	assert.Subset(t, vid.List(), videos)

	// test subtitle media
	subs := list.FilterSubtitles()
	assert.Equal(t, 6, subs.Len())
	assert.Subset(t, subs.List(), subtitles)

	// test movie media
	mov := list.FilterMovies()
	assert.Equal(t, 3, mov.Len())
	assert.Subset(t, mov.List(), movies)

	// test episode media
	epi := list.FilterEpisodes()
	assert.Equal(t, 3, epi.Len())
	assert.Subset(t, epi.List(), episodes)

	// test modified in past media
	mod := list.FilterModified(1 * time.Millisecond)
	assert.Equal(t, 12, mod.Len())

	// test modified in future media
	none := list.FilterModified(-1 * time.Millisecond)
	assert.Equal(t, 0, none.Len())

	p := func(m types.Media) bool { return m == fakevideo{fakelocal{inception}} }
	fil := list.Filter(p)
	assert.Equal(t, 1, fil.Len())
	assert.Contains(t, fil.List(), fakevideo{fakelocal{inception}})
}

func TestMediaJSON(t *testing.T) {
	data, err := json.Marshal(list)
	require.NoError(t, err)

	str := string(data)
	for _, m := range list.List() {
		item, err := json.Marshal(m)
		require.NoError(t, err)
		assert.Contains(t, str, string(item))
	}
}
