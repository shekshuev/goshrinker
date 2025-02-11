package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/shekshuev/shortener/internal/app/jwt"
	"github.com/shekshuev/shortener/internal/app/middleware"
	"github.com/shekshuev/shortener/internal/app/models"
	"github.com/shekshuev/shortener/internal/app/service"
	"github.com/shekshuev/shortener/internal/app/store"
)

// URLHandler обрабатывает HTTP-запросы для управления сокращёнными URL.
type URLHandler struct {
	service service.Service
	Router  *chi.Mux
}

// NewURLHandler создаёт новый экземпляр URLHandler с зарегистрированными маршрутами.
func NewURLHandler(service service.Service) *URLHandler {
	router := chi.NewRouter()
	router.Use(middleware.RequestAuth)
	router.Use(middleware.RequestLogger)
	router.Use(middleware.GzipCompressor)
	h := &URLHandler{service: service, Router: router}
	router.Post("/", h.createURLHandler)
	router.Post("/api/shorten", h.createURLHandlerJSON)
	router.Post("/api/shorten/batch", h.batchCreateURLHandlerJSON)
	router.Get("/{shorted}", h.getURLHandler)
	router.Get("/api/user/urls", h.getUserURLsHandler)
	router.Delete("/api/user/urls", h.deleteUserURLsHandler)
	router.Get("/ping", h.pingURLHandler)
	return h
}

// createURLHandler обрабатывает создание короткого URL из обычного.
// Запрос: `POST /`, тело — строка с URL.
// Ответ: 201 Created + короткий URL, либо 409 Conflict, если URL уже существует.
func (h *URLHandler) createURLHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	cookie, err := jwt.GetAuthCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	userID, err := jwt.GetUserID(cookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	shortURL, err := h.service.CreateShortURL(string(body), userID)
	switch {
	case errors.Is(err, store.ErrAlreadyExists):
		w.WriteHeader(http.StatusConflict)
	case err != nil:
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	default:
		w.WriteHeader(http.StatusCreated)
	}
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// createURLHandlerJSON обрабатывает создание короткого URL через JSON.
// Запрос: `POST /api/shorten`, тело — JSON {"url": "http://example.com"}.
// Ответ: 201 Created + JSON {"result": "short_url"}, либо 409 Conflict.
func (h *URLHandler) createURLHandlerJSON(w http.ResponseWriter, r *http.Request) {
	var createDTO models.ShortURLCreateDTO
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(body, &createDTO); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	cookie, err := jwt.GetAuthCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	userID, err := jwt.GetUserID(cookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	shortURL, err := h.service.CreateShortURL(createDTO.URL, userID)

	switch {
	case errors.Is(err, store.ErrAlreadyExists):
		w.WriteHeader(http.StatusConflict)
	case err != nil:
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	default:
		w.WriteHeader(http.StatusCreated)
	}

	readDTO := models.ShortURLReadDTO{Result: shortURL}
	resp, err := json.Marshal(readDTO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// getURLHandler обрабатывает редирект по сокращённому URL.
// Запрос: `GET /{shorted}`.
// Ответ: 307 Temporary Redirect на оригинальный URL или 410 Gone, если URL удалён.
func (h *URLHandler) getURLHandler(w http.ResponseWriter, r *http.Request) {
	urlPath := path.Base(r.URL.Path)
	if longURL, err := h.service.GetLongURL(urlPath); err == nil {
		http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
	} else {
		if err == store.ErrAlreadyDeleted {
			w.WriteHeader(http.StatusGone)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

	}
}

// getUserURLsHandler получает список всех сокращённых URL пользователя.
// Запрос: `GET /api/user/urls`.
// Ответ: 200 OK + JSON или 204 No Content, если URL нет.
func (h *URLHandler) getUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := jwt.GetAuthCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
	}
	userID, err := jwt.GetUserID(cookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application/json")
	if readDTO, err := h.service.GetUserURLs(userID); err == nil {
		resp, err := json.Marshal(readDTO)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

}

// deleteUserURLsHandler удаляет список URL пользователя.
// Запрос: `DELETE /api/user/urls`, тело — JSON-массив сокращённых URL.
// Ответ: 202 Accepted.
func (h *URLHandler) deleteUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := jwt.GetAuthCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
	}
	userID, err := jwt.GetUserID(cookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var urls []string
	if err = json.Unmarshal(body, &urls); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	go h.service.DeleteURLs(userID, urls)
	w.WriteHeader(http.StatusAccepted)
}

// pingURLHandler проверяет доступность базы данных.
// Запрос: `GET /ping`.
// Ответ: 200 OK, если БД работает, иначе 500 Internal Server Error.
func (h *URLHandler) pingURLHandler(w http.ResponseWriter, _ *http.Request) {
	err := h.service.CheckDBConnection()
	if err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// batchCreateURLHandlerJSON создаёт несколько сокращённых URL за один запрос.
// Запрос: `POST /api/shorten/batch`, тело — JSON-массив объектов { "url": "http://example.com" }.
// Ответ: 201 Created + JSON-массив результатов, либо 409 Conflict.
func (h *URLHandler) batchCreateURLHandlerJSON(w http.ResponseWriter, r *http.Request) {
	var createDTO []models.BatchShortURLCreateDTO
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal([]byte(body), &createDTO); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	cookie, err := jwt.GetAuthCookie(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	userID, err := jwt.GetUserID(cookie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	readDTO, err := h.service.BatchCreateShortURL(createDTO, userID)
	switch {
	case errors.Is(err, store.ErrAlreadyExists):
		w.WriteHeader(http.StatusConflict)
	case err != nil:
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	default:
		w.WriteHeader(http.StatusCreated)
	}

	resp, err := json.Marshal(readDTO)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
