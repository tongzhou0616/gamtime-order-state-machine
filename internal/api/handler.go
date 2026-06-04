package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gametime/order-state-machine/internal/order"
)

type Handler struct {
	service *order.Service
}

func NewHandler(svc *order.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/orders", h.createOrder)
	mux.HandleFunc("GET /api/v1/orders/{id}", h.getOrder)
	mux.HandleFunc("POST /api/v1/orders/{id}/authorize", h.authorizePayment)
	mux.HandleFunc("POST /api/v1/orders/{id}/complete", h.completeOrder)
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	o, err := h.service.CreateOrder()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, o)
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	o, err := h.service.GetOrder(id)
	if err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (h *Handler) authorizePayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	o, err := h.service.AuthorizePayment(id)
	if err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (h *Handler) completeOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	o, err := h.service.CompleteOrder(id)
	if err != nil {
		writeError(w, statusForError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, order.ErrOrderNotFound):
		return http.StatusNotFound
	case errors.Is(err, order.ErrInvalidTransition):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}
