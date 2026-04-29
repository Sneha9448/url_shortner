package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"url-shortener/internal/models"
	"url-shortener/internal/services"

	"github.com/gorilla/mux"
)

type URLHandler struct {
	urlService *services.URLService
	domain     string
}

func NewURLHandler(urlService *services.URLService, domain string) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		domain:     domain,
	}
}

// Shorten handles the POST request to shorten a URL
func (h *URLHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	if req.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "URL is required"})
		return
	}

	// Default TTL to 0 (no expiration) if not provided, or handle as needed
	// In our service, 0 duration means no expiration.
	shortCode, err := h.urlService.ShortenURL(r.Context(), req.URL, req.TTL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to shorten URL"})
		return
	}

	shortURL := h.domain + "/" + shortCode

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.ShortenResponse{ShortURL: shortURL})
}

// Redirect handles the GET request to redirect to the original URL
func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["short_code"]

	if shortCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Short code is required"))
		return
	}

	longURL, err := h.urlService.GetOriginalURL(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, services.ErrURLNotFound) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("URL not found or has expired"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}
