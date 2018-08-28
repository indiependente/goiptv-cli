package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/indiependente/goiptv"
	"github.com/indiependente/gospinner"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.ErrorLevel)
	log.SetFormatter(&log.JSONFormatter{})

	start := time.Now()
	readers := goiptv.ScrapeAll("extinf sky calcio")
	folderName := "data_" + start.Format("2006-01-02")

	_ = os.Mkdir(folderName, 0755)
	go spinner.Spin(os.Stdout, 100*time.Millisecond)
	fmt.Print("  Scraping and generating playlists... ")

	var i int
	for r := range readers {
		data, err := ioutil.ReadAll(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not read from reader: %v", err)
		}
		log.WithFields(log.Fields{"bytes": len(data)}).Debug("bytes received")
		i++
		ioutil.WriteFile(fmt.Sprintf(folderName+"/iptv%d.m3u", i), data, 0644)
	}
	log.WithFields(log.Fields{"seconds": time.Since(start).Seconds()}).Debug("time elapsed")
	plural := ""
	if i > 1 {
		plural = "s"
	}
	fmt.Printf("\nSuccessfully retrieved %d m3u playlist%s!\n", i, plural)
}
