package main

import (
	_ "github.com/glebarez/go-sqlite"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/tombell/migrate/cmd/migrate/commands"
)

func main() {
	commands.Execute()
}
