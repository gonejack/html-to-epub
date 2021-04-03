package main

import (
	"fmt"
	"log"
	"os"

	_ "embed"

	"github.com/gonejack/html-to-epub/cmd"
	"github.com/spf13/cobra"
)

var (
	//go:embed cover.png
	defaultCover []byte

	cover   *string
	title   *string
	author  *string
	output  *string
	verbose = false

	prog = &cobra.Command{
		Use:   "html-to-epub [-o output] [--title title] [--cover cover] *.html",
		Short: "Command line tool for converting html to epub.",
		Run: func(c *cobra.Command, args []string) {
			err := run(c, args)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	log.SetOutput(os.Stdout)

	prog.Flags().SortFlags = false
	prog.PersistentFlags().SortFlags = false

	output = prog.PersistentFlags().StringP(
		"output",
		"o",
		"output.epub",
		"output filename",
	)
	cover = prog.PersistentFlags().StringP(
		"cover",
		"",
		"",
		"set epub cover image",
	)
	title = prog.PersistentFlags().StringP(
		"title",
		"",
		"HTML",
		"set epub title",
	)
	author = prog.PersistentFlags().StringP(
		"author",
		"",
		"HTML to Epub",
		"set epub author",
	)
	prog.PersistentFlags().BoolVarP(
		&verbose,
		"verbose",
		"v",
		false,
		"verbose",
	)
}

func run(c *cobra.Command, args []string) error {
	_, err := os.Stat(*output)
	if !os.IsNotExist(err) {
		return fmt.Errorf("output file %s already exist", *output)
	}

	exec := cmd.HtmlToEpub{
		DefaultCover: defaultCover,

		ImagesDir: "images",

		Cover:   *cover,
		Title:   *title,
		Author:  *author,
		Verbose: verbose,
	}

	return exec.Run(args, *output)
}

func main() {
	_ = prog.Execute()
}
