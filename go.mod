module github.com/gonejack/html-to-epub

go 1.16

require (
	github.com/PuerkitoBio/goquery v1.6.1
	github.com/bmaupin/go-epub v0.5.3
	github.com/dustin/go-humanize v1.0.0
	github.com/gabriel-vasile/mimetype v1.2.0
	github.com/schollz/progressbar/v3 v3.7.6
	github.com/spf13/cobra v1.1.3
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
)

replace github.com/bmaupin/go-epub => github.com/gonejack/go-epub v1.0.0
