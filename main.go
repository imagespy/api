package main

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/imagespy/api/cmd"
)

func main() {
	cmd.Execute()
}
