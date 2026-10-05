package controllers

import (
	"math"
	"net/http"
	"testing"

	"github.com/tombell/memoir/internal/errors"
)

func TestPagination(t *testing.T) {
	for _, tt := range []struct {
		name, page, perPage   string
		wantPage, wantPerPage int64
		invalid               string
	}{
		{name: "defaults", wantPage: 1, wantPerPage: 10},
		{name: "second page", page: "2", perPage: "3", wantPage: 2, wantPerPage: 3},
		{name: "maximum size", page: "1", perPage: "100", wantPage: 1, wantPerPage: 100},
		{name: "largest offset", page: "2147483648", perPage: "1", wantPage: 2147483648, wantPerPage: 1},
		{name: "zero page", page: "0", invalid: "page"},
		{name: "negative size", perPage: "-1", invalid: "per_page"},
		{name: "malformed page", page: "1.5", invalid: "page"},
		{name: "malformed size", perPage: "ten", invalid: "per_page"},
		{name: "excessive size", perPage: "101", invalid: "per_page"},
		{name: "int64 overflow", page: "9223372036854775808", invalid: "page"},
		{name: "multiplication overflow", page: "9223372036854775807", perPage: "100", invalid: "page"},
		{name: "int32 offset overflow", page: "2147483649", perPage: "1", invalid: "page"},
		{name: "offset size interaction", page: "214748366", perPage: "10", invalid: "page"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			page, perPage, err := Pagination(tt.page, tt.perPage)
			if tt.invalid != "" {
				var e *errors.Error
				if !errors.As(err, &e) || e.Status() != http.StatusBadRequest || len(e.Message()[tt.invalid]) == 0 {
					t.Fatalf("want 400 for %s, got %v", tt.invalid, err)
				}
				return
			}
			if err != nil || page != tt.wantPage || perPage != tt.wantPerPage {
				t.Fatalf("got (%d, %d, %v), want (%d, %d, nil)", page, perPage, err, tt.wantPage, tt.wantPerPage)
			}
		})
	}
}

func TestNewMeta(t *testing.T) {
	for _, tt := range []struct{ total, page, size, pages int64 }{
		{0, 1, 10, 0}, {10, 1, 10, 1}, {11, 2, 10, 2}, {11, 3, 10, 2},
		{math.MaxInt64, 1, 100, math.MaxInt64/100 + 1},
	} {
		got := NewMeta(tt.total, tt.page, tt.size)
		if got.CurrentPage != tt.page || got.PerPage != tt.size || got.Total != tt.total || got.TotalPages != tt.pages {
			t.Fatalf("NewMeta(%d, %d, %d) = %+v", tt.total, tt.page, tt.size, got)
		}
	}
}
