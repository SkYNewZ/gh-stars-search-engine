package http

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/SkYNewZ/gh-stars-search-engine/internal/engine"
	"github.com/SkYNewZ/gh-stars-search-engine/internal/slogx"
	"github.com/SkYNewZ/gh-stars-search-engine/ui"
)

const defaultPageSize int = 10

func (s *server) searchHandler(w http.ResponseWriter, r *http.Request) {
	// read q query param
	q := r.URL.Query().Get("q")
	if q == "" {
		s.responseErrorAsJSON(w, r, http.StatusBadRequest, "missing q query param")
		return
	}

	searchResponseFields := []string{
		"name_with_owner",
		"description",
		"url",
		"primary_language.name",
		"primary_language.color",
	}

	// read fields query param
	if additionalFields := r.URL.Query().Get("fields"); additionalFields != "" {
		searchResponseFields = append(searchResponseFields, strings.Split(additionalFields, ",")...)
	}

	// pagination
	pageSize := parseQueryParamPositive(r.URL.Query().Get("size"), defaultPageSize)
	from := parseQueryParamPositive(r.URL.Query().Get("from"), 0)

	ctx, cancel := context.WithTimeout(r.Context(), s.searchTimeout)
	defer cancel()

	res, err := s.search.Search(
		ctx,
		q,
		engine.WithSearchFrom(from),
		engine.WithSearchSize(pageSize),
		engine.WithSearchFields(searchResponseFields...),
	)
	if err != nil {
		s.responseErrorAsJSON(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	s.responseAsJSON(w, r, http.StatusOK, res)
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func (s *server) uiHandler(w http.ResponseWriter, r *http.Request) {
	_ui, err := fs.Sub(ui.Dist, "dist")
	if err != nil {
		s.logger.With(slogx.Err(err)).Error("failed to get sub filesystem")
		http.Error(w, "cannot render ui", http.StatusInternalServerError)
		return
	}

	http.FileServer(http.FS(_ui)).ServeHTTP(w, r)
}

// responseAsJSON writes the data as JSON to the response writer.
func (s *server) responseAsJSON(w http.ResponseWriter, _ *http.Request, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.With(slogx.Err(err)).Error("failed to encode response")
	}
}

func (s *server) responseErrorAsJSON(w http.ResponseWriter, r *http.Request, code int, msg string) {
	body := map[string]any{
		"code":    code,
		"status":  http.StatusText(code),
		"message": msg,
	}

	s.responseAsJSON(w, r, code, body)
}

// parseQueryParamPositive parses the query param v as an int.
// If the value is not an int, returns def.
// If the value is negative, returns def.
func parseQueryParamPositive(v string, def int) int {
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}

	if i <= 0 {
		return def
	}

	return i
}
