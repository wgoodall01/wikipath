package main

import (
	"bufio"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"runtime"
	"strings"
	"time"
)

func NewFileError(msg string) *cli.ExitError {
	return cli.NewExitError(msg, 1)
}

func prompt(prompt string) string {
	in := bufio.NewReader(os.Stdin)
	fmt.Printf("%-15s: ", prompt)
	valRaw, _ := in.ReadString('\n')
	return strings.TrimSpace(valRaw)
}

func cliTicker(prompt string, tick string) {
	width := 100
	status := prompt + tick + strings.Repeat(" ", width)
	status = status[:width] + "\r"
	fmt.Print(status)
}

var indexCmd = cli.Command{
	Name:  "index",
	Usage: "Build an intermediate index of articles.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "archive, a",
			Usage: "Wiki archive *multistream.xml.bz2 file",
		},
		cli.StringFlag{
			Name:  "index, i",
			Usage: "Wiki *-multistream-index.txt.bz2 file",
		},
		cli.StringFlag{
			Name:  "out, o",
			Usage: "The output .wpindex file",
		},
	},
	Action: func(c *cli.Context) error {
		// Open the archive and index
		fmt.Print("Opening archive...  ")
		archiveFile, fileErr := os.Open(c.String("archive"))
		if fileErr != nil {
			return NewFileError("Could not open archive.")
		}
		fmt.Print("[done]\n")

		fmt.Print("Opening index...    ")
		indexFile, indexErr := os.Open(c.String("index"))
		if indexErr != nil {
			return NewFileError("Could not open index.")
		}
		fmt.Print("[done]\n")

		fmt.Print("Opening output...   ")
		outFile, outErr := os.Create(c.String("out"))
		if outErr != nil {
			return NewFileError("Could not open output file.")
		}
		fmt.Print("[done]\n")

		tStart := time.Now()

		// Set up wpindex writer, channel for articles
		writer := NewWpindexWriter(outFile)
		articles := make(chan *StrippedArticle, 512)
		ec := NewErrorContext()
		ec.Start()
		go func() {
			n := 0

			for sa := range articles {
				n++
				if n%500 == 0 {
					cliTicker("Saving wpindex...   ", fmt.Sprintf("[article:%d  title:'%s']", n, sa.Title))
				}
				writer.WriteArticle(sa)
			}

			closeErr := writer.Close()
			if closeErr != nil {
				ec.Cancel(closeErr)
			}
			ec.Done()
		}()

		loadErr := LoadWikiCompressed(indexFile, archiveFile, func(a *Article) bool {
			if a.Redirect.Title != "" {
				// Do nothing if it's a redirect
			} else {
				// Item is normal, save it.
				sa := StrippedArticle{Title: a.Title, Links: ParseLinks(a.Text)}
				articles <- &sa
			}
			return true
		})

		if loadErr != nil {
			return cli.NewExitError("Error: Failed to read from file", 10)
		}

		close(articles)

		writerErr := ec.Wait()
		if writerErr != nil {
			return cli.NewExitError("Error: "+writerErr.Error(), 18)
		}

		dLoad := time.Since(tStart).Seconds()
		cliTicker("Saving wpindex...   ", fmt.Sprintf("[done in %4.2fs]", dLoad))
		fmt.Println()

		return nil
	},
}

var startCmd = cli.Command{
	Name:  "start",
	Usage: "Start interactive mode",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "wpindex, i",
			Usage: "Previously-built *.wpindex file",
		},
	},
	Action: func(c *cli.Context) error {
		// Open the index
		fmt.Print("Opening index...    ")
		indexFile, indexErr := os.Open(c.String("wpindex"))
		if indexErr != nil {
			return NewFileError("Could not open index.")
		}
		fmt.Print("[done]\n")

		// Create WpindexReader
		reader, readerErr := NewWpindexReader(indexFile)
		if readerErr != nil {
			return NewFileError("Could not understand index.")
		}

		cliTicker("Loading wpindex...  ", "")

		// Load all the articles.
		tLoad := time.Now()
		ind := NewIndex()

		articles := make(chan *StrippedArticle, 512)
		ec := NewErrorContext()
		ec.Start()
		go func() {
			n := 0
			for sa := range articles {
				n++
				if n%500 == 0 {
					cliTicker("Loading wpindex...  ", fmt.Sprintf("[article:%d  title: %s]", n, sa.Title))
				}
				ind.AddArticle(sa)
			}
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
		cliTicker("Loading wpindex...  ", fmt.Sprintf("[done in %4.2fs]", dLoad))
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

			names[0] = prompt("First Article")
			names[1] = prompt("Second Article")

			for i, _ := range names {
				items[i] = ind.Get(names[i])
				if items[i] == nil {
					fmt.Printf("Error: Can't find article '%s'", names[i])
					continue InputLoop
				}
			}

			tSearch := time.Now()
			fmt.Printf("\nSearching for path... ")
			nSteps := 5
			paths := ind.FindPath(items[0], items[1], nSteps)
			dSearch := time.Since(tSearch).Seconds()
			fmt.Printf("[done in %4.2f]\n", dSearch)

			if len(paths) == 0 {
				fmt.Printf("No paths found in %d steps.", nSteps)
			} else {
				fmt.Printf("Path: ")
				for i, item := range paths[0] {
					if i != 0 {
						fmt.Printf(" -> ")
					}
					fmt.Print(item.Title)
				}
				fmt.Println()
			}

			fmt.Println()
			ind.Reset()
		}

		return nil
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "wikipath"
	app.HelpName = app.Name
	app.Usage = "Find a path of links between two wiki pages."

	app.Commands = []cli.Command{indexCmd, startCmd}

	app.Run(os.Args)
}
