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

type BillingHandler struct {
	billingUsecase *usecase.BillingUsecase
}

func NewBillingHandler(billingUsecase *usecase.BillingUsecase) *BillingHandler {
	return &BillingHandler{billingUsecase: billingUsecase}
}

func (h *BillingHandler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	list, total, err := h.billingUsecase.GetInvoices(status, search, page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondMeta(w, "Daftar tagihan / invoice kasir berhasil diambil", list, page, limit, total)
}

func (h *BillingHandler) GetInvoiceByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	inv, err := h.billingUsecase.GetInvoiceByID(id)
	if err != nil {
		RespondNotFound(w, "Invoice tidak ditemukan")
		return
	}
	RespondSuccess(w, "Data invoice berhasil diambil", inv)
}

func (h *BillingHandler) GetInvoiceByVisit(w http.ResponseWriter, r *http.Request) {
	visitID := chi.URLParam(r, "visit_id")
	inv, err := h.billingUsecase.GetInvoiceByVisit(visitID)
	if err != nil {
		RespondNotFound(w, "Invoice untuk kunjungan ini belum ada")
		return
	}
	RespondSuccess(w, "Data invoice kunjungan berhasil diambil", inv)
}

func (h *BillingHandler) GenerateBillingFromVisit(w http.ResponseWriter, r *http.Request) {
	visitID := chi.URLParam(r, "visit_id")
	claims := middleware.GetCurrentUser(r)

	inv, err := h.billingUsecase.GenerateBillingFromVisit(visitID, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Rincian tagihan pelayanan berhasil dikalkulasi", inv)
}

func (h *BillingHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	inv, err := h.billingUsecase.CreateInvoice(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Invoice tagihan kasir berhasil dibuat", inv)
}

func (h *BillingHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	invoiceID := chi.URLParam(r, "id")
	var req domain.ProcessPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	inv, err := h.billingUsecase.ProcessPayment(invoiceID, req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Pembayaran berhasil diproses dan kuitansi diterbitkan", inv)
}

// Shifts
func (h *BillingHandler) GetActiveShift(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetCurrentUser(r)
	shift, err := h.billingUsecase.GetActiveShift(claims.UserID)
	if err != nil {
		RespondSuccess(w, "Tidak ada shift kasir yang sedang aktif", nil)
		return
	}
	RespondSuccess(w, "Data shift aktif berhasil diambil", shift)
}

func (h *BillingHandler) OpenShift(w http.ResponseWriter, r *http.Request) {
	var req domain.OpenShiftRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	claims := middleware.GetCurrentUser(r)
	shift, err := h.billingUsecase.OpenShift(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Shift kasir berhasil dibuka", shift)
}

func (h *BillingHandler) CloseShift(w http.ResponseWriter, r *http.Request) {
	shiftID := chi.URLParam(r, "shift_id")
	var req domain.CloseShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	err := h.billingUsecase.CloseShift(shiftID, req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Shift kasir berhasil ditutup dan laporan rekonsiliasi kasir dicatat", nil)
}

func (h *BillingHandler) GetShiftHistory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetCurrentUser(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}

	kasirID := ""
	if claims.RoleCode == "CASHIER" {
		kasirID = claims.UserID
	}

	list, err := h.billingUsecase.GetShiftHistory(kasirID, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondSuccess(w, "Riwayat shift kasir berhasil diambil", list)
}
