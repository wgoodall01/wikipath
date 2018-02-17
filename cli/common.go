package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"
)

func NewFileError(msg string, args ...interface{}) *cli.ExitError {
	return cli.NewExitError("wikipath: "+fmt.Sprintf(msg, args...)+"\n", 1)
}

func NewInternalError(msg string, args ...interface{}) *cli.ExitError {
	return cli.NewExitError("wikipath: "+fmt.Sprintf(msg, args...)+"\n", 2)
}

func NewUsageError(msg string, args ...interface{}) *cli.ExitError {
	return cli.NewExitError("wikipath: "+fmt.Sprintf(msg, args...)+"\n", 3)
}

func Prompt(prompt string) string {
	in := bufio.NewReader(os.Stdin)
	fmt.Printf("%-15s: ", prompt)
	valRaw, _ := in.ReadString('\n')
	return strings.TrimSpace(valRaw)
}

func PrintTicker(prompt string, tick string) {
	width := 100
	status := prompt + tick + strings.Repeat(" ", width)
	status = status[:width] + "\r"
	fmt.Print(status)
}

type RateMeasure struct {
	count   int
	average float32
	ticker  *time.Ticker
}

func NewRateMeasure(seconds float32) *RateMeasure {
	rm := &RateMeasure{
		count:   0,
		average: 0,
	}

	rm.ticker = time.NewTicker(time.Duration(seconds*1000) * time.Millisecond)

	go func() {
		for range rm.ticker.C {
			rm.average = float32(rm.count) / seconds
			rm.count = 0
		}
	}()

	return rm
}

func (this *RateMeasure) Stop() {
	this.ticker.Stop()
}

func (this *RateMeasure) Count(n int) {
	this.count += n
}

func (this *RateMeasure) Average() float32 {
	return this.average
}

type flags struct {
	WikiArchivePath cli.StringFlag
	WikiIndexPath   cli.StringFlag
	WpindexPath     cli.StringFlag
}

var WpFlags flags = flags{
	WpindexPath: cli.StringFlag{
		Name:   "wpindex, i",
		Usage:  "Path to *.wpindex file",
		EnvVar: "WPINDEX_PATH",
		Value:  "./wikis/enwiki.wpindex",
	},
	WikiArchivePath: cli.StringFlag{
		Name:   "wiki-archive, wa",
		Usage:  "Wiki archive *-multistream.xml.bz2 file.",
		EnvVar: "WIKI_ARCHIVE_PATH",
		Value:  "./wikis/enwiki-multistream.xml.bz2",
	},
	WikiIndexPath: cli.StringFlag{
		Name:   "wiki-index, wi",
		Usage:  "Wiki index *-multistream-index.txt.bz2 file.",
		EnvVar: "WIKI_INDEX_PATH",
		Value:  "./wikis/enwiki-multistream-index.txt.bz2",
	},
}
