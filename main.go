package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"os"
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

var findCmd = cli.Command{
	Name:  "find",
	Usage: "Find articles from the archive.",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Usage: "Wiki archive XML file",
		},
	},
	Action: func(c *cli.Context) error {
		// Open the archive file
		path := c.String("file")
		archiveFile, fileErr := os.Open(path)
		if fileErr != nil {
			return NewFileError("Could not open wiki")
		}

		// Articles to get
		articleNames := make([]string, len(c.Args()))
		for i, arg := range c.Args() {
			articleNames[i] = NormalizeArticleTitle(arg)
		}

		callback := func(a *Article) error {
			for i, name := range articleNames {
				if NormalizeArticleTitle(a.Title) == name {

					fmt.Printf("%20s : %-50s\n", "Article Title", a.Title)
					fmt.Printf("%20s : %-50s\n", "Redirect", a.Redirect.Title)
					fmt.Printf("%20s : %-50d\n", "ID", a.Id)
					fmt.Printf("%20s : %-50d\n", "Namespace", a.Namespace)
					fmt.Printf("%20s : %-50s\n", "Timestamp", a.RevisionTimestamp)
					fmt.Printf("%20s : %-50s\n", "Author", a.RevisionAuthor)

					fmt.Printf("Text ::\n\n")
					fmt.Println(a.Text)
					fmt.Println("\n::\n\n")

					fmt.Printf("Links ::\n")
					for _, link := range ParseLinks(a.Text) {
						fmt.Print(link + ", ")
					}
					fmt.Println("::\n")

					articleNames[i] = ""

					if i == len(articleNames)-1 {
						fmt.Println("Done searching")
						return errors.New("Done searching.")
					} else {
						return nil
					}

				}
			}
			return nil
		}

		// Parse the archive
		parseErr := LoadWiki(archiveFile, callback)
		return parseErr
	},
}

var startCmd = cli.Command{
	Name:  "start",
	Usage: "Start interactive mode",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Usage: "Path to archive XML file",
		},
	},
	Action: func(c *cli.Context) error {
		// Open the archive
		fmt.Print("Opening archive...  ")
		tOpen := time.Now()
		path := c.String("file")
		archiveFile, fileErr := os.Open(path)
		if fileErr != nil {
			return NewFileError("Could not open wiki")
		}
		dOpen := time.Since(tOpen).Seconds()
		fmt.Printf("[done in %4.2fs]\n", dOpen)

		// Load all the articles.
		fmt.Print("Loading articles... ")
		tLoad := time.Now()
		ind := NewIndex()
		LoadWikiToIndex(archiveFile, ind)
		dLoad := time.Since(tLoad).Seconds()
		fmt.Printf("[done in %4.2fs]\n", dLoad)

		// Index all the articles.
		fmt.Print("Making index...     ")
		tBuild := time.Now()
		ind.Build()
		dBuild := time.Since(tBuild).Seconds()
		fmt.Printf("[done in %4.2fs]\n", dBuild)

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

	app.Commands = []cli.Command{findCmd, startCmd}

	app.Run(os.Args)
}
