package app

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/sankt-petersbug/shakesearch/store"
)

var errorHandler = func(c *fiber.Ctx, err error) error {
	if e, ok := err.(*fiber.Error); ok {
		return c.Status(e.Code).JSON(e)
	}
	code := fiber.StatusInternalServerError
	return c.Status(code).JSON(fiber.Map{
		"code":    code,
		"message": "Internal Server Error",
	})
}

type Store interface {
	ListTitles() []store.Title
	GetWorkByID(id string) (store.ShakespeareWork, error)
	Search(options store.SearchOptions) (store.SearchResult, error)
}

// NewApp returns initialized fiber app
func NewApp(works []store.ShakespeareWork) (*fiber.App, error) {
	bleveSearcher, err := store.NewBleveStore(false)
	if err != nil {
		return nil, err
	}
	log.Info("Start indexing documents")
	start := time.Now()
	if err := bleveSearcher.BatchIndex(works); err != nil {
		return nil, err
	}
	duration := time.Since(start)
	log.Infof("Finished indexing. Took %d seconds", int(duration.Seconds()))

	return newFiberApp(bleveSearcher), nil
}

func newFiberApp(s Store) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
	})
	app.Static("/", "./static")
	app.Get("/titles", func(c *fiber.Ctx) error {
		return c.JSON(s.ListTitles())
	})
	app.Get("/works/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		work, err := s.GetWorkByID(id)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("work not found: %s", id))
		}
		return c.JSON(work)
	})
	app.Get("/search", func(c *fiber.Ctx) error {
		options := store.SearchOptions{
			PageSize:   20,
			PageNumber: 1,
			SortBy:     []string{"Title", "LineNumber"}, // TODO: case insensitive sort by options
		}
		if err := c.QueryParser(&options); err != nil {
			fmt.Println(err)
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		searchResult, err := s.Search(options)
		if err != nil {
			return err
		}
		return c.JSON(searchResult)
	})
	log.Info("Initialize app")
	return app
}
