package api

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tombell/memoir/internal/stores/trackliststore"
)

const mixBody = `{"name":"Shared mix title","date":"2026-10-09T12:00:00Z","url":"https://example.test/mix","artwork":"art.png","tracks":[["Track","Artist","128","8A","House"],["Track","Artist","128","8A","House"]]}`
const updateBody = `{"name":"Changed title","date":"2026-10-09T12:00:00Z","url":"https://example.test/changed"}`

func TestPublicTracklistsAndOwnerOnlyWrites(t *testing.T) {
	f := newFixture(t)
	verify := f.registerLogin(t, "owner@example.test")
	if w := f.request("POST", "/tracklists", mixBody); w.Code != 403 {
		t.Fatal("unverified user could create tracklist")
	}
	if w := f.request("POST", "/auth/verify-email", `{"token":"`+verify+`"}`); w.Code != 204 {
		t.Fatal("verify failed")
	}
	ownerCookies := append([]*http.Cookie(nil), f.cookies...)
	w := f.request("POST", "/tracklists", mixBody)
	if w.Code != 201 {
		t.Fatalf("create: %d %s", w.Code, w.Body)
	}
	var created struct {
		Data trackliststore.Tracklist `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	list := created.Data
	if list.Owner.ID == "" || list.TrackCount != 2 {
		t.Fatal("create lacks owner or tracks")
	}
	if w := f.request("POST", "/tracklists", mixBody); w.Code != 422 {
		t.Fatal("same-owner duplicate name accepted")
	}
	f.cookies = append([]*http.Cookie(nil), ownerCookies[:1]...)
	verifyOther := f.registerLogin(t, "other@example.test")
	if w := f.request("POST", "/auth/verify-email", `{"token":"`+verifyOther+`"}`); w.Code != 204 {
		t.Fatal("other verification failed")
	}
	if w := f.request("POST", "/tracklists", mixBody); w.Code != 201 {
		t.Fatalf("different-owner same name: %d %s", w.Code, w.Body)
	}
	for _, method := range []string{"PATCH", "DELETE"} {
		body := updateBody
		if method == "DELETE" {
			body = ""
		}
		if w := f.request(method, "/tracklists/"+list.ID, body); w.Code != 404 {
			t.Fatalf("other user %s: %d %s", method, w.Code, w.Body)
		}
	}
	var remaining int
	if err := f.store.QueryRow(context.Background(), "SELECT count(*) FROM tracklist_tracks WHERE tracklist_id = $1", list.ID).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 2 {
		t.Fatal("unauthorized delete removed track associations")
	}
	forged := mixBody[:len(mixBody)-1] + `,"owner_id":"` + list.Owner.ID + `"}`
	if w := f.request("POST", "/tracklists", forged); w.Code != 400 {
		t.Fatalf("forged owner not rejected: %d", w.Code)
	}
	f.cookies = nil
	for _, path := range []string{"/tracklists", "/tracklists/" + list.ID, "/tracks/" + list.Tracks[0].ID, "/tracks/search?q=Track", "/tracks/mostplayed"} {
		w := f.request("GET", path, "")
		if w.Code != 200 {
			t.Fatalf("public %s: %d %s", path, w.Code, w.Body)
		}
		if bytes.Contains(w.Body.Bytes(), []byte("owner@example.test")) || bytes.Contains(w.Body.Bytes(), []byte("password_hash")) {
			t.Fatal("public response leaks account details")
		}
	}
	w = f.request("GET", "/tracklists?user_id="+list.Owner.ID+"&track_id="+list.Tracks[0].ID, "")
	var index struct {
		Data []trackliststore.Tracklist `json:"data"`
		Meta struct {
			TotalPages float64 `json:"total_pages"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &index); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || len(index.Data) != 1 || index.Data[0].Owner.ID != list.Owner.ID || index.Meta.TotalPages != 1 {
		t.Fatalf("combined filters mismatch: %d %s", w.Code, w.Body)
	}
	f.cookies = ownerCookies
	if w := f.request("PATCH", "/tracklists/"+list.ID, updateBody); w.Code != 200 {
		t.Fatalf("owner update: %d %s", w.Code, w.Body)
	}
	w = f.request("DELETE", "/tracklists/"+list.ID, "")
	if w.Code != 204 || w.Body.Len() != 0 {
		t.Fatalf("owner delete: %d %s", w.Code, w.Body)
	}
	if w := f.request("GET", "/tracklists/"+list.ID, ""); w.Code != 404 {
		t.Fatal("deleted mix still exists")
	}
	if w := f.request("GET", "/tracks/"+list.Tracks[0].ID, ""); w.Code != 200 {
		t.Fatal("deleting mix removed shared track")
	}
}

func TestSharedTokenCannotWrite(t *testing.T) {
	f := newFixture(t)
	r := httptest.NewRequest("POST", "/tracklists", bytes.NewBufferString(mixBody))
	r.Header.Set("API-Token", "legacy-token")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", f.cfg.Auth.Origin)
	r.Header.Set("X-CSRF-Token", f.csrf)
	r.AddCookie(f.cookies[0])
	w := httptest.NewRecorder()
	f.server.router.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("shared token bypassed session: %d", w.Code)
	}
}

func TestArtworkAndTrackInputRejection(t *testing.T) {
	f := newFixture(t)
	verify := f.registerLogin(t, "owner@example.test")
	if w := f.request("POST", "/auth/verify-email", `{"token":"`+verify+`"}`); w.Code != 204 {
		t.Fatal("verify failed")
	}
	if w := f.request("POST", "/artwork", `{}`); w.Code != 400 {
		t.Fatalf("missing artwork: %d", w.Code)
	}
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	file, err := form.CreateFormFile("artwork", "picture.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("<html>not an image</html>")); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "/artwork", &body)
	r.Header.Set("Content-Type", form.FormDataContentType())
	r.Header.Set("Origin", f.cfg.Auth.Origin)
	r.Header.Set("X-CSRF-Token", f.csrf)
	for _, cookie := range f.cookies {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	f.server.router.ServeHTTP(w, r)
	if w.Code != 415 {
		t.Fatalf("non-image artwork: %d %s", w.Code, w.Body)
	}
	body.Reset()
	form = multipart.NewWriter(&body)
	file, err = form.CreateFormFile("artwork", "large.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(bytes.Repeat([]byte("x"), (8<<20)+1)); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	r = httptest.NewRequest("POST", "/artwork", &body)
	r.Header.Set("Content-Type", form.FormDataContentType())
	r.Header.Set("Origin", f.cfg.Auth.Origin)
	r.Header.Set("X-CSRF-Token", f.csrf)
	for _, cookie := range f.cookies {
		r.AddCookie(cookie)
	}
	w = httptest.NewRecorder()
	f.server.router.ServeHTTP(w, r)
	if w.Code != 413 {
		t.Fatalf("oversized artwork: %d %s", w.Code, w.Body)
	}
	bad := `{"name":"Bad mix","date":"2026-10-09T12:00:00Z","url":"https://example.test","artwork":"art.png","tracks":[["too short"]]}`
	if w := f.request("POST", "/tracklists", bad); w.Code != 422 {
		t.Fatalf("malformed track row: %d %s", w.Code, w.Body)
	}
}
