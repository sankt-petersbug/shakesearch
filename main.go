package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"unicode"

	log "github.com/sirupsen/logrus"

	"github.com/sankt-petersbug/shakesearch/app"
	"github.com/sankt-petersbug/shakesearch/store"
)

func sanitizeTitle(s string) string {
	return strings.Map(func(c rune) rune {
		if unicode.IsLetter(c) {
			return c
		}
		return -1
	}, s)
}

func readData(fpath string) ([]store.ShakespeareWork, error) {
	log.Infof("Reading data from %s", fpath)
	byt, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	var works []store.ShakespeareWork
	if err := json.Unmarshal(byt, &works); err != nil {
		return nil, err
	}
	for _, work := range works {
		work.ID = sanitizeTitle(work.Title)
	}
	log.Infof("Total %d works found", len(works))
	return works, nil
}

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	works, err := readData("data.json")
	if err != nil {
		panic(err)
	}

	app, err := app.NewApp()
	if err != nil {
		panic(err)
	}
	go func() {
		if err := app.Load(works); err != nil {
			panic(err)
		}
	}()

	app.Listen(port)
	log.Infof("Server running on %s", port)
}
