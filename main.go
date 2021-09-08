package main

import (
	_ "embed"
	"log"

	"github.com/gonejack/html-to-epub/cmd"
)

//go:embed cover.png
var cover []byte

func main() {
	c := cmd.HtmlToEpub{
		DefaultCover: cover,
		ImagesDir:    "images",
	}
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}
