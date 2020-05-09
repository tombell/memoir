package jsondate

import (
	"fmt"
	"strings"
	"time"
)

const yearMonthDayFormat = "2006-01-02"

// Date is type of time.Time that marshals and unmarshals to a YYYY-MM-DD format
// date.
type Date struct {
	time.Time
}

// MarshalJSON marshals the timestamp into YYYY-MM-DD format.
func (t Date) MarshalJSON() ([]byte, error) {
	stamp := []byte(fmt.Sprintf("\"%s\"", t.Format(yearMonthDayFormat)))
	return stamp, nil
}

// UnmarshalJSON unmarshals the timestamp from YYYY-MM-DD format.
func (t *Date) UnmarshalJSON(data []byte) error {
	parsed := strings.Trim(string(data), "\"")
	date, err := time.Parse(yearMonthDayFormat, parsed)
	if err != nil {
		return err
	}
	t.Time = date
	return nil
}
