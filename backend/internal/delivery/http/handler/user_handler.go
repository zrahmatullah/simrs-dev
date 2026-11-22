package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/middleware"
	"simrs-asoy/backend/internal/usecase"
)

type UserHandler struct {
	userUsecase *usecase.UserUsecase
}

func NewUserHandler(userUsecase *usecase.UserUsecase) *UserHandler {
	return &UserHandler{userUsecase: userUsecase}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	roleID, _ := strconv.Atoi(r.URL.Query().Get("role_id"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	users, total, err := h.userUsecase.GetUsers(search, roleID, page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondMeta(w, "Daftar pengguna berhasil diambil", users, page, limit, total)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, err := h.userUsecase.GetUserByID(id)
	if err != nil {
		RespondNotFound(w, "Pengguna tidak ditemukan")
		return
	}
	RespondSuccess(w, "Data pengguna berhasil diambil", user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	user, err := h.userUsecase.CreateUser(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Pengguna baru berhasil ditambahkan", user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req domain.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	err := h.userUsecase.UpdateUser(id, req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Data pengguna berhasil diperbarui", nil)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetCurrentUser(r)

	if claims.UserID == id {
		RespondBadRequest(w, "Anda tidak dapat menghapus akun Anda sendiri", nil)
		return
	}

	err := h.userUsecase.DeleteUser(id, claims.UserID, claims.Username)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondSuccess(w, "Pengguna berhasil dihapus", nil)
}
