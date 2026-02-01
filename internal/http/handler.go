package http

import (
	"errors"
	"log/slog"

	"encoding/json"
	"net/http"
	"strings"

	"api-project/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	store    storage.UrlStorage
	log      *slog.Logger
	validate *validator.Validate
}

func New(store storage.UrlStorage, log *slog.Logger) *Handler {
	v := validator.New()

	return &Handler{
		store:    store,
		log:      log,
		validate: v,
	}
}

type createUrlRequest struct {
	Url string `json:"url" validate:"required,url"`
}

type createUrlResponse struct {
	ID       int64  `json:"id"`
	Alias    string `json:"alias"`
	ShortUrl string `json:"short_url"`
}

func (h *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	var req createUrlRequest

	// 1) читаем JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json",
		})
		return
	}

	req.Url = strings.TrimSpace(req.Url)

	if err := h.validate.Struct(req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":  "validation failed",
			"fields": validationErrorsToMap(err),
		})
		return
	}

	id, alias, err := h.store.Save(r.Context(), req.Url)
	if err != nil {
		h.log.Error("failed to save url", slog.Any("err", err))
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal error",
		})
		return
	}

	// 4) строим short url (пока просто из Host заголовка)
	shortUrl := "http://" + r.Host + "/" + alias

	// 5) отдаём ответ
	h.writeJSON(w, http.StatusCreated, createUrlResponse{
		ID:       id,
		Alias:    alias,
		ShortUrl: shortUrl,
	})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		// если даже ответ не смогли закодить — логируем
		h.log.Error("failed to write json", slog.Any("err", err))
	}
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	alias := chi.URLParam(r, "alias")
	alias = strings.TrimSpace(alias)
	
	if alias == "" {
		http.NotFound(w, r)
		return
	}

	var nf storage.ErrNotFound

	original, err := h.store.Get(r.Context(), alias)
	if err != nil {
		if errors.Is(err, &nf) {
			http.NotFound(w, r)
			return
		}
		h.log.Error("failed to get url", slog.Any("err", err))
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	http.Redirect(w, r, original, http.StatusFound) // 302
}

func (h *Handler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	alias := chi.URLParam(r, "alias")
	alias = strings.TrimSpace(alias)
	if alias == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "alias is required"})
		return
	}

	var nf storage.ErrNotFound

	err := h.store.Delete(r.Context(), alias)
	if err != nil {
		if errors.Is(err, &nf) {
			h.writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		h.log.Error("failed to delete url", slog.Any("err", err))
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204
}

func validationErrorsToMap(err error) map[string]string {
	out := make(map[string]string)

	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		out["_"] = "invalid request"
		return out
	}

	for _, fe := range ve {
		field := strings.ToLower(fe.Field())

		switch fe.Tag() {
		case "required":
			out[field] = "is required"
		case "url":
			out[field] = "must be a valid URL (include http/https)"
		default:
			out[field] = "is invalid"
		}
	}

	return out
}
