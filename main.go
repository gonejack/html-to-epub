package main

import (
	_ "embed"
	"log"

	"github.com/gonejack/html-to-epub/html2epub"
)

//go:embed cover.png
var defaultCover []byte

func main() {
	cmd := html2epub.HtmlToEpub{
		Options:      html2epub.MustParseOptions(),
		DefaultCover: defaultCover,
	}
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
