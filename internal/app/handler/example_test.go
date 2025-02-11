package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/shekshuev/shortener/internal/app/config"
	"github.com/shekshuev/shortener/internal/app/handler"
	"github.com/shekshuev/shortener/internal/app/mocks"
	"github.com/shekshuev/shortener/internal/app/models"
	"github.com/shekshuev/shortener/internal/app/service"
)

// ExampleURLHandler_createURLHandler демонстрирует создание короткого URL через `POST /`.
func ExampleURLHandler_createURLHandler() {
	cfg := config.GetConfig()
	store := mocks.NewURLStore()
	svc := service.NewURLService(store, &cfg)
	h := handler.NewURLHandler(svc)

	reqBody := []byte("http://example.com")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	h.Router.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	// Output: 201
}

// ExampleURLHandler_createURLHandlerJSON демонстрирует создание короткого URL через `POST /api/shorten`.
func ExampleURLHandler_createURLHandlerJSON() {
	cfg := config.GetConfig()
	store := mocks.NewURLStore()
	svc := service.NewURLService(store, &cfg)
	h := handler.NewURLHandler(svc)

	reqBody, _ := json.Marshal(models.ShortURLCreateDTO{URL: "http://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Router.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	// Output: 201
}

// ExampleURLHandler_getURLHandler демонстрирует редирект по сокращённому URL `GET /{shorted}`.
func ExampleURLHandler_getURLHandler() {
	cfg := config.GetConfig()
	store := mocks.NewURLStore()
	svc := service.NewURLService(store, &cfg)
	h := handler.NewURLHandler(svc)

	reqBody, _ := json.Marshal(models.ShortURLCreateDTO{URL: "http://example.com"})
	createReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	createReq.Header.Set("Content-Type", "application/json")

	rrCreate := httptest.NewRecorder()
	h.Router.ServeHTTP(rrCreate, createReq)

	if rrCreate.Code != http.StatusCreated {
		fmt.Println("failed to create short URL")
		return
	}

	shortURL := strings.TrimSpace(rrCreate.Body.String())
	shortID := strings.TrimPrefix(shortURL, cfg.BaseURL+"/")

	req := httptest.NewRequest(http.MethodGet, "/"+shortID, nil)
	rr := httptest.NewRecorder()
	h.Router.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	// Output: 307
}

// ExampleURLHandler_getUserURLsHandler демонстрирует получение всех сокращённых URL пользователя через `GET /api/user/urls`.
func ExampleURLHandler_getUserURLsHandler() {
	cfg := config.GetConfig()
	store := mocks.NewURLStore()
	svc := service.NewURLService(store, &cfg)
	h := handler.NewURLHandler(svc)

	reqBody, _ := json.Marshal(models.ShortURLCreateDTO{URL: "http://example.com"})
	createReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	createReq.Header.Set("Content-Type", "application/json")

	rrCreate := httptest.NewRecorder()
	h.Router.ServeHTTP(rrCreate, createReq)

	if rrCreate.Code != http.StatusCreated {
		fmt.Println("failed to create short URL")
		return
	}

	createResult := rrCreate.Result()
	defer createResult.Body.Close()

	cookies := createResult.Cookies()

	getReq := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)

	for _, cookie := range cookies {
		getReq.AddCookie(cookie)
	}

	rrGet := httptest.NewRecorder()
	h.Router.ServeHTTP(rrGet, getReq)

	defer rrGet.Result().Body.Close()

	fmt.Println(rrGet.Code)
	// Output: 200
}

// ExampleURLHandler_deleteUserURLsHandler демонстрирует удаление списка URL пользователя через `DELETE /api/user/urls`.
func ExampleURLHandler_deleteUserURLsHandler() {
	cfg := config.GetConfig()
	store := mocks.NewURLStore()
	svc := service.NewURLService(store, &cfg)
	h := handler.NewURLHandler(svc)

	reqBody, _ := json.Marshal(models.ShortURLCreateDTO{URL: "http://example.com"})
	createReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(reqBody))
	createReq.Header.Set("Content-Type", "application/json")

	rrCreate := httptest.NewRecorder()
	h.Router.ServeHTTP(rrCreate, createReq)

	if rrCreate.Code != http.StatusCreated {
		fmt.Println("failed to create short URL")
		return
	}
	createResult := rrCreate.Result()
	defer createResult.Body.Close()

	cookies := createResult.Cookies()

	shortURL := strings.TrimSpace(rrCreate.Body.String())
	shortID := strings.TrimPrefix(shortURL, cfg.BaseURL+"/")

	deleteBody, _ := json.Marshal([]string{shortID})
	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(deleteBody))
	deleteReq.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		deleteReq.AddCookie(cookie)
	}

	rrDelete := httptest.NewRecorder()
	h.Router.ServeHTTP(rrDelete, deleteReq)

	fmt.Println(rrDelete.Code)
	// Output: 202
}

// ExampleURLHandler_batchCreateURLHandlerJSON демонстрирует пакетное создание сокращённых URL через `POST /api/shorten/batch`.
func ExampleURLHandler_batchCreateURLHandlerJSON() {
	cfg := config.GetConfig()
	store := mocks.NewURLStore()
	svc := service.NewURLService(store, &cfg)
	h := handler.NewURLHandler(svc)

	reqBody, _ := json.Marshal([]models.BatchShortURLCreateDTO{
		{CorrelationID: "1", OriginalURL: "http://google.com"},
		{CorrelationID: "2", OriginalURL: "http://github.com"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.Router.ServeHTTP(rr, req)

	fmt.Println(rr.Code)
	// Output: 201
}
