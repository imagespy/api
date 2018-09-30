package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/imagespy/api/store"
	"github.com/imagespy/api/web"
)

var (
	httpAddress = flag.String("http.address", ":3001", "Address to bind to")
)

func main() {
	flag.Parse()
	s, err := store.NewGormStore("root:root@tcp(127.0.0.1:33306)/imagespy?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	if os.Getenv("MIGRATE") == "1" {
		s.Migrate()
	}

	handler := web.Init(s)
	log.Fatal(http.ListenAndServe(*httpAddress, handler))
}
