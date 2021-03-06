package app

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/rakyll/statik/fs"

	// Importing static files for webapp
	_ "github.com/tympanix/supper/statik"

	"github.com/tympanix/supper/api"
	"github.com/tympanix/supper/app/cfg"
	"github.com/tympanix/supper/media"
	"github.com/tympanix/supper/media/list"
	"github.com/tympanix/supper/types"
)

const webRoot = "../web/build"

var filetypes = []string{
	".avi", ".mkv", ".mp4", ".m4v", ".flv", ".mov", ".wmv", ".webm", ".mpg", ".mpeg",
}

// Application is an configuration instance of the application
type Application struct {
	types.Provider
	*http.ServeMux
	cfg      types.Config
	scrapers []types.Scraper
}

// New returns a new application from the cli context
func New(cfg types.Config) *Application {
	app := &Application{
		Provider: cfg.Providers()[0],
		cfg:      cfg,
		ServeMux: http.NewServeMux(),
		scrapers: cfg.Scrapers(),
	}

	api := api.New(app)
	app.ServeMux.Handle("/api/", http.StripPrefix("/api", api))

	app.ServeMux.Handle("/", app.webAppHandler())

	return app
}

// NewFromDefault construct an application using the default config
func NewFromDefault() types.App {
	return New(cfg.Default)
}

// Config returns the configuration for the application
func (a *Application) Config() types.Config {
	return a.cfg
}

func (a *Application) webAppHandler() http.Handler {
	webfs, err := fs.New()

	if err != nil {
		log.Fatal(err)
	}

	const index = "/index.html"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := webfs.Open(r.URL.Path)
		if len(path.Ext(r.URL.Path)) == 0 || err != nil {
			if f, err = webfs.Open(index); err != nil {
				http.Error(w, "404: not found", http.StatusNotFound)
				return
			}
		}
		http.ServeContent(w, r, r.URL.Path, time.Unix(0, 0), f)
	})
}

// Scrapers returns the list of supported scrapers
func (a *Application) Scrapers() []types.Scraper {
	return a.scrapers
}

// FindMedia searches for media files
func (a *Application) FindMedia(roots ...string) (types.LocalMediaList, error) {
	medialist := make([]types.LocalMedia, 0)

	for _, root := range roots {
		if _, err := os.Stat(root); os.IsNotExist(err) {
			return nil, err
		}

		err := filepath.Walk(root, func(filepath string, f os.FileInfo, err error) error {
			if f == nil {
				return errors.New("invalid file path")
			}
			if f.IsDir() {
				return nil
			}
			if strings.HasPrefix(f.Name(), ".") {
				return nil
			}
			med, err := media.NewLocalFile(filepath)
			if err != nil {
				return nil
			}
			if media.IsSample(med) {
				return nil
			}
			medialist = append(medialist, med)
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return list.NewLocalMedia(medialist...), nil
}
