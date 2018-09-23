package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/imagespy/api/store"
	"github.com/imagespy/api/web"
)

func main() {
	s, err := store.NewGormStore("root:root@tcp(127.0.0.1:33306)/imagespy?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	if os.Getenv("MIGRATE") == "1" {
		s.Migrate()
	}

	// img, err := s.FindImageWithTagsByTag("index.docker.io/prom/prometheus", "v2.3.2")
	// if err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		fmt.Printf("%#v\n", err)
	// 	}
	// 	log.Fatal(err)
	// }
	//
	// fmt.Printf("%+v\n", img)
	// for _, tag := range img.Tags {
	// 	fmt.Printf("%+v\n", tag)
	// }
	//
	// for _, platform := range img.Platforms {
	// 	fmt.Printf("%+v\n", platform)
	// }
	// log.Println("Requesting image from registry...")
	// regImg, err := registry.NewImage("prom/prometheus:v2.3.2", false)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// vp := versionparser.FindForVersion("v2.3.2")
	//
	// log.Println("Creating image in DB...")
	// image, tag, err := s.CreateImageFromRegistryImage(vp.Distinction(), regImg)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// fmt.Printf("%+v\n", image)
	// fmt.Printf("%+v\n", tag)
	// fmt.Printf("%+v\n", image.Platforms[0])
	// for _, l := range image.Platforms[0].Layers {
	// 	fmt.Println(l.Digest)
	// }

	// fmt.Println("")
	// fmt.Println("--------------------------------------------------------")
	// fmt.Println("")
	//
	// image2, err := s.FindImageWithTagsByTag("index.docker.io/prom/prometheus", "v2.3.2")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// fmt.Printf("%#v\n", image2)
	// for _, tag2 := range image2.Tags {
	// 	fmt.Printf("%#v\n", tag2)
	// }
	//
	// for _, platform2 := range image2.Platforms {
	// 	fmt.Printf("%#v\n", platform2)
	// 	for _, layer2 := range platform2.Layers {
	// 		fmt.Printf("  %#v\n", layer2)
	// 	}
	// }

	handler := web.Init(s)
	log.Fatal(http.ListenAndServe(":3001", handler))
}
