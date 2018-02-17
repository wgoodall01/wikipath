package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	. "github.com/wgoodall01/wikipath/wp"
)

var IndexShowCmd = cli.Command{
	Name:  "index-show",
	Usage: "Show an article's entry in the index.",
	Flags: []cli.Flag{WpFlags.WpindexPath},
	Action: func(c *cli.Context) error {
		// Open the index
		indexFile, indexErr := os.Open(c.String("wpindex"))
		if indexErr != nil {
			return NewFileError("could not open wiki index '%s'", WpFlags.WpindexPath.Value)
		}

		fmt.Printf("Searching %s...\n", c.String("wpindex"))

		// Only one article
		args := c.Args()
		if len(args) != 1 {
			return NewUsageError("Only 1 article should be passed. Got %d", len(args))
		}

		// Create WpindexReader
		reader, readerErr := NewWpindexReader(indexFile)

		if readerErr != nil {
			return NewFileError("Could not understand index.")
		}

		for {
			sa, _ := reader.ReadArticle()
			if sa == nil {
				break
			}
			if NormalizeArticleTitle(sa.Title) == NormalizeArticleTitle(c.Args()[0]) {
				fmt.Println(sa.Title)
				for _, l := range sa.Links {
					fmt.Println("  " + l)
				}
				break
			}

		}

		return nil
	},
}
