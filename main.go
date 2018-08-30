package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/indiependente/goiptv"
	"github.com/indiependente/gospinner"
	flags "github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

var opts struct {
	Channels []string `short:"c" long:"channel" description:"A list of tv channels" default:"sky calcio"`
	Debug    bool     `short:"d" long:"debug" description:"Run program with debug information turned on" optional:"yes"`
}

func main() {

	numPlaylists, timeElapsed := scrapeChannels(opts.Channels)

	log.WithFields(log.Fields{"seconds": timeElapsed}).Debug("time elapsed")
	plural := ""
	if numPlaylists > 1 {
		plural = "s"
	}
	fmt.Printf("\nSuccessfully retrieved %d m3u playlist%s!\n", numPlaylists, plural)
}

func scrapeChannels(channels []string) (int, float64) {
	start := time.Now()
	var i int
	for _, c := range channels {
		log.WithFields(log.Fields{"channel": c}).Debug("search")
		readers := goiptv.ScrapeAll("extinf " + c)
		folderName := "data_" + start.Format("2006-01-02")

		_ = os.Mkdir(folderName, 0755)
		go spinner.Spin(os.Stdout, 100*time.Millisecond)
		fmt.Print("  Scraping and generating playlists... ")

		for r := range readers {
			data, err := ioutil.ReadAll(r)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Could not read from reader")
			}

			log.WithFields(log.Fields{"bytes": len(data)}).Debug("bytes received")
			i++

			err = ioutil.WriteFile(fmt.Sprintf(folderName+"/iptv%d.m3u", i), data, 0644)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Could not write file")
			}
		}
	}
	return i, time.Since(start).Seconds()
}

func init() {

	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse arguments: %v", err)
		os.Exit(1)
	}
	if opts.Debug {
		fmt.Printf("Debug mode active\n")
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}
	log.SetFormatter(&log.JSONFormatter{})
	if opts.Channels == nil {
		fmt.Printf("No tv channel argument provided. Defaults research to Sky Calcio\n")
	}

}
