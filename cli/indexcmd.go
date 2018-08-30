package main

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"

	. "github.com/wgoodall01/wikipath/wp"
)

// IndexCmd is the command to start building a `*.wpindex` file from
// a wiki archive.
var IndexCmd = cli.Command{
	Name:  "index",
	Usage: "Build an intermediate index of articles.",
	Flags: []cli.Flag{WpFlags.WpindexPath, WpFlags.WikiArchivePath, WpFlags.WikiIndexPath},
	Action: func(c *cli.Context) error {
		// Open the archive and index
		archivePath := c.String("wiki-archive")
		archiveFile, fileErr := os.Open(archivePath)
		if fileErr != nil {
			return NewFileError("Could not open wiki archive '%s'", archivePath)
		}

		indexPath := c.String("wiki-index")
		indexFile, indexErr := os.Open(indexPath)
		if indexErr != nil {
			return NewFileError("Could not open wiki index '%s'", indexPath)
		}

		outPath := c.String("wpindex")
		outFile, outErr := os.Create(outPath)
		if outErr != nil {
			return NewFileError("Could not open output file '%s'", outPath)
		}

		tStart := time.Now()

		// Set up wpindex writer, channel for articles
		writer := NewWpindexWriter(outFile)
		articles := make(chan *StrippedArticle, 512)
		ec := NewErrorContext()
		ec.Start()
		go func() {
			n := 0
			rate := NewRateMeasure(1)
			for sa := range articles {
				n++
				rate.Count(1)
				if n%500 == 0 {
					PrintTicker("Saving wpindex...   ", fmt.Sprintf("[rate:%4.2f  id:%d  title:'%s']", rate.Average(), sa.ID, sa.Title))
				}
				writer.WriteArticle(sa) // write article to *.wpindex
			}
			rate.Stop()

			closeErr := writer.Close()
			if closeErr != nil {
				ec.Cancel(closeErr)
			}
			ec.Done()
		}()

		loadErr := LoadWikiCompressed(indexFile, archiveFile, func(a *Article) bool {
			sa := NewStrippedArticle(a)
			articles <- sa
			return true
		})

		if loadErr != nil {
			return NewInternalError("failed to parse wiki archive")
		}

		close(articles)

		writerErr := ec.Wait()
		if writerErr != nil {
			return NewInternalError("failed to write to *.wpindex file: %v", writerErr.Error())
		}

		dLoad := time.Since(tStart).Seconds()
		PrintTicker("Saving wpindex...   ", fmt.Sprintf("[done in %4.2fs]", dLoad))
		fmt.Println()

		return nil
	},
}
