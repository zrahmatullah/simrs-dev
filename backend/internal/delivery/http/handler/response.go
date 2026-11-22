package handler

import (
	"encoding/json"
	"math"
	"net/http"

	"simrs-asoy/backend/internal/domain"
)

func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func RespondSuccess(w http.ResponseWriter, message string, data interface{}) {
	RespondJSON(w, http.StatusOK, domain.ResponseEnvelope{
		Success: true,
		Message: message,
		Data:    data,
		Errors:  nil,
	})
}

func RespondCreated(w http.ResponseWriter, message string, data interface{}) {
	RespondJSON(w, http.StatusCreated, domain.ResponseEnvelope{
		Success: true,
		Message: message,
		Data:    data,
		Errors:  nil,
	})
}

func RespondMeta(w http.ResponseWriter, message string, data interface{}, page, limit, total int) {
	totalPage := int(math.Ceil(float64(total) / float64(limit)))
	if totalPage == 0 {
		totalPage = 1
	}

	meta := &domain.Meta{
		Page:      page,
		Limit:     limit,
		TotalData: total,
		TotalPage: totalPage,
	}

	RespondJSON(w, http.StatusOK, domain.ResponseEnvelope{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
		Errors:  nil,
	})
}

func RespondError(w http.ResponseWriter, status int, message string, err interface{}) {
	RespondJSON(w, status, domain.ResponseEnvelope{
		Success: false,
		Message: message,
		Data:    nil,
		Errors:  err,
	})
}

func RespondBadRequest(w http.ResponseWriter, message string, err interface{}) {
	RespondError(w, http.StatusBadRequest, message, err)
}

func RespondInternalError(w http.ResponseWriter, err error) {
	RespondError(w, http.StatusInternalServerError, "Terjadi kesalahan internal pada server", err.Error())
}

func RespondNotFound(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusNotFound, message, nil)
}
