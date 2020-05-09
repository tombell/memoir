package jsondate_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/tombell/memoir/internal/jsondate"
)

type response struct {
	Date jsondate.Date
}

func TestDateMarshalJSON(t *testing.T) {
	now := time.Now()

	resp := &response{Date: jsondate.Date{now}}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("could not json marshal the response: %v", err)
	}

	expected := fmt.Sprintf(`{"Date":"%s"}`, now.Format("2006-01-02"))
	actual := string(data)

	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}

func TestDateUnmarshalJSON(t *testing.T) {
	data := []byte(`{"Date":"2022-01-01"}`)

	var resp response
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Errorf("expected json unmarshal to not return an error, got %v", err)
	}

	year := resp.Date.Year()
	month := resp.Date.Month()
	day := resp.Date.Day()

	fmt.Println(resp.Date)

	if year != 2022 {
		t.Errorf("expect year to be 2020, got %v", year)
	}

	if month != 1 {
		t.Errorf("expect month to be 1, got %v", year)
	}

	if day != 1 {
		t.Errorf("expect day to be 1, got %v", year)
	}
}
func TestDateUnmarshalJSONInvalid(t *testing.T) {
	data := []byte(`{"Date":"0000-40-20"}`)

	var resp response
	if err := json.Unmarshal(data, &resp); err == nil {
		t.Error("expected json unmarshal to return an error, got nil")
	}
}
