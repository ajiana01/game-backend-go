package inventory

import (
	"net/http"
	"strconv"

	"github.com/ajiana01/game-backend-go/internal/httpapi"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

// List godoc
// @Summary Get a player's inventory
// @Tags inventory
// @Produce json
// @Param playerID path string true "Player ID" format(uuid)
// @Success 200 {object} ListResponse
// @Failure 404 {object} httpapi.ErrorResponse
// @Failure 500 {object} httpapi.ErrorResponse
// @Router /api/v1/players/{playerID}/inventory [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context(), r.PathValue("playerID"))
	if err != nil {
		httpapi.WriteServiceError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, ListResponse{Items: items})
}

// Add godoc
// @Summary Add an inventory item
// @Description Adds the quantity to an existing item or creates a new item.
// @Tags inventory
// @Accept json
// @Produce json
// @Param playerID path string true "Player ID" format(uuid)
// @Param input body AddInput true "Item data"
// @Success 200 {object} Item
// @Failure 400 {object} httpapi.ErrorResponse
// @Failure 404 {object} httpapi.ErrorResponse
// @Failure 422 {object} httpapi.ErrorResponse
// @Failure 500 {object} httpapi.ErrorResponse
// @Router /api/v1/players/{playerID}/inventory/items [post]
func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	var input AddInput
	if !httpapi.DecodeJSON(w, r, &input) {
		return
	}
	item, err := h.service.Add(r.Context(), r.PathValue("playerID"), input)
	if err != nil {
		httpapi.WriteServiceError(w, err)
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, item)
}

// Remove godoc
// @Summary Remove an inventory item
// @Description Removes the requested quantity. The item is deleted when its quantity reaches zero.
// @Tags inventory
// @Param playerID path string true "Player ID" format(uuid)
// @Param itemID path string true "Item ID"
// @Param quantity query int false "Quantity to remove" default(1) minimum(1)
// @Success 204
// @Failure 400 {object} httpapi.ErrorResponse
// @Failure 404 {object} httpapi.ErrorResponse
// @Failure 422 {object} httpapi.ErrorResponse
// @Failure 500 {object} httpapi.ErrorResponse
// @Router /api/v1/players/{playerID}/inventory/items/{itemID} [delete]
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	quantity := 1
	if raw := r.URL.Query().Get("quantity"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "quantity must be an integer")
			return
		}
		quantity = parsed
	}
	if err := h.service.Remove(r.Context(), r.PathValue("playerID"), r.PathValue("itemID"), quantity); err != nil {
		httpapi.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
