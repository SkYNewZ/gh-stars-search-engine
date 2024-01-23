package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/v2/analysis/lang/en"
	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/SkYNewZ/gh-stars-search-engine/internal/engine"
	"github.com/SkYNewZ/gh-stars-search-engine/internal/github"
	ihttp "github.com/SkYNewZ/gh-stars-search-engine/internal/http"
	"github.com/SkYNewZ/gh-stars-search-engine/internal/logging"
	"github.com/SkYNewZ/gh-stars-search-engine/internal/slogx"
)

//go:generate go run go-simpler.org/sloggen --config .slog.config.yaml --dir internal

const (
	indexPath         string = "ghs.belve"
	indexingBatchSize int    = 100
)

func main() {
	ctx := context.Background()
	logger := logging.New(slog.LevelDebug)
	traceClient := &http.Client{Transport: logging.NewLoggerTransport(logger.With(slogx.Component("http")))}

	logger.Debug("creating GitHub graphQL client")
	client, err := github.New(ctx, github.WithHTTPClient(traceClient), github.WithLogger(logger.With(slogx.Component("github")))) // default reads GITHUB_TOKEN
	if err != nil {
		logger.With(slogx.Err(err)).Error("failed to create GitHub client")
		os.Exit(-1)
	}

	logger.Debug("creating search engine")
	mapper, err := buildGitHubRepositoryIndexMapping()
	if err != nil {
		logger.With(slogx.Err(err)).Error("failed to create mapping")
		os.Exit(-1)
	}

	search, err := engine.New(getEnvOrDefault("BELVE_STORAGE_PATH", indexPath), logger.With(slogx.Component("engine")), mapper) // use the default mapping
	if err != nil {
		logger.With(slogx.Err(err)).Error("failed to create search engine")
		os.Exit(-1)
	}

	logger.Debug("configure scheduler")
	schedulerLogger := logger.With(slogx.Component("scheduler"))
	scheduler, err := setupScheduler(getEnvOrDefault("LOCATION", "Europe/Paris"), schedulerLogger)
	if err != nil {
		logger.With(slogx.Err(err)).Error("failed to create scheduler")
		os.Exit(-1)
	}

	if _, err := scheduler.AddFunc(getEnvOrDefault("REFRESH_JOB_SCHEDULE", "0 */12 * * *"), index(ctx, client, search, schedulerLogger)); err != nil {
		logger.With(slogx.Err(err)).Error("failed to add index job to scheduler")
		os.Exit(-1)
	}

	logger.Debug("configure HTTP server")
	srv := ihttp.NewServer(logger.With(slogx.Component("server")), search, time.Minute)

	go srv.Start()
	go scheduler.Run()
	if os.Getenv("NO_INITIAL_INDEX") == "" {
		go index(ctx, client, search, logger)() // index once at startup
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	srv.Stop(ctx)
	scheduler.Stop()
}

// index periodically all stars.
func index(ctx context.Context, g github.Client, s engine.Engine, logger *slog.Logger) func() {
	return func() {
		logger.Info("fetching stars")
		repos := make([]engine.Indexable, 0)
		for starredRepo := range g.GetStars(ctx) {
			repos = append(repos, starredRepo.Repository)
		}

		logger.Debug(fmt.Sprintf("indexing %d stars", len(repos)))
		if err := s.BatchIndex(repos, indexingBatchSize); err != nil {
			logger.With(slogx.Err(err)).Error("failed to index stars")
		}
	}
}

func buildGitHubRepositoryIndexMapping() (mapping.IndexMapping, error) {
	// a generic reusable mapping for english text
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName

	// a generic reusable mapping for keyword text
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	// readme field mapping
	readmeMapping := bleve.NewTextFieldMapping()
	readmeMapping.Store = false // do not store the content of the field
	readmeMapping.Analyzer = en.AnalyzerName

	repoMapping := bleve.NewDocumentMapping()
	repoMapping.AddFieldMappingsAt("id", keywordFieldMapping)
	repoMapping.AddFieldMappingsAt("name_with_owner", englishTextFieldMapping)
	repoMapping.AddFieldMappingsAt("description", englishTextFieldMapping)
	repoMapping.AddFieldMappingsAt("readme", readmeMapping)

	repoMapping.AddFieldMappingsAt("primary_language.id", keywordFieldMapping)
	repoMapping.AddFieldMappingsAt("primaryLanguage.name", keywordFieldMapping)
	repoMapping.AddFieldMappingsAt("primaryLanguage.color", keywordFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultAnalyzer = en.AnalyzerName
	indexMapping.DefaultMapping = repoMapping

	if err := indexMapping.Validate(); err != nil {
		return nil, fmt.Errorf("invalid mapping: %w", err)
	}

	return indexMapping, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
