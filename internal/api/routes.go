package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/tombell/middle"
	"github.com/tombell/middle/ware"

	"github.com/tombell/memoir/internal/api/middleware"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/controllers/artworkcontroller"
	"github.com/tombell/memoir/internal/controllers/tracklistscontroller"
	"github.com/tombell/memoir/internal/controllers/trackscontroller"

	"github.com/tombell/memoir/internal/stores/artworkstore"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

func routes(
	logger *slog.Logger,
	router *http.ServeMux,
	config *config.Config,
	tracklistStore *trackliststore.Store,
	trackStore *trackstore.Store,
	artworkStore *artworkstore.Store,
) {
	base := middle.Use(
		ware.Logger(logger),
		ware.RequestID(uuid.NewString),
		ware.RequestLogging(),
		ware.CORS(ware.CORSOptions{
			AllowedMethods: []string{"GET", "POST", "PATCH"},
			AllowedHeaders: []string{"API-Token", "Content-Type"},
		}),
	)

	api := middle.Use(
		base,
		ware.Recovery(),
	)

	authorized := middle.Use(
		base,
		middleware.Authorize(config.API.Token),
		ware.Recovery(),
	)

	router.Handle("GET /tracklists", api(rw(tracklistscontroller.Index(tracklistStore))))
	router.Handle("GET /tracklists/{id}", api(rw(tracklistscontroller.Show(tracklistStore))))
	router.Handle("POST /tracklists", authorized(rw(tracklistscontroller.Create(tracklistStore))))
	router.Handle("PATCH /tracklists/{id}", authorized(rw(tracklistscontroller.Update(tracklistStore))))

	router.Handle("GET /tracks/{id}", api(rw(trackscontroller.Show(trackStore))))
	router.Handle("GET /tracks/{id}/tracklists", api(rw(tracklistscontroller.ByTrack(trackStore, tracklistStore))))
	router.Handle("GET /tracks/mostplayed", api(w(trackscontroller.MostPlayed(trackStore))))
	router.Handle("GET /tracks/search", api(rw(trackscontroller.Search(trackStore))))

	router.Handle("POST /artwork", authorized(rw(artworkcontroller.Create(artworkStore))))

	router.Handle("OPTIONS /{path...}", api(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	router.Handle("/{path...}", api(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})))
}
