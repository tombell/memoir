package templates

import "text/template"

func WithData(values ...any) (map[string]any, error) {
	return nil, nil
}

var funcs = template.FuncMap{
	"WithData": WithData,
}
