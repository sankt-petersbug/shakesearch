package main

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/sankt-petersbug/shakesearch/app"
	"github.com/sankt-petersbug/shakesearch/store"
)

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
	for i := 0; i < len(works); i++ {
		works[i].ID = i + 1
	}
	log.Infof("Total %d works found", len(works))
	return works, nil
}

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	works, err := readData("data.json")
	if err != nil {
		panic(err)
	}
	app, err := app.NewApp(works)
	if err != nil {
		panic(err)
	}
	addr := ":3000"
	if err := app.Listen(addr); err != nil {
		panic(err)
	}
	log.Infof("Server running on %s", addr)
}
