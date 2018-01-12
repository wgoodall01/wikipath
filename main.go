package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
	"time"
)

func NewFileError(msg string) *cli.ExitError {
	return cli.NewExitError(msg, 1)
}

var listCmd = cli.Command{
	Name:  "list",
	Usage: "List all articles in an archive",
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

		// Article visitor
		callback := func(a Article) error {
			fmt.Printf("%10d%3d%30s%10s\n", a.Id, a.Namespace, a.Title, a.RevisionAuthor)
			links := ParseLinks(a.Text)
			for _, link := range links {
				fmt.Println(link)
			}
			fmt.Println()
			return nil
		}

		// Parse the archive
		parseErr := ParseWikiXML(archiveFile, callback)
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
		fmt.Print("Opening archive... ")
		tOpen := time.Now()
		path := c.String("file")
		archiveFile, fileErr := os.Open(path)
		if fileErr != nil {
			return NewFileError("Could not open wiki")
		}
		dOpen := time.Since(tOpen).Seconds()
		fmt.Printf("[done in %.2fs]\n", dOpen)

		// Load all the articles.
		fmt.Print("Loading articles... ")
		tLoad := time.Now()
		ind := NewIndex()
		visitor := func(a Article) error {
			ind.AddArticle(&a)
			return nil
		}
		ParseWikiXML(archiveFile, visitor)
		dLoad := time.Since(tLoad).Seconds()
		fmt.Printf("[done in %.2fs]\n", dLoad)

		// Index all the articles.
		fmt.Print("Making index... ")
		tBuild := time.Now()
		ind.Build()
		dBuild := time.Since(tBuild).Seconds()
		fmt.Printf("[done in %.2fs]\n", dBuild)

		// Find a path.
		running := true
		for running {
			fmt.Print("\n\n")
			var articleA string
			var articleB string

			fmt.Print("First Article  : ")
			fmt.Scanln(&articleA)

			fmt.Print("Second Article : ")
			fmt.Scanln(&articleB)

			fmt.Printf("Searching for path from %s -> %s...\n\n", articleA, articleB)

			paths := ind.FindPath(ind.Get(articleA), ind.Get(articleB), 4)
			for _, item := range paths[0] {
				fmt.Print(item.Title + ",")
			}
			fmt.Println()
			ind.Reset()
		}

		fmt.Print("Finding from 'Potato' -> 'Cyan'...")
		tFind := time.Now()
		paths := ind.FindPath(ind.Get("Potato"), ind.Get("Cyan"), 3)
		dFind := time.Since(tFind).Seconds()
		fmt.Printf("[done in %.2fs]", dFind)

		for _, path := range paths {
			for _, item := range path {
				fmt.Print(item.Title + ", ")
			}
			fmt.Println()
		}

		return nil
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "wikipath"
	app.HelpName = app.Name
	app.Usage = "Find a path of links between two wiki pages."

	app.Commands = []cli.Command{listCmd, startCmd}

	app.Run(os.Args)
}
