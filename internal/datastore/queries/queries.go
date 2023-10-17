package queries

import (
	"embed"
)

//go:embed */*.sql
var files embed.FS

func query(filename string) string {
	data, _ := files.ReadFile(filename)
	return string(data)
}
