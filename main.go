package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/urfave/cli"
	"os"
)

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
			return fileErr
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
		path := c.String("file")
		archiveFile, fileErr := os.Open(path)
		if fileErr != nil {
			return fileErr
		}
		fmt.Print("[done]\n")

		// Load all the articles.
		fmt.Print("Loading articles... ")
		ind := NewIndex()
		visitor := func(a Article) error {
			ind.AddArticle(&a)
			return nil
		}
		ParseWikiXML(archiveFile, visitor)
		fmt.Print("[done]\n")

		// Index all the articles.
		fmt.Print("Making index... ")
		ind.Build()
		fmt.Print("[done]\n")

		spew.Dump(ind.Get("A"))

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
