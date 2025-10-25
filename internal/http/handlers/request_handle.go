package handlers

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/dto"
	"github.com/stock-ahora/api-stock/internal/service/request"
)

type RequestHandler struct {
	Service request.RequestService
}

func (h *RequestHandler) List(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	clientAccountId, err, _ := getClientAccountIdHeader(w, r)
	page, size := parsePagination(r)

	requests, err := h.Service.List(ctx, clientAccountId, page, size)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

func (h *RequestHandler) Create(w http.ResponseWriter, r *http.Request) {

	clientAccountID, err, done := getClientAccountIdHeader(w, r)
	if done {
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error al procesar el formulario: "+err.Error(), http.StatusBadRequest)
		return
	}

	requestType := r.FormValue("type")
	if requestType == "" {
		http.Error(w, "El campo 'type' es obligatorio", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error al obtener el archivo: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileType := detectContentType(file, fileHeader)

	requestDto := &dto.CreateRequestDto{
		Type:            dto.ParseTypeStatus(requestType),
		File:            file,
		FileName:        fileHeader.Filename,
		FileSize:        fileHeader.Size,
		FileType:        fileType,
		ClientAccountId: clientAccountID,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	req, err := h.Service.Create(requestDto, ctx)
	if err != nil {
		http.Error(w, "messaging busy", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(req)
	if err != nil {
		return
	}
}

func getClientAccountIdHeader(w http.ResponseWriter, r *http.Request) (uuid.UUID, error, bool) {
	clientAccountIDStr := r.Header.Get("X-Client-Account-Id")
	clientAccountID, err := uuid.Parse(clientAccountIDStr)
	if clientAccountIDStr == "" && err != nil {
		http.Error(w, "El encabezado 'X-Client-Account-Id' es obligatorio", http.StatusBadRequest)
		return uuid.UUID{}, nil, true
	}
	return clientAccountID, err, false
}

func (h *RequestHandler) Get(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	segments := strings.Split(r.URL.Path, "/")
	idStr := segments[len(segments)-1] // última parte del path
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "UUID inválido", http.StatusBadRequest)
		return
	}
	req, err := h.Service.Get(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(req)
}

func detectContentType(file multipart.File, header *multipart.FileHeader) string {
	if ct := header.Header.Get("Content-Type"); ct != "" {
		return ct
	}
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	return http.DetectContentType(buf[:n])
}

func parsePagination(r *http.Request) (page, size int) {
	q := r.URL.Query()

	// PAGE
	pageStr := q.Get("page")
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	} else {
		page = 1
	}

	// SIZE
	sizeStr := q.Get("size")
	if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
		size = s
	} else {
		size = 10
	}

	return
}
