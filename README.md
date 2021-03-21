# html-to-epub

Command line tool for converting html to epub.

### Install
```shell
> go get github.com/gonejack/html-to-epub
```

### Usage
```shell
> html-to-epub *.html
```
```
Usage:
  html-to-epub [-o output] [--title title] [--cover cover] *.html

Flags:
  -o, --output string   output filename (default "output.epub")
      --title string    epub title (default "HTML")
      --cover string    cover image
  -v, --verbose         verbose
  -h, --help            help for html-to-epub
```
