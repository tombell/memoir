package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/tombell/middle/ware"

	"github.com/tombell/memoir/internal/config"
	"github.com/tombell/memoir/internal/controllers/tracklistscontroller"
	"github.com/tombell/memoir/internal/stores/trackliststore"
	"github.com/tombell/memoir/internal/stores/trackstore"
)

func TestRoutesAuthorizationAndPanicResponsesOverHTTP(t *testing.T) {
	configuration := &config.Config{}
	configuration.API.Token = "test-token"
	api := New(discardLogger(), configuration, nil, nil, nil)
	server := httptest.NewServer(api.router)
	defer server.Close()

	for _, test := range []struct {
		name    string
		method  string
		path    string
		token   string
		status  int
		message string
	}{
		{"missing token", http.MethodPost, "/tracklists", "", http.StatusUnauthorized, "unauthorized"},
		{"invalid token", http.MethodPost, "/tracklists", "wrong", http.StatusForbidden, "forbidden"},
		{"public panic", http.MethodGet, "/tracklists/4d95cc38-92be-426d-84e9-e1ea6bb5bb91", "", http.StatusInternalServerError, "something went wrong"},
		{"authorized panic", http.MethodDelete, "/tracklists/4d95cc38-92be-426d-84e9-e1ea6bb5bb91", "test-token", http.StatusInternalServerError, "something went wrong"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(test.method, server.URL+test.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			if test.token != "" {
				request.Header.Set("API-Token", test.token)
			}
			response, err := server.Client().Do(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != test.status || response.Header.Get("Content-Type") != "application/json" {
				t.Fatalf("response = %d %v", response.StatusCode, response.Header)
			}
			var envelope map[string]map[string][]string
			if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
				t.Fatal(err)
			}
			want := map[string]map[string][]string{"errors": {"message": {test.message}}}
			if !reflect.DeepEqual(envelope, want) {
				t.Fatalf("envelope = %v, want %v", envelope, want)
			}
		})
	}
}

func TestCreateResponseOverHTTP(t *testing.T) {
	tracklist := &trackliststore.Tracklist{
		ID: "4d95cc38-92be-426d-84e9-e1ea6bb5bb91", Name: "Mix",
		Date:       time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		Tracks:     []*trackstore.Track{{ID: "track-id", Name: "Song", Artist: "Artist", BPM: 128, Key: "AM"}},
		TrackCount: 1,
	}
	handler := rw(func(context.Context, tracklistscontroller.CreateRequest) (*tracklistscontroller.CreateResponse, error) {
		return &tracklistscontroller.CreateResponse{Tracklist: tracklist}, nil
	})
	server := httptest.NewServer(ware.Logger(discardLogger())(handler))
	defer server.Close()
	response, err := server.Client().Post(server.URL, "application/json", bytes.NewBufferString("{}"))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated || response.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("response = %d %v, want 201 JSON", response.StatusCode, response.Header)
	}
	var output tracklistscontroller.CreateResponse
	if err := json.NewDecoder(response.Body).Decode(&output); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(output.Tracklist, tracklist) {
		t.Fatalf("data = %+v, want %+v", output.Tracklist, tracklist)
	}
}

func TestActionHandlersReplaceEncodingFailure(t *testing.T) {
	for name, handler := range encodingFailureHandlers() {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ware.Logger(discardLogger())(handler).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
			if recorder.Code != http.StatusInternalServerError || !json.Valid(recorder.Body.Bytes()) {
				t.Fatalf("response = %d %q, want 500 JSON", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestActionHandlersDoNotRetryWriterFailure(t *testing.T) {
	handlers := map[string]http.Handler{
		"read and write": rw(func(context.Context, struct{}) (*struct{ Data string }, error) {
			return &struct{ Data string }{Data: "mix"}, nil
		}),
		"write only": w(func(context.Context) (*struct{ Data string }, error) {
			return &struct{ Data string }{Data: "mix"}, nil
		}),
	}
	for name, handler := range handlers {
		t.Run(name, func(t *testing.T) {
			writer := &failedResponseWriter{}
			ware.Logger(discardLogger())(handler).ServeHTTP(writer, httptest.NewRequest(http.MethodGet, "/", nil))
			if writer.writes != 1 || !reflect.DeepEqual(writer.statuses, []int{http.StatusOK}) {
				t.Fatalf("retried response: writes=%d statuses=%v", writer.writes, writer.statuses)
			}
		})
	}
}

func encodingFailureHandlers() map[string]http.Handler {
	return map[string]http.Handler{
		"read and write": rw(func(context.Context, struct{}) (*struct{ Data chan int }, error) {
			return &struct{ Data chan int }{Data: make(chan int)}, nil
		}),
		"write only": w(func(context.Context) (*struct{ Data chan int }, error) {
			return &struct{ Data chan int }{Data: make(chan int)}, nil
		}),
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type failedResponseWriter struct {
	statuses []int
	writes   int
}

func (*failedResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (w *failedResponseWriter) WriteHeader(status int) {
	w.statuses = append(w.statuses, status)
}

func (w *failedResponseWriter) Write([]byte) (int, error) {
	w.writes++
	return 0, errors.New("writer failed")
}
