package payload_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/tombell/memoir/internal/api/payload"
	"github.com/tombell/memoir/internal/controllers/tracklistscontroller"
)

func TestWriteNoContentOverHTTP(t *testing.T) {
	results := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writer := &observedWriter{ResponseWriter: w}
		err := payload.Write(writer, &tracklistscontroller.DeleteResponse{})
		if writer.writes != 0 {
			err = fmt.Errorf("204 attempted %d body writes, Write returned %v", writer.writes, err)
		}
		results <- err
	}))
	defer server.Close()

	request, err := http.NewRequest(http.MethodDelete, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusNoContent || len(body) != 0 {
		t.Fatalf("response = %d %q, want 204 with no body", response.StatusCode, body)
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "" {
		t.Errorf("204 Content-Type = %q, want absent", contentType)
	}
	if err := <-results; err != nil {
		t.Fatal(err)
	}
}

func TestWriteNoContentSkipsJSONEncoding(t *testing.T) {
	writer := &observedWriter{ResponseWriter: httptest.NewRecorder()}
	if err := payload.Write(writer, &unencodableNoContent{}); err != nil {
		t.Fatal(err)
	}
	if writer.writes != 0 || !reflect.DeepEqual(writer.statuses, []int{http.StatusNoContent}) {
		t.Fatalf("writes = %d, statuses = %v", writer.writes, writer.statuses)
	}
}

func TestWriteEncodingFailureLeavesResponseUncommitted(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := &observedWriter{ResponseWriter: recorder}
	output := &struct {
		Location string   `header:"Location" json:"-"`
		Data     chan int `json:"data"`
	}{Location: "/created", Data: make(chan int)}

	err := payload.Write(writer, output)
	if err == nil {
		t.Fatal("expected an encoding error")
	}
	if len(writer.statuses) != 0 || writer.writes != 0 || len(recorder.Header()) != 0 {
		t.Fatalf("encoding failure changed response: statuses=%v writes=%d headers=%v",
			writer.statuses, writer.writes, recorder.Header())
	}

	payload.WriteError(discardLogger(), writer, err)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	assertErrorEnvelope(t, recorder.Body.Bytes())
	if recorder.Header().Get("Location") != "" {
		t.Fatal("error response retained the success Location header")
	}
}

func TestWriteJSONAndHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "text/plain")
	output := &struct {
		Location string `header:"Location" json:"-"`
		Data     string `json:"data"`
	}{Location: "/created", Data: "mix"}
	if err := payload.Write(recorder, output); err != nil {
		t.Fatal(err)
	}
	response := recorder.Result()
	defer response.Body.Close()
	if response.Header.Get("Content-Type") != "application/json" || response.Header.Get("Location") != "/created" {
		t.Fatalf("headers = %v", response.Header)
	}
	if got := recorder.Body.String(); got != "{\"data\":\"mix\"}\n" {
		t.Fatalf("body = %q", got)
	}
}

func TestWriteJSONValues(t *testing.T) {
	for _, value := range []any{
		map[string]string{"data": "mix"}, []string{"mix"}, "mix", nil,
		struct{ Data string }{Data: "mix"}, (*struct{ Data string })(nil),
	} {
		t.Run(fmt.Sprintf("%T", value), func(t *testing.T) {
			recorder := httptest.NewRecorder()
			if err := payload.Write(recorder, value); err != nil {
				t.Fatal(err)
			}
			if !json.Valid(recorder.Body.Bytes()) {
				t.Fatalf("invalid JSON: %q", recorder.Body.String())
			}
		})
	}
}

func TestWriteErrorDoesNotRetryFailedResponse(t *testing.T) {
	for _, failure := range []struct {
		name string
		err  error
	}{
		{name: "writer error", err: io.ErrClosedPipe},
		{name: "short write", err: nil},
	} {
		t.Run(failure.name, func(t *testing.T) {
			writer := &failingWriter{err: failure.err}
			err := payload.Write(writer, &struct{ Data string }{Data: "mix"})
			if err == nil {
				t.Fatal("expected write failure")
			}
			want := failure.err
			if want == nil {
				want = io.ErrShortWrite
			}
			if !errors.Is(err, want) {
				t.Fatalf("error = %v, want wrapped %v", err, want)
			}
			payload.WriteError(discardLogger(), writer, fmt.Errorf("wrapped: %w", err))
			if writer.writes != 1 || !reflect.DeepEqual(writer.statuses, []int{http.StatusOK}) {
				t.Fatalf("retried failed response: writes=%d statuses=%v", writer.writes, writer.statuses)
			}
		})
	}
}

func TestWriteErrorDoesNotRetryFailedErrorResponse(t *testing.T) {
	writer := &failingWriter{err: io.ErrClosedPipe}
	payload.WriteError(discardLogger(), writer, errors.New("database failure"))
	if writer.writes != 1 || !reflect.DeepEqual(writer.statuses, []int{http.StatusInternalServerError}) {
		t.Fatalf("retried failed error response: writes=%d statuses=%v", writer.writes, writer.statuses)
	}
}

func assertErrorEnvelope(t *testing.T, body []byte) {
	t.Helper()
	var envelope map[string]map[string][]string
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("invalid error envelope %q: %v", body, err)
	}
	want := map[string]map[string][]string{"errors": {"message": {"something went wrong"}}}
	if !reflect.DeepEqual(envelope, want) {
		t.Fatalf("envelope = %v, want %v", envelope, want)
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type unencodableNoContent struct{}

func (*unencodableNoContent) StatusCode() int {
	return http.StatusNoContent
}

func (*unencodableNoContent) MarshalJSON() ([]byte, error) {
	return nil, errors.New("204 must not invoke MarshalJSON")
}

type observedWriter struct {
	http.ResponseWriter
	statuses []int
	writes   int
}

func (w *observedWriter) WriteHeader(status int) {
	w.statuses = append(w.statuses, status)
	w.ResponseWriter.WriteHeader(status)
}

func (w *observedWriter) Write(body []byte) (int, error) {
	w.writes++
	return w.ResponseWriter.Write(body)
}

type failingWriter struct {
	header   http.Header
	statuses []int
	writes   int
	err      error
	body     bytes.Buffer
}

func (w *failingWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *failingWriter) WriteHeader(status int) {
	w.statuses = append(w.statuses, status)
}

func (w *failingWriter) Write(body []byte) (int, error) {
	w.writes++
	n := len(body) / 2
	w.body.Write(body[:n])
	return n, w.err
}
