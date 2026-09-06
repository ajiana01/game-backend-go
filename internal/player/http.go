package player

import (
	"net/http"

	"github.com/ajiana01/portfolio-go/internal/httpapi"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

// Register godoc
// @Summary Register a player
// @Description Creates a new player with level 1, zero EXP, and zero gold.
// @Tags players
// @Accept json
// @Produce json
// @Param input body CreateInput true "Player data"
// @Success 201 {object} Player
// @Failure 400 {object} httpapi.ErrorResponse
// @Failure 409 {object} httpapi.ErrorResponse
// @Failure 422 {object} httpapi.ErrorResponse
// @Failure 500 {object} httpapi.ErrorResponse
// @Router /api/v1/players [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if !httpapi.DecodeJSON(w, r, &input) {
		return
	}
	value, err := h.service.Register(r.Context(), input)
	if err != nil {
		httpapi.WriteServiceError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, value)
}

// Get godoc
// @Summary Get a player profile
// @Tags players
// @Produce json
// @Param playerID path string true "Player ID" format(uuid)
// @Success 200 {object} Player
// @Failure 404 {object} httpapi.ErrorResponse
// @Failure 500 {object} httpapi.ErrorResponse
// @Router /api/v1/players/{playerID} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	value, err := h.service.Get(r.Context(), r.PathValue("playerID"))
	if err != nil {
		httpapi.WriteServiceError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, value)
}

// Update godoc
// @Summary Update a player profile
// @Description Updates the player's username.
// @Tags players
// @Accept json
// @Produce json
// @Param playerID path string true "Player ID" format(uuid)
// @Param input body UpdateInput true "Updated player data"
// @Success 200 {object} Player
// @Failure 400 {object} httpapi.ErrorResponse
// @Failure 404 {object} httpapi.ErrorResponse
// @Failure 409 {object} httpapi.ErrorResponse
// @Failure 422 {object} httpapi.ErrorResponse
// @Failure 500 {object} httpapi.ErrorResponse
// @Router /api/v1/players/{playerID} [patch]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var input UpdateInput
	if !httpapi.DecodeJSON(w, r, &input) {
		return
	}
	value, err := h.service.Update(r.Context(), r.PathValue("playerID"), input)
	if err != nil {
		httpapi.WriteServiceError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, value)
}
