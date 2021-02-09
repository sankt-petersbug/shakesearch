package store

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/search/query"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrWorkNotFound is returned when trying to access work not stored in Searcher
	ErrWorkNotFound = errors.New("work not found")
)

func getFragment(frag map[string][]string) string {
	v, ok := frag["Text"]
	if !ok {
		return ""
	}
	if len(v) > 0 {
		return v[0]
	}
	return ""
}

func parseZeroPaddedNumber(s string) (int, error) {
	return strconv.Atoi(strings.TrimLeft(s, "0"))
}

func toZeroPaddedString(n int) string {
	width := 10
	return fmt.Sprintf("%0*d", width, n)
}

// SearchOptions represents the search options
// TODO: let user provide highlighter
type SearchOptions struct {
	Query      string   `query:"q"`
	Fuzziness  int      `query:"fuzziness"`
	WorkID     string   `query:"workId"`
	PageNumber int      `query:"page[number]"`
	PageSize   int      `query:"page[size]"`
	SortBy     []string `query:"sortBy"`
}

// Offset returns the number of records that will be skipped
func (s *SearchOptions) Offset() int {
	return (s.PageNumber - 1) * s.PageSize
}

// SortBySlice returns a slice of orders(field and direction) to sort the search result
// it seems that the fiber queryparser failed to parse ","" delimited query params
// correctly for mixedcase param name so this function is created to ensure that
// the parameter is properly parsedd
func (s *SearchOptions) SortBySlice() []string {
	var sortBy []string
	for _, term := range s.SortBy {
		sortBy = append(sortBy, strings.Split(term, ",")...)
	}
	return sortBy
}

// Meta represents non-standard meta-information in SearchResult
type Meta struct {
	Highlight    Highlight `json:"highlight"`
	PageNumber   int       `json:"pageNumber"`
	PageSize     int       `json:"pageSize"`
	TotalResults int       `json:"totalResults"`
}

// Highlight represents the search highlight related information
type Highlight struct {
	PostTag string `json:"postTag"`
	PreTag  string `json:"preTag"`
}

// Hit represents matched document(a single line)
type Hit struct {
	Line       string  `json:"line"`
	LineNumber int     `json:"lineNumber"`
	Score      float64 `json:"score"`
	Title      string  `json:"title"`
	WorkID     string  `json:"workId"`
}

// SearchResult represents the result of search
type SearchResult struct {
	Data []Hit `json:"data"`
	Meta Meta  `json:"meta"`
	// TODO: add pagination links
}

// ShakespeareWork represents Shakespeare's work(poem, play, sonnet, ...)
type ShakespeareWork struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Title represents a title of Shakespeare's work
type Title struct {
	Title  string `json:"title"`
	WorkID string `json:"workId"`
}

// Document represents a single line of Shakespeare's work
type Document struct {
	LineNumber string
	Text       string
	Title      string
	WorkID     string
}

// BleveStore implements methods to find and search Shakespeare's works
type BleveStore struct {
	index bleve.Index
	works map[string]ShakespeareWork
	lines map[string]Document
}

// BatchIndex batch inserts and indexes a slice of ShakespeareWork
func (b *BleveStore) BatchIndex(data []ShakespeareWork) error {
	batchSize := 10000
	batchCount := 1
	count := 1
	batch := b.index.NewBatch()
	for _, work := range data {
		b.works[work.ID] = work
		for i, line := range strings.Split(work.Content, "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			docID := strconv.Itoa(count)
			doc := Document{
				LineNumber: toZeroPaddedString(i + 1),
				Text:       line,
				Title:      work.Title,
				WorkID:     work.ID,
			}
			if err := batch.Index(docID, doc); err != nil {
				return err
			}
			b.lines[docID] = doc
			batchCount++
			count++
			if batchCount >= batchSize {
				if err := b.index.Batch(batch); err != nil {
					return err
				}
				batch = b.index.NewBatch()
				batchCount = 0
			}
		}
		log.Infof("Indexed: %s, (%d docs)", work.Title, count)
	}
	if batchCount > 0 {
		if err := b.index.Batch(batch); err != nil {
			return err
		}
	}
	return nil
}

func (b *BleveStore) parseResult(result *bleve.SearchResult, v *SearchResult) error {
	v.Meta.TotalResults = int(result.Total)
	for _, hit := range result.Hits {
		doc, ok := b.lines[hit.ID]
		if !ok {
			return errors.New(fmt.Sprintf("line not found: %s", hit.ID))
		}
		line := getFragment(hit.Fragments)
		if line == "" {
			line = doc.Text
		}

		lineNumber, err := parseZeroPaddedNumber(doc.LineNumber)
		if err != nil {
			return err
		}

		v.Data = append(
			v.Data,
			Hit{
				Score:      hit.Score,
				Line:       line,
				LineNumber: lineNumber,
				Title:      doc.Title,
				WorkID:     doc.WorkID,
			},
		)
	}

	return nil
}

// Search searches indexed documents using the search options provided
func (b *BleveStore) Search(options SearchOptions) (SearchResult, error) {
	searchResult := SearchResult{
		Data: make([]Hit, 0), // serialized to [] not null for easier parsing.
		Meta: Meta{
			Highlight: Highlight{
				PreTag:  "<mark>",
				PostTag: "</mark>",
			},
			PageNumber: options.PageNumber,
			PageSize:   options.PageSize,
		},
	}

	req, err := newSearchRequest(options)
	if err != nil {
		return searchResult, err
	}
	result, err := b.index.Search(req)
	if err != nil {
		return searchResult, err
	}
	if err := b.parseResult(result, &searchResult); err != nil {
		return searchResult, err
	}

	return searchResult, nil
}

// GetWorkByID returns a ShakespeareWork with matching id
func (b *BleveStore) GetWorkByID(id string) (ShakespeareWork, error) {
	var work ShakespeareWork
	work, ok := b.works[id]
	if !ok {
		return work, ErrWorkNotFound
	}
	return work, nil
}

// ListTitles returns a slice of work titles
func (b *BleveStore) ListTitles() []Title {
	var titles []Title
	for _, v := range b.works {
		titles = append(titles, Title{Title: v.Title, WorkID: v.ID})
	}
	sort.Slice(titles, func(i, j int) bool {
		return titles[i].Title < titles[j].Title
	})
	return titles
}

func newSearchRequest(options SearchOptions) (*bleve.SearchRequest, error) {
	var searchQuery query.Query
	if options.Query == "" {
		searchQuery = bleve.NewMatchAllQuery()
	} else {
		var queries []query.Query
		for _, term := range strings.Fields(options.Query) {
			matchQuery := bleve.NewMatchQuery(term)
			matchQuery.SetField("Text")
			matchQuery.SetFuzziness(options.Fuzziness)
			queries = append(queries, matchQuery)
		}
		searchQuery = bleve.NewDisjunctionQuery(queries...)
	}
	if options.WorkID != "" {
		idQuery := bleve.NewTermQuery(options.WorkID)
		idQuery.SetField("WorkID")
		searchQuery = bleve.NewConjunctionQuery(
			searchQuery,
			idQuery,
		)
	}

	req := bleve.NewSearchRequestOptions(
		searchQuery,
		options.PageSize,
		options.Offset(),
		false,
	)
	req.SortBy(options.SortBySlice())
	req.Highlight = bleve.NewHighlight()
	return req, nil
}

func createIndex(useInMemory bool) (bleve.Index, error) {
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = en.AnalyzerName

	mapping := bleve.NewIndexMapping()
	mapping.DefaultMapping.AddFieldMappingsAt("LineNumber", keywordFieldMapping)
	mapping.DefaultMapping.AddFieldMappingsAt("Title", keywordFieldMapping)
	mapping.DefaultMapping.AddFieldMappingsAt("Text", textFieldMapping)
	mapping.DefaultMapping.AddFieldMappingsAt("WorkID", keywordFieldMapping)

	index, err := bleve.NewMemOnly(mapping)
	if !useInMemory {
		index, err = bleve.New("shakesearch.bleve", mapping)
	}
	if err != nil {
		return nil, err
	}

	return index, nil
}

// NewBleveStore creates a new Bleve based store
func NewBleveStore(useInMemory bool) (*BleveStore, error) {
	index, err := createIndex(useInMemory)
	if err != nil {
		return nil, err
	}
	s := &BleveStore{
		index: index,
		works: make(map[string]ShakespeareWork),
		lines: make(map[string]Document),
	}
	return s, nil
}
