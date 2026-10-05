package payload

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/tombell/memoir/internal/errors"
)

type jsonInput struct {
	Name   string     `json:"name"`
	Tracks [][]string `json:"tracks"`
}

func TestReadJSONMediaTypes(t *testing.T) {
	for _, contentType := range []string{
		"application/json",
		"application/json; charset=utf-8",
		`application/json; charset="utf-8"`,
		"APPLICATION/JSON; Charset=UTF-8",
	} {
		t.Run(contentType, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
				`{"name":"Mix","tracks":[["Track","Artist","128.5","Am","House"]]}`))
			r.Header.Set("Content-Type", contentType)
			input, err := Read[jsonInput](r)
			if err != nil {
				t.Fatal(err)
			}
			want := jsonInput{Name: "Mix", Tracks: [][]string{{"Track", "Artist", "128.5", "Am", "House"}}}
			if !reflect.DeepEqual(input, want) {
				t.Fatalf("input = %#v, want %#v", input, want)
			}
		})
	}

	for _, contentType := range []string{
		"", "text/plain", "application/xml", "multipart/form-data; boundary=test",
		"application/json; charset", "application/json; charset=utf-8; charset=ascii",
	} {
		t.Run("unsupported/"+contentType, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Mix"}`))
			r.Header.Set("Content-Type", contentType)
			_, err := Read[jsonInput](r)
			assertReadStatus(t, err, http.StatusUnsupportedMediaType)
		})
	}
}

func TestReadRejectsMalformedAndTrailingJSON(t *testing.T) {
	for _, body := range []string{
		"", " ", `{"name":`, `{"name":"Mix",}`, `{"name":123}`,
		`{"tracks":[["Track",123,"128","Am","House"]]}`,
		`{"name":"Mix"} {}`, `{"name":"Mix"} null`, `{"name":"Mix"} garbage`,
	} {
		t.Run(body, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			_, err := Read[jsonInput](r)
			assertReadStatus(t, err, http.StatusBadRequest)
		})
	}

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{\"name\":\"Mix\"} \n\t"))
	r.Header.Set("Content-Type", "application/json")
	if _, err := Read[jsonInput](r); err != nil {
		t.Fatalf("trailing whitespace should be accepted: %v", err)
	}
}

func TestReadJSONBodyLimit(t *testing.T) {
	for _, tc := range []struct {
		name          string
		size          int
		contentLength int64
		status        int
	}{
		{"at limit", int(MaxJSONBodyBytes), MaxJSONBodyBytes, 0},
		{"at limit without length", int(MaxJSONBodyBytes), -1, 0},
		{"over limit", int(MaxJSONBodyBytes) + 1, MaxJSONBodyBytes + 1, http.StatusRequestEntityTooLarge},
		{"over limit without length", int(MaxJSONBodyBytes) + 1, -1, http.StatusRequestEntityTooLarge},
		{"over limit with understated length", int(MaxJSONBodyBytes) + 1, 2, http.StatusRequestEntityTooLarge},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := "{}" + strings.Repeat(" ", tc.size-2)
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			r.ContentLength = tc.contentLength
			_, err := Read[jsonInput](r)
			if tc.status == 0 {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			e := assertReadStatus(t, err, tc.status)
			if !strings.Contains(e.Message()["message"][0], "1048576 bytes") {
				t.Fatalf("size error should state the limit: %v", e.Message())
			}
		})
	}
}

func TestReadRequestFields(t *testing.T) {
	type input struct {
		ID      string  `path:"id"`
		Token   string  `header:"API-Token"`
		Query   string  `query:"q"`
		Page    *string `query:"page"`
		PerPage *string `query:"per_page"`
	}
	r := httptest.NewRequest(http.MethodGet, "/?q=some+track&page=2&per_page=", nil)
	r.SetPathValue("id", "track-id")
	r.Header.Set("API-Token", "token")
	// Path/query inputs do not require a JSON body, even with this header.
	r.Header.Set("Content-Type", "application/json")
	in, err := Read[input](r)
	if err != nil {
		t.Fatal(err)
	}
	if in.ID != "track-id" || in.Token != "token" || in.Query != "some track" ||
		in.Page == nil || *in.Page != "2" || in.PerPage == nil || *in.PerPage != "" {
		t.Fatalf("unexpected decoded fields: %#v", in)
	}

	in, err = Read[input](httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil || in.Page != nil || in.PerPage != nil {
		t.Fatalf("omitted query parameters should remain nil: %#v, %v", in, err)
	}
}

func TestReadRejectsMalformedQuery(t *testing.T) {
	type input struct {
		Page *string `query:"page"`
	}
	for _, query := range []string{"page=%zz", "page=1;per_page=10"} {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.URL.RawQuery = query
		_, err := Read[input](r)
		assertReadStatus(t, err, http.StatusBadRequest)
	}
}

func TestReadMultipartIsNotLimitedAsJSON(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("artwork", "cover.jpg")
	if err != nil {
		t.Fatal(err)
	}
	contents := strings.Repeat("x", int(MaxJSONBodyBytes)+1)
	if _, err := io.WriteString(file, contents); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/artwork", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	input, err := Read[struct {
		Artwork *File `file:"artwork"`
	}](r)
	if err != nil {
		t.Fatal(err)
	}
	if r.MultipartForm != nil {
		t.Cleanup(func() { _ = r.MultipartForm.RemoveAll() })
	}
	if input.Artwork == nil {
		t.Fatal("multipart file was not decoded")
	}
	t.Cleanup(func() { _ = input.Artwork.File.Close() })
	got, err := io.ReadAll(input.Artwork.File)
	if err != nil || string(got) != contents {
		t.Fatalf("multipart contents did not survive decoding: size %d, error %v", len(got), err)
	}
}

func assertReadStatus(t *testing.T, err error, status int) *errors.Error {
	t.Helper()
	var e *errors.Error
	if !errors.As(err, &e) || e.Status() != status {
		t.Fatalf("error = %v, want HTTP status %d", err, status)
	}
	return e
}
