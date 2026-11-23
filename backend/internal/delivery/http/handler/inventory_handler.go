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

type InventoryHandler struct {
	invUsecase *usecase.InventoryUsecase
}

func NewInventoryHandler(invUsecase *usecase.InventoryUsecase) *InventoryHandler {
	return &InventoryHandler{invUsecase: invUsecase}
}

func (h *InventoryHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	categoryID, _ := strconv.Atoi(r.URL.Query().Get("category_id"))
	isLowStock := r.URL.Query().Get("low_stock") == "true"
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	items, total, err := h.invUsecase.GetItems(search, categoryID, isLowStock, page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondMeta(w, "Daftar stok barang medis dan BHP berhasil diambil", items, page, limit, total)
}

func (h *InventoryHandler) GetItemByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	item, err := h.invUsecase.GetItemByID(id)
	if err != nil {
		RespondNotFound(w, "Barang tidak ditemukan")
		return
	}
	RespondSuccess(w, "Data barang berhasil diambil", item)
}

func (h *InventoryHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var item domain.InventoryItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	created, err := h.invUsecase.CreateItem(&item, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Item obat / alkes baru berhasil ditambahkan", created)
}

func (h *InventoryHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	var item domain.InventoryItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}
	item.ID = id

	claims := middleware.GetCurrentUser(r)
	err := h.invUsecase.UpdateItem(&item, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Data item berhasil diperbarui", nil)
}

// Purchase Orders
func (h *InventoryHandler) GetPOs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	list, total, err := h.invUsecase.GetPOs(status, page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondMeta(w, "Daftar Purchase Order berhasil diambil", list, page, limit, total)
}

func (h *InventoryHandler) CreatePO(w http.ResponseWriter, r *http.Request) {
	var req domain.CreatePORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	po, err := h.invUsecase.CreatePO(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Purchase Order pengadaan berhasil diterbitkan", po)
}

func (h *InventoryHandler) ReceivePO(w http.ResponseWriter, r *http.Request) {
	poID := chi.URLParam(r, "id")
	claims := middleware.GetCurrentUser(r)

	err := h.invUsecase.ReceivePO(poID, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondSuccess(w, "Penerimaan barang dari supplier berhasil diverifikasi dan stok gudang utama bertambah", nil)
}

// Mutations
func (h *InventoryHandler) CreateStockMutation(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateStockMutationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	m, err := h.invUsecase.CreateStockMutation(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Mutasi / penyesuaian stok berhasil dicatat", m)
}

func (h *InventoryHandler) GetStockMutations(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.Atoi(r.URL.Query().Get("item_id"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}

	list, err := h.invUsecase.GetStockMutations(itemID, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}

	RespondSuccess(w, "Riwayat mutasi kartu stok berhasil diambil", list)
}

// Purchase Requisitions (PR)
func (h *InventoryHandler) GetPRs(w http.ResponseWriter, r *http.Request) {
	list, err := h.invUsecase.GetPRs()
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar Purchase Requisition berhasil diambil", list)
}

func (h *InventoryHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req domain.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	pr, err := h.invUsecase.CreatePR(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondCreated(w, "Purchase Requisition berhasil diterbitkan", pr)
}

// Goods Receipt Notes (GRN)
func (h *InventoryHandler) GetGRNs(w http.ResponseWriter, r *http.Request) {
	list, err := h.invUsecase.GetGRNs()
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar Goods Receipt Note berhasil diambil", list)
}

func (h *InventoryHandler) CreateGRN(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateGRNRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	grn, err := h.invUsecase.CreateGRN(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondCreated(w, "Penerimaan barang (GRN) berhasil dicatat dan stok gudang bertambah", grn)
}

// -------------------------------------------------------------
// Stock Opname & Pemusnahan Obat Handlers (Fase 1)
// -------------------------------------------------------------

func (h *InventoryHandler) CreateStockOpname(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateStockOpnameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	so, err := h.invUsecase.CreateStockOpname(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondCreated(w, "Stock opname berhasil disimpan dan penyesuaian stok telah diterapkan", so)
}

func (h *InventoryHandler) GetStockOpnames(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	list, total, err := h.invUsecase.GetStockOpnames(page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondMeta(w, "Data stock opname berhasil diambil", list, page, limit, total)
}

func (h *InventoryHandler) CreateMedicineDisposal(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateMedicineDisposalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	claims := middleware.GetCurrentUser(r)
	disp, err := h.invUsecase.CreateMedicineDisposal(req, claims.UserID, claims.Username)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondCreated(w, "Berita acara pemusnahan obat berhasil diterbitkan dan stok gudang terpotong", disp)
}

func (h *InventoryHandler) GetMedicineDisposals(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	list, total, err := h.invUsecase.GetMedicineDisposals(page, limit)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondMeta(w, "Data pemusnahan obat berhasil diambil", list, page, limit, total)
}
