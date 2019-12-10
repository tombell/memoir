package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/tombell/memoir/internal/decode"
)

const dateTimeFormat = "2006-01-02"

func parseSeratoExport(filepath string) ([][]string, error) {
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

	return records[2:], nil
}

func parseRekordboxExport(filepath string) ([][]string, error) {
	f, err := os.Open(filepath)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	var records [][]string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := decode.UTF16(scanner.Bytes())
		parts := strings.Split(line, "\t")

		if parts[0] == "#" || len(parts) <= 1 {
			continue
		}

		record := []string{
			strings.TrimSpace(parts[2]),
			strings.TrimSpace(parts[3]),
			strings.TrimSpace(parts[6]),
			strings.TrimSpace(parts[9]),
			strings.TrimSpace(parts[5]),
		}

		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return records, nil
}
