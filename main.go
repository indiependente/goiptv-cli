package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"time"

	"github.com/indiependente/goiptv/v2"
	spinner "github.com/indiependente/gospinner"
	flags "github.com/jessevdk/go-flags"
	au "github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	// BINARYNAME is the name of the binary.
	BINARYNAME = "goiptv-cli"
	// BINARYVERSION is the version of the binary.
	BINARYVERSION = "v1.1.3"
)

var opts struct {
	TimeSpan string   `short:"t" long:"timespan" description:"The timespan in which to search for playlists. Allowed values are: \"H\" (last hour) \"D\" (last day) \"W\" (last week)" default:"D" optional:"yes" choice:"H" choice:"D" choice:"W"` // nolint: staticcheck
	Channels []string `short:"c" long:"channel" description:"A list of tv channels" default:"sky calcio"`
	Debug    bool     `short:"d" long:"debug" description:"Run program with debug information turned on" optional:"yes"`
	Version  bool     `short:"v" long:"version" description:"Shows the program version" optional:"yes"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	start := time.Now()

	playlists, err := scrapeChannels(opts.Channels, opts.TimeSpan)
	if err != nil {
		return errors.Wrap(err, "could not get data")
	}
	numPlaylists := len(playlists)

	log.WithFields(log.Fields{"seconds": time.Since(start)}).Debug("time elapsed")

	err = persist(playlists)
	if err != nil {
		return errors.Wrap(err, "could not persist playlists")
	}

	plural := ""
	if numPlaylists > 1 {
		plural = "s"
	}
	if numPlaylists == 0 {
		fmt.Println(au.Red("\nNo playlists found! ⛔️").Bold())
	} else {
		fmt.Println(au.Sprintf(au.Bold("\nSuccessfully downloaded %d playlist%s in %.2f seconds! ⚡️"), au.Green(numPlaylists), plural, au.Blue(time.Since(start))))
	}
	return nil
}

func scrapeChannels(channels []string, timeSpan string) ([][]byte, error) {

	var eg errgroup.Group

	go spinner.Spin(os.Stdout, 100*time.Millisecond)
	fmt.Print("  Scraping and generating playlists... 🧐")
	iptvScraper := goiptv.NewIPTVScraper(timeSpan)

	content := make([][]byte, 0, len(channels))
	dataCh := make(chan []byte, len(channels))

	for _, c := range channels {
		c := c
		eg.Go(func() error {
			log.WithFields(log.Fields{"channel": c}).Debug("search")
			readers := iptvScraper.ScrapeAll("extinf " + c)

			for r := range readers {
				data, err := ioutil.ReadAll(r)
				if err != nil {
					return errors.Wrap(err, "could not read from reader")
				}
				log.WithFields(log.Fields{"bytes": len(data)}).Debug("bytes received")
				dataCh <- data
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, errors.Wrap(err, "could not scrape channels")
	}

	close(dataCh)

	for d := range dataCh {
		content = append(content, d)
	}

	return content, nil
}

func persist(playlists [][]byte) error {
	folderName := "data_" + time.Now().Format("2006-01-02")
	_ = os.Mkdir(folderName, 0755) // create and ignore issues
	for i, p := range playlists {
		err := ioutil.WriteFile(fmt.Sprintf(folderName+"/iptv%d.m3u", i), p, 0644)
		if err != nil {
			return errors.Wrap(err, "could not write file")
		}
	}
	return nil
}

func init() {

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		fmt.Printf("%s %s\n", BINARYNAME, BINARYVERSION)
		os.Exit(0)
	}

	if opts.Debug {
		fmt.Println(au.Cyan("Debug mode active"))
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}
	log.SetFormatter(&log.JSONFormatter{})

	if len(opts.Channels) == 1 && strings.EqualFold(opts.Channels[0], "") {
		fmt.Println(au.Brown("No tv channel argument provided. Defaults research to EXTINF. 🤓"))
	}

}
