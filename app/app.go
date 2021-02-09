package app

import (
	"errors"
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

type App struct {
	store *store.BleveStore
	api   *fiber.App
}

// Load loads data to the store
func (a *App) Load(works []store.ShakespeareWork) error {
	log.Info("Start indexing documents")
	start := time.Now()
	if err := a.store.BatchIndex(works); err != nil {
		return err
	}
	duration := time.Since(start)
	log.Infof("Finished indexing. Took %d seconds", int(duration.Seconds()))
	return nil
}

// Listen runs server on a port
func (a *App) Listen(port string) error {
	addr := fmt.Sprintf(":%s", port)
	return a.api.Listen(addr)
}

// NewApp initializes and returns a server app
func NewApp() (*App, error) {
	bleveStore, err := store.NewBleveStore(false)
	if err != nil {
		return nil, err
	}
	app := &App{
		store: bleveStore,
		api:   newFiberApp(bleveStore),
	}
	return app, nil
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
			if errors.Is(err, store.ErrWorkNotFound) {
				return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("work not found: %s", id))
			}
			return err
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
	log.Info("Initialized api")
	return app
}
