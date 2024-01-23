package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

type Indexable interface {
	GetID() string
}

//go:generate go run github.com/vburenin/ifacemaker --file $GOFILE --struct engine --iface Engine --pkg engine --output engine_iface.go
type engine struct {
	index  bleve.Index
	logger *slog.Logger
}

// New returns a new search Engine.
func New(path string, logger *slog.Logger, mapper mapping.IndexMapping) (Engine, error) {
	var index bleve.Index
	var err error

	// use default logger if none is provided
	if logger == nil {
		logger = slog.Default()
	}

	// use default mapping if none is provided
	if mapper == nil {
		mapper = bleve.NewIndexMapping()
	}

	index, err = bleve.New(path, mapper)
	if err != nil && errors.Is(err, bleve.ErrorIndexPathExists) {
		index, err = bleve.Open(path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open index: %w", err)
	}

	return &engine{
		index:  index,
		logger: logger,
	}, nil
}

// BatchIndex indexes the given data in batches of the given size.
func (e *engine) BatchIndex(data []Indexable, batchSize int) error {
	batch := e.index.NewBatch() // init the first batch
	batchCount := 0

	flushBatch := func() error {
		e.logger.Debug(fmt.Sprintf("indexing batch (%d docs)", batchCount))

		if err := e.index.Batch(batch); err != nil {
			return fmt.Errorf("failed to index batch: %w", err)
		}

		batch = e.index.NewBatch()
		batchCount = 0

		return nil
	}

	e.logger.Debug(fmt.Sprintf("indexing %d documents", len(data)))
	for _, d := range data {
		e.logger.Debug(fmt.Sprintf("indexing %s", d.GetID()))

		// add the document to the batch
		if err := batch.Index(d.GetID(), d); err != nil {
			return fmt.Errorf("failed to index document %s: %w", d.GetID(), err)
		}
		batchCount++

		// flush the batch if it's full
		if batchCount >= batchSize {
			if err := flushBatch(); err != nil {
				return err
			}
		}
	}

	// flush the last batch and return
	return flushBatch()
}

// Search executes the given query and returns the results.
// TODO: if we want to sort by starredAt, we need to index it as a date field and use sort https://blevesearch.com/docs/Sorting/
func (e *engine) Search(ctx context.Context, q string, from int, size int, fields ...string) (*bleve.SearchResult, error) {
	// https://blevesearch.com/docs/Query-String-Query/
	query := bleve.NewQueryStringQuery(q)

	search := bleve.NewSearchRequest(query)
	search.Fields = fields // return only the given fields
	search.From = from     // return results starting from the given index
	search.Size = size     // return the given number of results

	results, err := e.index.SearchInContext(ctx, search)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return results, nil
}
