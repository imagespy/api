package main

import (
	"flag"
	"net/http"

	"github.com/imagespy/api/registry"
	"github.com/imagespy/api/scrape"

	_ "github.com/go-sql-driver/mysql"
	"github.com/imagespy/api/store/gorm"
	"github.com/imagespy/api/web"
	log "github.com/sirupsen/logrus"
)

var (
	httpAddress = flag.String("http.address", ":3001", "Address to bind to")
	logLevel    = flag.String("log.level", "error", "Log Level")
)

func mustInitLogging() {
	lvl, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal(err)
	}

	log.SetLevel(lvl)
}

func main() {
	flag.Parse()
	mustInitLogging()
	s, err := gorm.New("root:root@tcp(127.0.0.1:33306)/imagespy?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	reg, err := registry.NewRegistry("", registry.Opts{})
	if err != nil {
		log.Fatal(err)
	}

	handler := web.Init(scrape.NewScraper(reg, s), s)
	log.Fatal(http.ListenAndServe(*httpAddress, handler))
}
