package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/tombell/middle/ware"

	"github.com/tombell/memoir/internal/api/payload"
)

func TestRecoveryUsesJSONErrorEnvelope(t *testing.T) {
	handler := recoveryHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("private database details")
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusInternalServerError || recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("response = %d %v", recorder.Code, recorder.Header())
	}
	var envelope map[string]map[string][]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	want := map[string]map[string][]string{"errors": {"message": {"something went wrong"}}}
	if !reflect.DeepEqual(envelope, want) {
		t.Fatalf("envelope = %v, want %v", envelope, want)
	}
}

func TestRecoveryDoesNotRewriteCommittedResponses(t *testing.T) {
	for _, test := range []struct {
		name    string
		commit  func(http.ResponseWriter)
		status  int
		body    string
		flushed bool
	}{
		{
			name: "explicit headers",
			commit: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusAccepted)
			},
			status: http.StatusAccepted,
		},
		{
			name: "implicit headers",
			commit: func(w http.ResponseWriter) {
				_, _ = w.Write([]byte("accepted"))
			},
			status: http.StatusOK,
			body:   "accepted",
		},
		{
			name: "flush",
			commit: func(w http.ResponseWriter) {
				if err := http.NewResponseController(w).Flush(); err != nil {
					panic(err)
				}
			},
			status:  http.StatusOK,
			flushed: true,
		},
		{
			name: "error after successful JSON",
			commit: func(w http.ResponseWriter) {
				_ = payload.Write(w, map[string]string{"data": "mix"})
				payload.WriteError(testLogger(), w, errors.New("late error"))
			},
			status: http.StatusOK,
			body:   "{\"data\":\"mix\"}\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			writer := &countingWriter{ResponseWriter: recorder}
			handler := recoveryHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				test.commit(w)
				panic("late panic")
			}))
			handler.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, "/", nil))
			if !reflect.DeepEqual(writer.statuses, []int{test.status}) {
				t.Fatalf("headers rewritten: %v", writer.statuses)
			}
			if recorder.Body.String() != test.body || recorder.Flushed != test.flushed {
				t.Fatalf("body=%q flushed=%v", recorder.Body.String(), recorder.Flushed)
			}
		})
	}
}

func TestRecoveryDoesNotRetryWriterFailure(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := &countingWriter{ResponseWriter: recorder, writeErr: io.ErrClosedPipe}
	handler := recoveryHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("mix"))
		panic("writer failed")
	}))
	handler.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, "/", nil))
	if writer.writes != 1 || !reflect.DeepEqual(writer.statuses, []int{http.StatusOK}) {
		t.Fatalf("writer retried: writes=%d statuses=%v", writer.writes, writer.statuses)
	}
}

func TestRecoveryPreservesAbortHandler(t *testing.T) {
	defer func() {
		if value := recover(); value != http.ErrAbortHandler {
			t.Fatalf("panic = %v, want http.ErrAbortHandler", value)
		}
	}()
	handler := recoveryHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

func recoveryHandler(next http.Handler) http.Handler {
	return ware.Logger(testLogger())(Recovery()(next))
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type countingWriter struct {
	http.ResponseWriter
	statuses []int
	writes   int
	writeErr error
}

func (w *countingWriter) WriteHeader(status int) {
	w.statuses = append(w.statuses, status)
	w.ResponseWriter.WriteHeader(status)
}

func (w *countingWriter) Write(body []byte) (int, error) {
	w.writes++
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	return w.ResponseWriter.Write(body)
}

func (w *countingWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}
