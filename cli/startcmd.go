package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/urfave/cli"

	. "github.com/wgoodall01/wikipath/wp"
)

var StartCmd = cli.Command{
	Name:  "start",
	Usage: "Start interactive mode",
	Flags: []cli.Flag{WpFlags.WpindexPath},
	Action: func(c *cli.Context) error {
		// Open the index
		indexPath := c.String("wpindex")
		indexFile, indexErr := os.Open(indexPath)
		if indexErr != nil {
			return NewFileError("Could not open index file '%s'", indexPath)
		}

		// Create WpindexReader
		reader, readerErr := NewWpindexReader(indexFile)
		if readerErr != nil {
			return NewFileError("Could not understand index.")
		}

		PrintTicker("Loading wpindex...  ", "")

		// Load all the articles.
		tLoad := time.Now()
		ind := NewIndex()

		articles := make(chan *StrippedArticle, 512)
		ec := NewErrorContext()
		ec.Start()
		go func() {
			n := 0
			rate := NewRateMeasure(0.5)
			for sa := range articles {
				n++
				rate.Count(1)
				if n%500 == 0 {
					PrintTicker("Loading wpindex...  ", fmt.Sprintf("[rate:%4.2f  article:%d  title: %s]", rate.Average(), sa.Id, sa.Title))
				}
				ind.AddArticle(sa)
			}
			rate.Stop()
			ec.Done()
		}()

		var wpindexErr error = nil
		for {
			var sa *StrippedArticle
			sa, wpindexErr = reader.ReadArticle()
			if wpindexErr != nil {
				break
			}
			articles <- sa
		}
		close(articles)

		ec.Wait()

		if wpindexErr != nil && wpindexErr != EOF {
			return cli.NewExitError("Error loading .wpindex file: "+wpindexErr.Error(), 2)
		}

		closeErr := reader.Close()
		if closeErr != nil {
			return cli.NewExitError("Couldn't close .wpindex reader", 3)
		}

		dLoad := time.Since(tLoad).Seconds()
		PrintTicker("Loading wpindex...  ", fmt.Sprintf("[done in %4.2f s]", dLoad))
		fmt.Println()

		// Index all the articles.
		fmt.Print("Making index...     ")
		tBuild := time.Now()
		ind.Build()
		dBuild := time.Since(tBuild).Seconds()
		fmt.Printf("[done in %4.2fs]\n", dBuild)

		// Run a GC
		fmt.Print("Running GC...       ")
		runtime.GC()
		fmt.Printf("[done]\n")

		// Find a path.
	InputLoop:
		for true {
			fmt.Print("\n\n")
			names := [2]string{}
			items := [2]*IndexItem{}

			names[0] = Prompt("First Article")
			names[1] = Prompt("Second Article")

			for i, _ := range names {
				items[i] = ind.Get(names[i])
				if items[i] == nil {
					fmt.Printf("Error: Can't find article '%s'", names[i])
					continue InputLoop
				}
			}

			fmt.Println()
			fmt.Printf("%20s  -> %8d\n", items[0].Title, len(items[0].Forward))
			fmt.Printf("%20s  <- %8d\n", items[1].Title, len(items[1].Reverse))

			tSearch := time.Now()
			fmt.Printf("\nSearching for path... ")
			nSteps := 10
			path := ind.FindPath(items[0], items[1], nSteps)
			dSearch := time.Since(tSearch).Seconds()
			fmt.Printf("[done in %4.2f]\n", dSearch)

			if path == nil {
				fmt.Printf("No paths found in %d steps.", nSteps)
			} else {
				fmt.Println("Path: ", path)
			}

			fmt.Println()
			ind.Reset()
		}

		return nil
	},
}
