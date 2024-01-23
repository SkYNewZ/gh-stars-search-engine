package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/SkYNewZ/gh-stars-search-engine/internal/slogx"
)

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
	from := r.URL.Query().Get("from")
	size := r.URL.Query().Get("size")
	parseInt := func(s string, def int) int {
		if i, err := strconv.Atoi(s); err == nil {
			return i
		}

		return def
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.searchTimeout)
	defer cancel()

	res, err := s.search.Search(ctx, q, parseInt(from, 0), parseInt(size, 10), searchResponseFields...)
	if err != nil {
		s.responseErrorAsJSON(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	s.responseAsJSON(w, r, http.StatusOK, res)
}

func (s *server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
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
