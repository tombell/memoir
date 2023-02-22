package datastore

import (
	"fmt"
	"strings"
)

func formatQuery(query string) string {
	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "( ", "(")
	query = strings.ReplaceAll(query, " )", ")")
	query = strings.Trim(query, " ")

	return query
}

func formatQueryArgs(query string, args []any) []any {
	formatted := make([]any, 0, (len(args)*2)+2)
	formatted = append(formatted, "sql")
	formatted = append(formatted, formatQuery(query))

	for i, arg := range args {
		formatted = append(formatted, fmt.Sprintf("$%d", i+1))
		formatted = append(formatted, fmt.Sprint(arg))
	}

	return formatted
}
