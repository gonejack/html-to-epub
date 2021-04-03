# html-to-epub
Command line tool for converting html to epub.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gonejack/html-to-epub)
![Build](https://github.com/gonejack/html-to-epub/actions/workflows/go.yml/badge.svg)
[![GitHub license](https://img.shields.io/github/license/gonejack/html-to-epub.svg?color=blue)](LICENSE)

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
      --cover string    set epub cover image
      --title string    set epub title (default "HTML")
      --author string   set epub author (default "HTML to Epub")
  -v, --verbose         verbose
  -h, --help            help for html-to-epub
```

### Preview
![](preview.png)
