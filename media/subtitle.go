package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tympanix/supper/media/parse"
	"github.com/tympanix/supper/types"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// Subtitle represents the information about a subtitle
type Subtitle struct {
	TypeNone
	forMedia types.Media
	lang     language.Tag
}

// NewSubtitle returns subtitle information by parsing the string. The string
// should describe some video material sufficiently (without extension). If the
// string ends with a language tag (e.g. .en .es. de) then the language will be
// parsed
func NewSubtitle(str string) (*Subtitle, error) {
	parts := strings.Split(str, ".")

	if len(parts) < 2 {
		return nil, errors.New("error parsing subtitle file")
	}

	langext := parts[len(parts)-1]
	tag := language.Make(langext)

	var medstr string
	if tag != language.Und {
		medstr = strings.TrimSuffix(str, langext)
	} else {
		medstr = str
	}

	med, err := NewFromString(medstr)
	if err != nil {
		return nil, err
	}

	return &Subtitle{
		forMedia: med,
		lang:     tag,
	}, nil
}

// HearingImpaired returns false since this information in unparseable from a simple filename
func (l *Subtitle) HearingImpaired() bool {
	return false
}

// Language returns the language of the subtitle
func (l *Subtitle) Language() language.Tag {
	return l.lang
}

// Identity returns the identity string for the media which the subtitle matches
func (l *Subtitle) Identity() string {
	return fmt.Sprintf("%v:%v", l.ForMedia().Identity(), l.Language().String())
}

// Merge merges the media belonging to the subtitle
func (l *Subtitle) Merge(other types.Media) error {
	return l.ForMedia().Merge(other)
}

// Similar returns true if other media is also a subtitle for the a similar media
func (l *Subtitle) Similar(other types.Media) bool {
	if o, ok := other.TypeSubtitle(); ok {
		return l.ForMedia().Similar(o.ForMedia())
	}
	return false
}

// String returns the language of the subtitle
func (l *Subtitle) String() string {
	return display.English.Languages().Name(l.Language())
}

// Meta returns the metadata for media which the subtitle belongs
func (l *Subtitle) Meta() types.Metadata {
	return l.forMedia.Meta()
}

// TypeSubtitle returns true, since a subtitle is a subtitle
func (l *Subtitle) TypeSubtitle() (types.Subtitle, bool) {
	return l, true
}

// ForMedia returns the media the subtitle is matched against
func (l *Subtitle) ForMedia() types.Media {
	return l.forMedia
}

// NewLocalSubtitle returns a new local subtitle
func NewLocalSubtitle(path string) (types.LocalSubtitle, error) {
	if filepath.Ext(path) != ".srt" {
		return nil, errors.New("parsing non subtitle file as subtitle")
	}

	sub, err := NewSubtitle(parse.Filename(path))

	if err != nil {
		return nil, err
	}

	file, err := os.Stat(path)

	if err != nil {
		return nil, err
	}

	return &LocalSubtitle{
		FileInfo: file,
		Pather:   FilePath(path),
		Subtitle: sub,
	}, nil
}

// LocalSubtitle represents a subtitle stored on disk
type LocalSubtitle struct {
	os.FileInfo
	types.Pather
	*Subtitle
}

// MarshalJSON returns a JSON representation of the subtitle
func (l *LocalSubtitle) MarshalJSON() (b []byte, err error) {
	return json.Marshal(struct {
		File string       `json:"filename"`
		Code language.Tag `json:"code"`
		Lang string       `json:"language"`
	}{
		l.Name(),
		l.Language(),
		l.Subtitle.String(),
	})
}
