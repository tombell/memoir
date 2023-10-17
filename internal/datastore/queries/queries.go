package queries

import (
	"embed"
)

//go:embed */*.sql
var files embed.FS

func query(filename string) string {
	data, err := files.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(data)
}
