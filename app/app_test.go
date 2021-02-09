package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/sankt-petersbug/shakesearch/store"
)

var defaultErr = errors.New("error")

func init() {
	log.SetOutput(ioutil.Discard)
}

type fakeStore struct {
	listTitlesFunc  func() []store.Title
	getWorkByIDFunc func(id int) (store.ShakespeareWork, error)
	searchFunc      func(store.SearchOptions) (store.SearchResult, error)
}

func (f *fakeStore) ListTitles() []store.Title {
	if f.listTitlesFunc != nil {
		return f.listTitlesFunc()
	}
	return nil
}

func (f *fakeStore) GetWorkByID(id int) (store.ShakespeareWork, error) {
	if f.getWorkByIDFunc != nil {
		return f.getWorkByIDFunc(id)
	}
	return store.ShakespeareWork{ID: id}, nil
}

func (f *fakeStore) Search(options store.SearchOptions) (store.SearchResult, error) {
	if f.searchFunc != nil {
		return f.searchFunc(options)
	}
	return store.SearchResult{}, nil
}

func readRespBody(body io.ReadCloser) (store.ShakespeareWork, error) {
	defer body.Close()
	var work store.ShakespeareWork

	byt, err := ioutil.ReadAll(body)
	if err != nil {
		return work, err
	}
	if err := json.Unmarshal(byt, &work); err != nil {
		return work, err
	}
	return work, nil
}

func TestRoute_WorkByID_Success(t *testing.T) {
	app := newFiberApp(&fakeStore{})
	req, err := http.NewRequest("GET", "/works/1", nil)
	if err != nil {
		panic(err)
	}

	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	work, err := readRespBody(resp.Body)
	assert.Nil(t, err)
	assert.Equal(t, 1, work.ID)
}

func TestRoute_WorkByID_Errors(t *testing.T) {
	testCases := []struct {
		name            string
		id              string
		getWorkByIDFunc func(id int) (store.ShakespeareWork, error)
		statusCode      int
	}{
		{
			name:       "non numeric id",
			id:         "str",
			statusCode: http.StatusBadRequest,
		},
		{
			name: "not found",
			id:   "1",
			getWorkByIDFunc: func(id int) (store.ShakespeareWork, error) {
				return store.ShakespeareWork{}, defaultErr
			},
			statusCode: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := newFiberApp(&fakeStore{
				getWorkByIDFunc: tc.getWorkByIDFunc,
			})
			req, err := http.NewRequest("GET", fmt.Sprintf("/works/%s", tc.id), nil)
			if err != nil {
				panic(err)
			}

			resp, err := app.Test(req)
			assert.Nil(t, err)
			assert.Equal(t, tc.statusCode, resp.StatusCode)
		})
	}
}

func TestRoute_Search_InvalidQueryParams(t *testing.T) {
	app := newFiberApp(&fakeStore{})
	req, err := http.NewRequest("GET", "/search?fuzziness=yes", nil)
	if err != nil {
		panic(err)
	}

	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRoute_Search_SearcherError(t *testing.T) {
	app := newFiberApp(&fakeStore{
		searchFunc: func(options store.SearchOptions) (store.SearchResult, error) {
			return store.SearchResult{}, defaultErr
		},
	})
	req, err := http.NewRequest("GET", "/search", nil)
	if err != nil {
		panic(err)
	}

	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
