package controllers

import (
	"math"
	"net/http"
	"testing"

	"github.com/tombell/memoir/internal/errors"
)

func TestPaginationDefaults(t *testing.T) {
	page, perPage, err := PaginationParams(nil, nil)
	if err != nil || page != 1 || perPage != 10 {
		t.Fatalf("pagination = %d, %d, %v, want 1, 10, nil", page, perPage, err)
	}
}

func TestPaginationBounds(t *testing.T) {
	for _, tc := range []struct {
		page, perPage string
		wantOffset    int64
	}{
		{"1", "100", 0},
		{"2", "10", 10},
		{"2147483648", "1", math.MaxInt32},
		{"214748365", "10", 2147483640},
		{"21474837", "100", 2147483600},
	} {
		t.Run(tc.page+"/"+tc.perPage, func(t *testing.T) {
			page, perPage, err := PaginationParams(&tc.page, &tc.perPage)
			if err != nil {
				t.Fatal(err)
			}
			offset := (page - 1) * perPage
			if offset != tc.wantOffset || int64(int32(offset)) != offset || int64(int32(perPage)) != perPage {
				t.Fatalf("pagination does not fit database parameters: page %d, per_page %d, offset %d", page, perPage, offset)
			}
		})
	}
}

func TestPaginationRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		page, perPage, field string
	}{
		{"", "10", "page"},
		{"0", "10", "page"},
		{"-1", "10", "page"},
		{"1.5", "10", "page"},
		{"abc", "10", "page"},
		{" 1", "10", "page"},
		{"9223372036854775808", "10", "page"},
		{"-9223372036854775808", "10", "page"},
		{"9223372036854775807", "100", "page"},
		{"2147483649", "1", "page"},
		{"214748366", "10", "page"},
		{"21474838", "100", "page"},
		{"1", "", "per_page"},
		{"1", "0", "per_page"},
		{"1", "-10", "per_page"},
		{"1", "abc", "per_page"},
		{"1", "10.5", "per_page"},
		{"1", "101", "per_page"},
		{"1", "2147483648", "per_page"},
		{"1", "9223372036854775807", "per_page"},
		{"1", "9223372036854775808", "per_page"},
	} {
		t.Run(tc.page+"/"+tc.perPage, func(t *testing.T) {
			_, _, err := PaginationParams(&tc.page, &tc.perPage)
			var e *errors.Error
			if !errors.As(err, &e) || e.Status() != http.StatusBadRequest || len(e.Message()[tc.field]) == 0 {
				t.Fatalf("error = %v, want 400 with field %q", err, tc.field)
			}
		})
	}
}

func TestParamAsIntRejectsNonpositiveValues(t *testing.T) {
	for _, value := range []string{"0", "-1", "invalid", "9223372036854775808"} {
		_, err := ParamAsInt(value, 10)
		var e *errors.Error
		if !errors.As(err, &e) || e.Status() != http.StatusBadRequest {
			t.Fatalf("ParamAsInt(%q) error = %v, want 400", value, err)
		}
	}
}
