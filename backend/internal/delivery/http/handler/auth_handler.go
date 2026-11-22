package handler

import (
	"encoding/json"
	"net/http"

	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/middleware"
	"simrs-asoy/backend/internal/usecase"
)

type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
	userUsecase *usecase.UserUsecase
}

func NewAuthHandler(authUsecase *usecase.AuthUsecase, userUsecase *usecase.UserUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
		userUsecase: userUsecase,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	resp, err := h.authUsecase.Login(req, ipAddress, userAgent)
	if err != nil {
		RespondError(w, http.StatusUnauthorized, err.Error(), map[string]string{"login": err.Error()})
		return
	}

	RespondSuccess(w, "Login berhasil", resp)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req domain.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	resp, err := h.authUsecase.RefreshToken(req.RefreshToken, ipAddress, userAgent)
	if err != nil {
		RespondError(w, http.StatusUnauthorized, err.Error(), map[string]string{"refresh": err.Error()})
		return
	}

	RespondSuccess(w, "Token berhasil diperbarui (Token Rotated)", resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req domain.LogoutRequest
	_ = json.NewDecoder(r.Body).Decode(&req) // Optional body

	var userID string
	if claims := middleware.GetCurrentUser(r); claims != nil {
		userID = claims.UserID
	}

	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	_ = h.authUsecase.Logout(req.RefreshToken, userID, ipAddress, userAgent)

	RespondSuccess(w, "Logout berhasil. Sesi telah diakhiri.", nil)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetCurrentUser(r)
	if claims == nil {
		RespondError(w, http.StatusUnauthorized, "Pengguna belum diautentikasi", nil)
		return
	}

	user, err := h.userUsecase.GetUserByID(claims.UserID)
	if err != nil {
		RespondNotFound(w, "Pengguna tidak ditemukan")
		return
	}

	RespondSuccess(w, "Data profil pengguna aktif berhasil diambil", user)
}
