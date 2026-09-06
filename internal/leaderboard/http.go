package leaderboard

import (
	"net/http"
	"strconv"

	"github.com/ajiana01/game-backend-go/internal/httpapi"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

// Submit godoc
// @Summary Submit a leaderboard score
// @Description Stores the score only when it is greater than the player's current best score.
// @Tags leaderboard
// @Accept json
// @Produce json
// @Param input body Score true "Player score"
// @Success 200 {object} Score
// @Failure 400 {object} httpapi.ErrorResponse
// @Failure 404 {object} httpapi.ErrorResponse
// @Failure 422 {object} httpapi.ErrorResponse
// @Failure 500 {object} httpapi.ErrorResponse
// @Router /api/v1/leaderboard/scores [post]
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	var input Score
	if !httpapi.DecodeJSON(w, r, &input) {
		return
	}
	if err := h.service.Submit(r.Context(), input.PlayerID, input.Score); err != nil {
		httpapi.WriteServiceError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, input)
}

// Top godoc
// @Summary Get the top players
// @Tags leaderboard
// @Produce json
// @Param limit query int false "Maximum number of players" default(10) minimum(1) maximum(100)
// @Success 200 {object} ListResponse
// @Failure 400 {object} httpapi.ErrorResponse
// @Failure 422 {object} httpapi.ErrorResponse
// @Failure 500 {object} httpapi.ErrorResponse
// @Router /api/v1/leaderboard [get]
func (h *Handler) Top(w http.ResponseWriter, r *http.Request) {
	limit := int64(10)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "limit must be an integer")
			return
		}
		limit = parsed
	}
	entries, err := h.service.Top(r.Context(), limit)
	if err != nil {
		httpapi.WriteServiceError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, ListResponse{Players: entries})
}
