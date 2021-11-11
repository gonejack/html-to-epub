# html-to-epub
This command line tool converts .html to .epub with images fetching.

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
Flags:
  -h, --help                     Show context-sensitive help.
  -o, --output="output.epub"     Output filename.
  -c, --cover=STRING             Set epub cover image.
      --title="HTML"             Set epub title.
      --author="HTML to Epub"    Set epub author.
  -v, --verbose                  Verbose printing.
```

### Screenshot
![](screenshot.png)
