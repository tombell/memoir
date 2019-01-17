package parser

import (
	"bytes"
	"encoding/csv"
	"io"
	"io/ioutil"
)

// ParseSeratoExport reads and parses the exported CSV file from Serato, and
// returns the parsed records.
func ParseSeratoExport(filepath string) ([][]string, error) {
	in, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(bytes.NewReader(in))

	var records [][]string

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records[1:], nil
}
