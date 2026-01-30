package http

import (
	"log/slog"

	"encoding/json"
	"net/http"
	"strings"

	"api-project/internal/storage"
)

type Handler struct {
	store storage.UrlStorage
	log   *slog.Logger
}

func New(store storage.UrlStorage, log *slog.Logger) *Handler {
	return &Handler{
		store: store,
		log:   log,
	}
}

type createUrlRequest struct {
	Url string `json:"url"`
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

	// 2) простая валидация
	req.Url = strings.TrimSpace(req.Url)
	if req.Url == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "url is required",
		})
		return
	}

	// 3) сохраняем в БД
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
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}


