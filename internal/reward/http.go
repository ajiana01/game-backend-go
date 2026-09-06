package reward

import (
	"net/http"

	"github.com/ajiana01/game-backend-go/internal/httpapi"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

// Grant godoc
// @Summary Grant a reward
// @Description Grants a predefined login, battle, or quest reward. Reusing the same idempotency key returns the original result without duplicating its effects.
// @Tags rewards
// @Accept json
// @Produce json
// @Param playerID path string true "Player ID" format(uuid)
// @Param Idempotency-Key header string true "Unique request key"
// @Param input body GrantInput true "Reward type"
// @Success 200 {object} GrantResult "Existing idempotent result"
// @Success 201 {object} GrantResult "Reward created"
// @Failure 400 {object} httpapi.ErrorResponse
// @Failure 404 {object} httpapi.ErrorResponse
// @Failure 409 {object} httpapi.ErrorResponse
// @Failure 422 {object} httpapi.ErrorResponse
// @Failure 500 {object} httpapi.ErrorResponse
// @Router /api/v1/players/{playerID}/rewards [post]
func (h *Handler) Grant(w http.ResponseWriter, r *http.Request) {
	var input GrantInput
	if !httpapi.DecodeJSON(w, r, &input) {
		return
	}
	result, err := h.service.Grant(r.Context(), r.PathValue("playerID"), r.Header.Get("Idempotency-Key"), input.Type)
	if err != nil {
		httpapi.WriteServiceError(w, err)
		return
	}
	status := http.StatusCreated
	if !result.Created {
		status = http.StatusOK
	}
	httpapi.WriteJSON(w, status, result)
}
