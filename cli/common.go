package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli"
)

// NewFileError creates an error for a file I/O issue.
// Exits with code `1`.
func NewFileError(msg string, args ...interface{}) *cli.ExitError {
	return cli.NewExitError("wikipath: "+fmt.Sprintf(msg, args...)+"\n", 1)
}

// NewInternalError creates an error for an internal problem.
// Exits with code `2`.
func NewInternalError(msg string, args ...interface{}) *cli.ExitError {
	return cli.NewExitError("wikipath: "+fmt.Sprintf(msg, args...)+"\n", 2)
}

// NewUsageError creates an error for improper usage.
// Exits with code `3`.
func NewUsageError(msg string, args ...interface{}) *cli.ExitError {
	return cli.NewExitError("wikipath: "+fmt.Sprintf(msg, args...)+"\n", 3)
}

// Prompt prompts the user for input on stdin, which it then returns.
func Prompt(prompt string) string {
	in := bufio.NewReader(os.Stdin)
	fmt.Printf("%-15s: ", prompt)
	valRaw, _ := in.ReadString('\n')
	return strings.TrimSpace(valRaw)
}

// PrintTicker prints a line which overwrites the last on a tty.
func PrintTicker(prompt string, tick string) {
	width := 100
	status := prompt + tick + strings.Repeat(" ", width)
	status = status[:width] + "\r"
	fmt.Print(status)
}

// RateMeasure measures how frequently something happens.
// Add a number of events by using Add(n), and get the average
// over the last `n` seconds with Average().
type RateMeasure struct {
	mu      sync.Mutex
	count   int
	average float32
	ticker  *time.Ticker
}

// NewRateMeasure creates a RateMeasure with the given interval
// in seconds.
func NewRateMeasure(seconds float32) *RateMeasure {
	rm := &RateMeasure{
		count:   0,
		average: 0,
	}

	rm.ticker = time.NewTicker(time.Duration(seconds*1000) * time.Millisecond)

	go func() {
		for range rm.ticker.C {
			rm.mu.Lock()
			rm.average = float32(rm.count) / seconds
			rm.count = 0
			rm.mu.Unlock()
		}
	}()

	return rm
}

// Stop stops the RateMeasure.
func (rm *RateMeasure) Stop() {
	rm.ticker.Stop()
}

// Count adds `n` events to the RateMeasure.
func (rm *RateMeasure) Count(n int) {
	rm.mu.Lock()
	rm.count += n
	rm.mu.Unlock()
}

// Average returns the average number of events
// over the last interval.
func (rm *RateMeasure) Average() float32 {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.average
}

type flags struct {
	WikiArchivePath cli.StringFlag
	WikiIndexPath   cli.StringFlag
	WpindexPath     cli.StringFlag
}

// WpFlags are CLI flags shared between subcommands.
var WpFlags = flags{
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
