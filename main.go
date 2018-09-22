package main

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/imagespy/api/store"
	"github.com/jinzhu/gorm"
)

func main() {
	s, err := store.NewGormStore("root:root@tcp(127.0.0.1:33306)/imagespy?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()
	img, err := s.FindImageByTag("index.docker.io/prom/prometheus", "v2.3.2")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("%#v\n", err)
		}
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", img)
	for _, tag := range img.Tags {
		fmt.Printf("%+v\n", tag)
	}

	for _, platform := range img.Platforms {
		fmt.Printf("%+v\n", platform)
	}
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
}
