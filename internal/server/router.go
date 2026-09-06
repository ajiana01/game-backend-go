package server

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	_ "github.com/ajiana01/game-backend-go/docs"
	"github.com/ajiana01/game-backend-go/internal/httpapi"
	"github.com/ajiana01/game-backend-go/internal/inventory"
	"github.com/ajiana01/game-backend-go/internal/leaderboard"
	"github.com/ajiana01/game-backend-go/internal/player"
	"github.com/ajiana01/game-backend-go/internal/reward"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type ReadyCheck func(r *http.Request) error

type StatusResponse struct {
	Status string `json:"status" example:"ok"`
}

type SystemHandler struct{ ready ReadyCheck }

// Health godoc
// @Summary Check API liveness
// @Tags system
// @Produce json
// @Success 200 {object} StatusResponse
// @Router /healthz [get]
func (h SystemHandler) Health(w http.ResponseWriter, _ *http.Request) {
	httpapi.WriteJSON(w, http.StatusOK, StatusResponse{Status: "ok"})
}

// Ready godoc
// @Summary Check API readiness
// @Description Checks whether PostgreSQL, MongoDB, and Redis are available.
// @Tags system
// @Produce json
// @Success 200 {object} StatusResponse
// @Failure 503 {object} httpapi.ErrorResponse
// @Router /readyz [get]
func (h SystemHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if err := h.ready(r); err != nil {
		httpapi.WriteError(w, http.StatusServiceUnavailable, "dependencies unavailable")
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, StatusResponse{Status: "ready"})
}

func NewRouter(
	logger *slog.Logger,
	players *player.Handler,
	inventories *inventory.Handler,
	rewards *reward.Handler,
	leaderboards *leaderboard.Handler,
	ready ReadyCheck,
) http.Handler {
	mux := http.NewServeMux()
	system := SystemHandler{ready: ready}
	mux.HandleFunc("GET /healthz", system.Health)
	mux.HandleFunc("GET /readyz", system.Ready)
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusTemporaryRedirect)
	})
	mux.Handle("GET /docs/", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
	))
	mux.HandleFunc("POST /api/v1/players", players.Register)
	mux.HandleFunc("GET /api/v1/players/{playerID}", players.Get)
	mux.HandleFunc("PATCH /api/v1/players/{playerID}", players.Update)
	mux.HandleFunc("GET /api/v1/players/{playerID}/inventory", inventories.List)
	mux.HandleFunc("POST /api/v1/players/{playerID}/inventory/items", inventories.Add)
	mux.HandleFunc("DELETE /api/v1/players/{playerID}/inventory/items/{itemID}", inventories.Remove)
	mux.HandleFunc("POST /api/v1/players/{playerID}/rewards", rewards.Grant)
	mux.HandleFunc("POST /api/v1/leaderboard/scores", leaderboards.Submit)
	mux.HandleFunc("GET /api/v1/leaderboard", leaderboards.Top)
	return recoverMiddleware(logger, loggingMiddleware(logger, mux))
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(started))
	})
}

func recoverMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if value := recover(); value != nil {
				logger.Error("panic recovered", "error", value, "stack", string(debug.Stack()))
				httpapi.WriteError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
