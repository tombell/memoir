package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/tombell/middle"
	"github.com/tombell/middle/ware"

	"github.com/tombell/memoir/internal/api/middleware"
	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/services/artworkservice"
	"github.com/tombell/memoir/internal/services/tracklistservice"
	"github.com/tombell/memoir/internal/services/trackservice"
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
	api := middle.Use(
		ware.Recovery(),
		ware.CORS(ware.CORSOptions{
			AllowedMethods: []string{"GET", "POST", "PATCH"},
			AllowedHeaders: []string{"API-Token"},
		}),
		ware.RequestLogging(),
		ware.RequestID(uuid.NewString),
		ware.Logger(logger),
	)

	authorized := middle.Use(
		middleware.Authorize(config.API.Token),
		api,
	)

	router.Handle("GET /tracklists", api(rw(tracklistservice.Index(tracklistStore))))
	router.Handle("GET /tracklists/{id}", api(rw(tracklistservice.Show(tracklistStore))))
	router.Handle("POST /tracklists", authorized(rw(tracklistservice.Create(tracklistStore))))
	router.Handle("PATCH /tracklists/{id}", authorized(rw(tracklistservice.Update(tracklistStore))))

	router.Handle("GET /tracks/{id}", api(rw(trackservice.Show(trackStore))))
	router.Handle("GET /tracks/{id}/tracklists", api(rw(tracklistservice.ByTrack(trackStore, tracklistStore))))
	router.Handle("GET /tracks/mostplayed", api(w(trackservice.MostPlayed(trackStore))))
	router.Handle("GET /tracks/search", api(rw(trackservice.Search(trackStore))))

	router.Handle("POST /artwork", authorized(rw(artworkservice.Upload(artworkStore))))

	router.Handle("OPTIONS /{path...}", api(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
}
