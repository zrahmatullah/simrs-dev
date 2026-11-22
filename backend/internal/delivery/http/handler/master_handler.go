package handler

import (
	"encoding/json"
	"net/http"

	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/usecase"
)

type MasterHandler struct {
	masterUsecase *usecase.MasterUsecase
}

func NewMasterHandler(masterUsecase *usecase.MasterUsecase) *MasterHandler {
	return &MasterHandler{masterUsecase: masterUsecase}
}

func (h *MasterHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.masterUsecase.GetRoles()
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar master role berhasil diambil", roles)
}

func (h *MasterHandler) GetPoliUnits(w http.ResponseWriter, r *http.Request) {
	units, err := h.masterUsecase.GetPoliUnits()
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar master unit/poli berhasil diambil", units)
}

func (h *MasterHandler) GetGuarantors(w http.ResponseWriter, r *http.Request) {
	guarantors, err := h.masterUsecase.GetGuarantors()
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar master penjamin berhasil diambil", guarantors)
}

func (h *MasterHandler) GetICD10(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	codes, err := h.masterUsecase.GetICD10(search)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar kode diagnosis ICD-10 berhasil diambil", codes)
}

func (h *MasterHandler) GetICD9(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	codes, err := h.masterUsecase.GetICD9(search)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar kode tindakan/prosedur ICD-9-CM berhasil diambil", codes)
}

func (h *MasterHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.masterUsecase.GetCategories()
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar master kategori obat & BHP berhasil diambil", cats)
}

func (h *MasterHandler) GetPromos(w http.ResponseWriter, r *http.Request) {
	promos, err := h.masterUsecase.GetPromos()
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar master promo berhasil diambil", promos)
}

func (h *MasterHandler) CreatePromo(w http.ResponseWriter, r *http.Request) {
	var req domain.CreatePromoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}

	promo, err := h.masterUsecase.CreatePromo(req)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Master promo berhasil ditambahkan", promo)
}

func (h *MasterHandler) GetSuppliers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	suppliers, err := h.masterUsecase.GetSuppliers(search)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar master supplier / PBF berhasil diambil", suppliers)
}

func (h *MasterHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}
	if req.Code == "" || req.Name == "" {
		RespondBadRequest(w, "Kode supplier dan nama supplier wajib diisi", nil)
		return
	}

	supplier, err := h.masterUsecase.CreateSupplier(req)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}

	RespondCreated(w, "Master supplier berhasil ditambahkan", supplier)
}

func (h *MasterHandler) CreatePoli(w http.ResponseWriter, r *http.Request) {
	var req domain.CreatePoliRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}
	if req.Code == "" || req.Name == "" {
		RespondBadRequest(w, "Kode poli dan nama poli wajib diisi", nil)
		return
	}

	poli, err := h.masterUsecase.CreatePoli(req)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondCreated(w, "Master unit/poli berhasil ditambahkan", poli)
}

func (h *MasterHandler) GetFaskes(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	faskes, err := h.masterUsecase.GetFaskes(search)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar master faskes/RS rujukan berhasil diambil", faskes)
}

func (h *MasterHandler) CreateFaskes(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateFaskesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}
	if req.Code == "" || req.Name == "" {
		RespondBadRequest(w, "Kode dan nama faskes/RS rujukan wajib diisi", nil)
		return
	}

	faskes, err := h.masterUsecase.CreateFaskes(req)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondCreated(w, "Master faskes/RS rujukan berhasil ditambahkan", faskes)
}

func (h *MasterHandler) GetSpesialis(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	spesialis, err := h.masterUsecase.GetSpesialis(search)
	if err != nil {
		RespondInternalError(w, err)
		return
	}
	RespondSuccess(w, "Daftar master dokter spesialis berhasil diambil", spesialis)
}

func (h *MasterHandler) CreateSpesialis(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateSpesialisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Format request body tidak valid", err.Error())
		return
	}
	if req.Code == "" || req.NamaDokter == "" || req.Spesialisasi == "" {
		RespondBadRequest(w, "Kode, nama dokter, dan bidang spesialisasi wajib diisi", nil)
		return
	}

	spesialis, err := h.masterUsecase.CreateSpesialis(req)
	if err != nil {
		RespondBadRequest(w, err.Error(), nil)
		return
	}
	RespondCreated(w, "Master dokter spesialis berhasil ditambahkan", spesialis)
}
