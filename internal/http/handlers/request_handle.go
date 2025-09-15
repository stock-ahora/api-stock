package handlers

import (
	"encoding/json"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	"github.com/stock-ahora/api-stock/internal/dto"
	"github.com/stock-ahora/api-stock/internal/service"
)

type RequestHandler struct {
	Service service.RequestService
}

func (h *RequestHandler) List(w http.ResponseWriter, r *http.Request) {
	requests, err := h.Service.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(requests)
}

func (h *RequestHandler) Create(w http.ResponseWriter, r *http.Request) {

	clientAccountIDStr := r.Header.Get("X-Client-Account-Id")
	clientAccountID, err := uuid.Parse(clientAccountIDStr)
	if clientAccountIDStr == "" && err != nil {
		http.Error(w, "El encabezado 'X-Client-Account-Id' es obligatorio", http.StatusBadRequest)
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

	req, err := h.Service.Create(requestDto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(req)
}

func (h *RequestHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "UUID inválido", http.StatusBadRequest)
		return
	}
	req, err := h.Service.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
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
