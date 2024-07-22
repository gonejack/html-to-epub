package html2epub

import (
	"path/filepath"

	"github.com/alecthomas/kong"
)

type Options struct {
	Cover   string `help:"Set epub cover image."`
	Title   string `default:"HTML" help:"Set epub title."`
	Author  string `default:"HTML to Epub" help:"Set epub author."`
	Output  string `short:"o" default:"output.epub" help:"Output filename."`
	Verbose bool   `short:"v" help:"Verbose printing."`
	About   bool   `help:"About."`
	Css		string `default:"" help:"A path to a CSS stylesheet, local or online"`
	
	ImagesDir string `hidden:"" default:"images"`

	HTML []string `arg:"" optional:""`
}

func MustParseOptions() (opts Options) {
	kong.Parse(&opts,
		kong.Name("html-to-epub"),
		kong.Description("This command line converts .html to .epub with images embed"),
		kong.UsageOnError(),
	)
	if len(opts.HTML) == 0 || opts.HTML[0] == "*.html" {
		opts.HTML, _ = filepath.Glob("*.html")
	}
	return
}
