package usecase

import (
	"errors"
	"fmt"

	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/repository"
	"simrs-asoy/backend/internal/security"
)

type InventoryUsecase struct {
	invRepo   *repository.InventoryRepository
	auditRepo *repository.AuditRepository
}

func NewInventoryUsecase(invRepo *repository.InventoryRepository, auditRepo *repository.AuditRepository) *InventoryUsecase {
	return &InventoryUsecase{invRepo: invRepo, auditRepo: auditRepo}
}

func (u *InventoryUsecase) GetItems(search string, categoryID int, isLowStock bool, page, limit int) ([]domain.InventoryItem, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.invRepo.FindItems(search, categoryID, isLowStock, page, limit)
}

func (u *InventoryUsecase) GetItemByID(id int) (*domain.InventoryItem, error) {
	return u.invRepo.GetItemByID(id)
}

func (u *InventoryUsecase) CreateItem(item *domain.InventoryItem, currentUserID, currentUsername string) (*domain.InventoryItem, error) {
	if item.ItemName == "" || item.Unit == "" {
		return nil, errors.New("nama barang dan satuan wajib diisi")
	}

	err := u.invRepo.CreateItem(item)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CREATE_ITEM",
		EntityType: "inventory_item",
		EntityID:   item.ItemCode,
		Details:    `{"item_name": "` + item.ItemName + `"}`,
	})

	return item, nil
}

func (u *InventoryUsecase) UpdateItem(item *domain.InventoryItem, currentUserID, currentUsername string) error {
	err := u.invRepo.UpdateItem(item)
	if err != nil {
		return err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "UPDATE_ITEM",
		EntityType: "inventory_item",
		EntityID:   item.ItemCode,
		Details:    `{"item_name": "` + item.ItemName + `"}`,
	})

	return nil
}

// Purchase Orders
func (u *InventoryUsecase) GetPOs(status string, page, limit int) ([]domain.PurchaseOrder, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.invRepo.FindPOs(status, page, limit)
}

func (u *InventoryUsecase) CreatePO(req domain.CreatePORequest, currentUserID, currentUsername string) (*domain.PurchaseOrder, error) {
	if req.SupplierName == "" || len(req.Items) == 0 {
		return nil, errors.New("nama supplier dan minimal satu item barang wajib diisi")
	}

	// OWASP: Validate PO Items Qty & UnitPrice
	for i, item := range req.Items {
		if err := security.ValidateQuantity(item.Qty, 100000); err != nil {
			return nil, fmt.Errorf("item PO ke-%d: %w", i+1, err)
		}
		if err := security.ValidatePrice(item.UnitPrice, 1_000_000_000); err != nil {
			return nil, fmt.Errorf("harga item PO ke-%d: %w", i+1, err)
		}
	}

	po := &domain.PurchaseOrder{
		SupplierID:   req.SupplierID,
		SupplierName: security.SanitizePlainText(req.SupplierName),
		OrderDate:    req.OrderDate,
		ExpectedDate: req.ExpectedDate,
		Catatan:      security.SanitizeText(req.Catatan),
		CreatedBy:    &currentUserID,
	}

	err := u.invRepo.CreatePO(po, req.Items)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CREATE_PO",
		EntityType: "purchase_order",
		EntityID:   po.ID,
		Details:    fmt.Sprintf(`{"po_number": "%s", "supplier": "%s"}`, po.PONumber, po.SupplierName),
	})

	return po, nil
}

func (u *InventoryUsecase) ReceivePO(poID, currentUserID, currentUsername string) error {
	err := u.invRepo.ReceivePO(poID, currentUserID)
	if err != nil {
		return err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "RECEIVE_PO",
		EntityType: "purchase_order",
		EntityID:   poID,
		Details:    `{"message": "Barang PO diterima dan stok gudang utama bertambah"}`,
	})

	return nil
}

// Stock Mutations
func (u *InventoryUsecase) CreateStockMutation(req domain.CreateStockMutationRequest, currentUserID, currentUsername string) (*domain.StockMutation, error) {
	if req.ItemID == 0 || req.TransactionType == "" {
		return nil, errors.New("item barang dan jenis transaksi wajib diisi")
	}

	// OWASP: Validate mutation quantity positive and within upper bound
	if err := security.ValidateQuantity(req.Qty, 100000); err != nil {
		return nil, err
	}

	m := &domain.StockMutation{
		ItemID:          req.ItemID,
		TransactionType: req.TransactionType,
		Qty:             req.Qty,
		BatchNumber:     security.SanitizePlainText(req.BatchNumber),
		ExpiryDate:      req.ExpiryDate,
		SourceDepo:      security.SanitizePlainText(req.SourceDepo),
		DestDepo:        security.SanitizePlainText(req.DestDepo),
		Notes:           security.SanitizeText(req.Notes),
	}

	err := u.invRepo.CreateStockMutation(m, currentUserID)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "MUTATION",
		EntityType: "stock_mutation",
		EntityID:   m.ID,
		Details:    fmt.Sprintf(`{"type": "%s", "qty": %d}`, m.TransactionType, m.Qty),
	})

	return m, nil
}

func (u *InventoryUsecase) GetStockMutations(itemID, limit int) ([]domain.StockMutation, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.invRepo.FindStockMutations(itemID, limit)
}

// Purchase Requisitions (PR)
func (u *InventoryUsecase) CreatePR(req domain.CreatePRRequest, currentUserID, currentUsername string) (*domain.PurchaseRequisition, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("minimal satu item barang wajib diisi dalam PR")
	}

	dept := req.Department
	if dept == "" {
		dept = "FARMASI"
	}

	pr := &domain.PurchaseRequisition{
		RequestedBy: &currentUserID,
		Department:  dept,
		Notes:       req.Notes,
	}

	err := u.invRepo.CreatePR(pr, req.Items)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CREATE_PR",
		EntityType: "purchase_requisition",
		EntityID:   pr.ID,
		Details:    `{"pr_number": "` + pr.PRNumber + `", "dept": "` + pr.Department + `"}`,
	})

	return pr, nil
}

func (u *InventoryUsecase) GetPRs() ([]domain.PurchaseRequisition, error) {
	return u.invRepo.FindPRs()
}

// Goods Receipt Notes (GRN)
func (u *InventoryUsecase) CreateGRN(req domain.CreateGRNRequest, currentUserID, currentUsername string) (*domain.GoodsReceiptNote, error) {
	if req.POID == "" || req.ReceiverName == "" {
		return nil, errors.New("PO ID dan nama penerima wajib diisi dalam GRN")
	}

	grn := &domain.GoodsReceiptNote{
		POID:                req.POID,
		ReceiverName:        req.ReceiverName,
		DeliveryOrderNumber: req.DeliveryOrderNumber,
		Notes:               req.Notes,
	}

	err := u.invRepo.CreateGRN(grn, currentUserID)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CREATE_GRN",
		EntityType: "goods_receipt_note",
		EntityID:   grn.ID,
		Details:    `{"grn_number": "` + grn.GRNNumber + `", "po_id": "` + grn.POID + `"}`,
	})

	return grn, nil
}

func (u *InventoryUsecase) GetGRNs() ([]domain.GoodsReceiptNote, error) {
	return u.invRepo.FindGRNs()
}
