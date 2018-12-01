package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/imagespy/api/cmd"
)

func main() {
	cmd.Execute()
}
