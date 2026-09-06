package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ajiana01/portfolio-go/internal/inventory"
	"github.com/ajiana01/portfolio-go/internal/leaderboard"
	"github.com/ajiana01/portfolio-go/internal/player"
	"github.com/ajiana01/portfolio-go/internal/reward"
)

func TestSwaggerUIAndSpecificationAreServed(t *testing.T) {
	t.Parallel()
	router := NewRouter(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		player.NewHandler(nil),
		inventory.NewHandler(nil),
		reward.NewHandler(nil),
		leaderboard.NewHandler(nil),
		nil,
	)

	t.Run("documentation redirect", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/docs", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		if response.Code != http.StatusTemporaryRedirect {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusTemporaryRedirect)
		}
		if location := response.Header().Get("Location"); location != "/docs/index.html" {
			t.Fatalf("Location = %q", location)
		}
	})

	t.Run("swagger ui", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/docs/index.html", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		if !strings.Contains(response.Body.String(), "SwaggerUIBundle") {
			t.Fatal("response does not contain Swagger UI")
		}
	})

	t.Run("swagger specification", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/docs/doc.json", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		var specification struct {
			Info struct {
				Title string `json:"title"`
			} `json:"info"`
			Paths map[string]json.RawMessage `json:"paths"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &specification); err != nil {
			t.Fatalf("decode specification: %v", err)
		}
		if specification.Info.Title != "Game Backend Services API" {
			t.Fatalf("title = %q", specification.Info.Title)
		}
		if _, ok := specification.Paths["/api/v1/players"]; !ok {
			t.Fatal("specification does not contain player endpoints")
		}
	})
}
