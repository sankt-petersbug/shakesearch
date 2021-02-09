package store

import (
	"io/ioutil"
	"testing"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func newTestStore(data []ShakespeareWork) *BleveStore {
	searcher, err := NewBleveStore(true)
	if err != nil {
		panic(err)
	}
	if err := searcher.BatchIndex(data); err != nil {
		panic(err)
	}
	return searcher
}

func TestBleveStore_ParseResult(t *testing.T) {
	data := []ShakespeareWork{
		{ID: "1", Title: "Title1", Content: "content1"},
		{ID: "2", Title: "Title2", Content: "content2"},
	}
	searcher := newTestStore(data)

	testCases := []struct {
		name     string
		result   *bleve.SearchResult
		total    int
		expected []Hit
	}{
		{
			name: "simple",
			result: &bleve.SearchResult{
				Total: 2,
				Hits: search.DocumentMatchCollection{
					&search.DocumentMatch{
						ID:    "1",
						Score: 1.0,
						Fragments: search.FieldFragmentMap{
							"Text": []string{"fragment"},
						},
					},
					&search.DocumentMatch{
						ID:    "2",
						Score: 1.0,
						Fragments: search.FieldFragmentMap{
							"Text": []string{"fragment"},
						},
					},
				},
			},
			total: 2,
			expected: []Hit{
				{Line: "fragment", LineNumber: 1, Score: 1.0, Title: "Title1", WorkID: "1"},
				{Line: "fragment", LineNumber: 1, Score: 1.0, Title: "Title2", WorkID: "2"},
			},
		},
		{
			name: "provide text if no fragment",
			result: &bleve.SearchResult{
				Total: 1,
				Hits: search.DocumentMatchCollection{
					&search.DocumentMatch{
						ID:        "1",
						Score:     1.0,
						Fragments: search.FieldFragmentMap{},
					},
				},
			},
			total: 1,
			expected: []Hit{
				{Line: "content1", LineNumber: 1, Score: 1.0, Title: "Title1", WorkID: "1"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var got SearchResult
			err := searcher.parseResult(tc.result, &got)
			assert.Nil(t, err)
			assert.Equal(t, tc.total, got.Meta.TotalResults)
			assert.Equal(t, tc.total, len(got.Data))
			assert.Equal(t, tc.expected, got.Data)
		})
	}
}

func TestBleveStore_Search(t *testing.T) {
	data := []ShakespeareWork{
		{ID: "1", Title: "First Title", Content: "first stemming"},
		{ID: "2", Title: "Second Title", Content: "second stemming"},
		{ID: "3", Title: "Third Title", Content: "third term"},
		{ID: "4", Title: "Last Title", Content: "picasso term"},
	}
	searcher := newTestStore(data)

	testCases := []struct {
		name     string
		options  SearchOptions
		expected []string
	}{
		{
			name: "return all if empty query",
			options: SearchOptions{
				PageNumber: 1,
				PageSize:   10,
			},
			expected: []string{"1", "2", "3", "4"},
		},
		{
			name: "exact match",
			options: SearchOptions{
				Query:      "picasso",
				PageNumber: 1,
				PageSize:   10,
			},
			expected: []string{"4"},
		},
		{
			name: "stemming",
			options: SearchOptions{
				Query:      "stem",
				PageNumber: 1,
				PageSize:   10,
			},
			expected: []string{"1", "2"},
		},
		{
			name: "multi terms",
			options: SearchOptions{
				Query:      "first nonmatch",
				PageNumber: 1,
				PageSize:   10,
			},
			expected: []string{"1"},
		},
		{
			name: "case-insensitive",
			options: SearchOptions{
				Query:      "FIRst",
				PageNumber: 1,
				PageSize:   10,
			},
			expected: []string{"1"},
		},
		{
			name: "fuzzy",
			options: SearchOptions{
				Query:      "dirst",
				Fuzziness:  1,
				PageNumber: 1,
				PageSize:   10,
			},
			expected: []string{"1"},
		},
		{
			name: "specific work",
			options: SearchOptions{
				Query:      "term",
				WorkID:     "3",
				PageNumber: 1,
				PageSize:   10,
			},
			expected: []string{"3"},
		},
		{
			name: "pagination",
			options: SearchOptions{
				Query:      "term",
				PageNumber: 2,
				PageSize:   1,
			},
			expected: []string{"4"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := searcher.Search(tc.options)
			assert.Nil(t, err)

			var got []string
			for _, hit := range result.Data {
				got = append(got, hit.WorkID)
			}
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestBleveStore_Search_SortBy(t *testing.T) {
	data := []ShakespeareWork{
		{ID: "1", Title: "TitleA", Content: "line1\nline2"},
		{ID: "2", Title: "TitleB", Content: "line3\nline4"},
	}
	searcher := newTestStore(data)

	result, err := searcher.Search(SearchOptions{
		SortBy:     []string{"-Title", "-LineNumber"},
		PageNumber: 1,
		PageSize:   10,
	})
	assert.Nil(t, err)

	var lines []string
	for _, hit := range result.Data {
		lines = append(lines, hit.Line)
	}

	assert.Equal(t, []string{"line4", "line3", "line2", "line1"}, lines)

}

func TestBleveStore_GetWorkByID(t *testing.T) {
	data := []ShakespeareWork{
		{ID: "1", Title: "TitleA", Content: "content"},
		{ID: "2", Title: "TitleB", Content: "content"},
	}
	searcher := newTestStore(data)

	work, err := searcher.GetWorkByID("1")
	assert.Nil(t, err)
	assert.Equal(t, data[0], work)
}

func TestBleveStore_GetWorkByID_NotFound(t *testing.T) {
	data := []ShakespeareWork{}
	searcher := newTestStore(data)

	_, err := searcher.GetWorkByID("1")
	assert.Equal(t, ErrWorkNotFound, err)
}

func TestBleveStore_ListTitles(t *testing.T) {
	data := []ShakespeareWork{
		{ID: "1", Title: "TitleA", Content: "content"},
		{ID: "2", Title: "TitleB", Content: "content"},
	}
	searcher := newTestStore(data)

	titles := searcher.ListTitles()
	var names []string
	for _, title := range titles {
		names = append(names, title.Title)
	}
	assert.Equal(t, []string{"TitleA", "TitleB"}, names)
}
