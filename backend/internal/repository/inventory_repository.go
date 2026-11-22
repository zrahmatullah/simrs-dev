package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"simrs-asoy/backend/internal/domain"
)

type InventoryRepository struct {
	db *sqlx.DB
}

func NewInventoryRepository(db *sqlx.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) GeneratePONumber() (string, error) {
	dateStr := time.Now().Format("200601")
	var count int
	query := `SELECT COUNT(*) FROM purchase_orders WHERE po_number LIKE $1`
	err := r.db.Get(&count, query, fmt.Sprintf("PO-%s-%%", dateStr))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PO-%s-%03d", dateStr, count+1), nil
}

func (r *InventoryRepository) FindItems(search string, categoryID int, isLowStock bool, page, limit int) ([]domain.InventoryItem, int, error) {
	offset := (page - 1) * limit
	whereClauses := []string{"ii.status_aktif = TRUE"}
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(ii.item_code ILIKE $%d OR ii.item_name ILIKE $%d OR ii.generic_name ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if categoryID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("ii.category_id = $%d", argIndex))
		args = append(args, categoryID)
		argIndex++
	}

	if isLowStock {
		whereClauses = append(whereClauses, "ii.total_stock <= ii.minimum_stock")
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM inventory_items ii WHERE %s`, whereSQL)
	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT ii.id, ii.category_id, COALESCE(ic.name, '') AS category_name,
		       ii.item_code, ii.item_name, COALESCE(ii.generic_name, '') AS generic_name, ii.unit,
		       ii.buy_price, ii.sell_price, ii.minimum_stock, ii.total_stock,
		       ii.depo_stock, ii.is_medicine, COALESCE(ii.rack_location, '') AS rack_location, ii.status_aktif, ii.created_at
		FROM inventory_items ii
		LEFT JOIN inventory_categories ic ON ii.category_id = ic.id
		WHERE %s
		ORDER BY ii.item_name ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, limit, offset)
	var items []domain.InventoryItem
	err = r.db.Select(&items, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *InventoryRepository) GetItemByID(id int) (*domain.InventoryItem, error) {
	query := `
		SELECT ii.id, ii.category_id, COALESCE(ic.name, '') AS category_name,
		       ii.item_code, ii.item_name, COALESCE(ii.generic_name, '') AS generic_name, ii.unit,
		       ii.buy_price, ii.sell_price, ii.minimum_stock, ii.total_stock,
		       ii.depo_stock, ii.is_medicine, COALESCE(ii.rack_location, '') AS rack_location, ii.status_aktif, ii.created_at
		FROM inventory_items ii
		LEFT JOIN inventory_categories ic ON ii.category_id = ic.id
		WHERE ii.id = $1
	`
	var it domain.InventoryItem
	err := r.db.Get(&it, query, id)
	if err != nil {
		return nil, err
	}
	return &it, nil
}

func (r *InventoryRepository) CreateItem(it *domain.InventoryItem) error {
	query := `
		INSERT INTO inventory_items (category_id, item_code, item_name, generic_name, unit,
		                             buy_price, sell_price, minimum_stock, total_stock, depo_stock,
		                             is_medicine, rack_location, status_aktif, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, true, NOW(), NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, it.CategoryID, it.ItemCode, it.ItemName, it.GenericName, it.Unit,
		it.BuyPrice, it.SellPrice, it.MinimumStock, it.TotalStock, it.DepoStock,
		it.IsMedicine, it.RackLocation).Scan(&it.ID, &it.CreatedAt)
}

func (r *InventoryRepository) UpdateItem(it *domain.InventoryItem) error {
	query := `
		UPDATE inventory_items
		SET category_id = $1, item_name = $2, generic_name = $3, unit = $4,
		    buy_price = $5, sell_price = $6, minimum_stock = $7,
		    is_medicine = $8, rack_location = $9, status_aktif = $10, updated_at = NOW()
		WHERE id = $11
	`
	_, err := r.db.Exec(query, it.CategoryID, it.ItemName, it.GenericName, it.Unit,
		it.BuyPrice, it.SellPrice, it.MinimumStock, it.IsMedicine, it.RackLocation, it.StatusAktif, it.ID)
	return err
}

// Purchase Orders
func (r *InventoryRepository) FindPOs(status string, page, limit int) ([]domain.PurchaseOrder, int, error) {
	offset := (page - 1) * limit
	whereClauses := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM purchase_orders WHERE %s`, whereSQL)
	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, po_number, supplier_id, supplier_name, TO_CHAR(order_date, 'YYYY-MM-DD') AS order_date,
		       TO_CHAR(expected_date, 'YYYY-MM-DD') AS expected_date, status, total_amount, COALESCE(catatan, '') AS catatan, created_at
		FROM purchase_orders
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, limit, offset)
	var list []domain.PurchaseOrder
	err = r.db.Select(&list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	// Fetch items
	for i := range list {
		itemsQuery := `
			SELECT poi.id, poi.po_id, poi.item_id, ii.item_name, ii.unit,
			       poi.qty, poi.unit_price, poi.subtotal, poi.created_at
			FROM purchase_order_items poi
			JOIN inventory_items ii ON poi.item_id = ii.id
			WHERE poi.po_id = $1
		`
		var items []domain.PurchaseOrderItem
		_ = r.db.Select(&items, itemsQuery, list[i].ID)
		list[i].Items = items
	}

	return list, total, nil
}

func (r *InventoryRepository) CreatePO(po *domain.PurchaseOrder, items []domain.POItemDTO) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if po.PONumber == "" {
		noPO, err := r.GeneratePONumber()
		if err != nil {
			return err
		}
		po.PONumber = noPO
	}

	var totalAmount float64
	for _, it := range items {
		totalAmount += it.UnitPrice * float64(it.Qty)
	}
	po.TotalAmount = totalAmount

	query := `
		INSERT INTO purchase_orders (po_number, supplier_id, supplier_name, order_date, expected_date, status, total_amount, catatan, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_DATE, $4, 'SUBMITTED', $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at
	`
	err = tx.QueryRow(query, po.PONumber, po.SupplierID, po.SupplierName, po.ExpectedDate, po.TotalAmount, po.Catatan, po.CreatedBy).
		Scan(&po.ID, &po.CreatedAt)
	if err != nil {
		return err
	}

	for _, it := range items {
		subtotal := it.UnitPrice * float64(it.Qty)
		itemQuery := `
			INSERT INTO purchase_order_items (po_id, item_id, qty, unit_price, subtotal, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
		`
		_, err = tx.Exec(itemQuery, po.ID, it.ItemID, it.Qty, it.UnitPrice, subtotal)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *InventoryRepository) ReceivePO(poID, receivedBy string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get items
	var items []domain.PurchaseOrderItem
	err = tx.Select(&items, "SELECT item_id, qty, unit_price FROM purchase_order_items WHERE po_id = $1", poID)
	if err != nil {
		return err
	}

	for _, it := range items {
		// Increase warehouse stock
		_, err = tx.Exec("UPDATE inventory_items SET total_stock = total_stock + $1, updated_at = NOW() WHERE id = $2", it.Qty, it.ItemID)
		if err != nil {
			return err
		}

		// Log mutation
		_, _ = tx.Exec(`
			INSERT INTO stock_mutations (item_id, transaction_type, qty, source_depo, dest_depo, notes, created_by, created_at)
			VALUES ($1, 'IN_PO', $2, 'SUPPLIER', 'GUDANG_UTAMA', 'Penerimaan Purchase Order', $3, NOW())
		`, it.ItemID, it.Qty, receivedBy)
	}

	// Mark PO as RECEIVED
	_, err = tx.Exec("UPDATE purchase_orders SET status = 'RECEIVED', updated_at = NOW() WHERE id = $1", poID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Stock Mutations (Transfer to Depo Farmasi, Adjustment)
func (r *InventoryRepository) CreateStockMutation(m *domain.StockMutation, userID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if m.TransactionType == "MUTASI_DEPO" {
		// Deduct warehouse stock, add to depo stock
		_, err = tx.Exec("UPDATE inventory_items SET total_stock = GREATEST(0, total_stock - $1), depo_stock = depo_stock + $1, updated_at = NOW() WHERE id = $2", m.Qty, m.ItemID)
		if err != nil {
			return err
		}
	} else if m.TransactionType == "ADJUSTMENT" {
		// Set total stock
		_, err = tx.Exec("UPDATE inventory_items SET total_stock = $1, updated_at = NOW() WHERE id = $2", m.Qty, m.ItemID)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO stock_mutations (item_id, transaction_type, qty, batch_number, expiry_date, source_depo, dest_depo, notes, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id, created_at
	`
	err = tx.QueryRow(query, m.ItemID, m.TransactionType, m.Qty, m.BatchNumber, m.ExpiryDate, m.SourceDepo, m.DestDepo, m.Notes, userID).
		Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *InventoryRepository) FindStockMutations(itemID int, limit int) ([]domain.StockMutation, error) {
	whereSQL := "1=1"
	args := []interface{}{}
	if itemID > 0 {
		whereSQL = "sm.item_id = $1"
		args = append(args, itemID)
	}

	query := fmt.Sprintf(`
		SELECT sm.id, sm.item_id, ii.item_name, sm.transaction_type, sm.qty,
		       COALESCE(sm.batch_number, '') AS batch_number,
		       COALESCE(TO_CHAR(sm.expiry_date, 'YYYY-MM-DD'), '') AS expiry_date,
		       COALESCE(sm.source_depo, '') AS source_depo,
		       COALESCE(sm.dest_depo, '') AS dest_depo,
		       COALESCE(sm.notes, '') AS notes,
		       sm.created_at
		FROM stock_mutations sm
		JOIN inventory_items ii ON sm.item_id = ii.id
		WHERE %s
		ORDER BY sm.created_at DESC
		LIMIT %d
	`, whereSQL, limit)

	var list []domain.StockMutation
	err := r.db.Select(&list, query, args...)
	return list, err
}

// Purchase Requisition (PR)
func (r *InventoryRepository) GeneratePRNumber() (string, error) {
	dateStr := time.Now().Format("20060102")
	var count int
	query := `SELECT COUNT(*) FROM purchase_requisitions WHERE pr_number LIKE $1`
	err := r.db.Get(&count, query, fmt.Sprintf("PR-%s-%%", dateStr))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PR-%s-%03d", dateStr, count+1), nil
}

func (r *InventoryRepository) CreatePR(pr *domain.PurchaseRequisition, items []domain.PRItemDTO) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if pr.PRNumber == "" {
		noPR, err := r.GeneratePRNumber()
		if err != nil {
			return err
		}
		pr.PRNumber = noPR
	}

	var totalEstimated float64
	for _, it := range items {
		totalEstimated += it.UnitPrice * float64(it.Qty)
	}
	pr.TotalEstimated = totalEstimated

	query := `
		INSERT INTO purchase_requisitions (pr_number, requested_by, request_date, department, status, total_estimated, notes, created_at, updated_at)
		VALUES ($1, $2, CURRENT_DATE, $3, 'SUBMITTED', $4, $5, NOW(), NOW())
		RETURNING id, created_at
	`
	err = tx.QueryRow(query, pr.PRNumber, pr.RequestedBy, pr.Department, pr.TotalEstimated, pr.Notes).
		Scan(&pr.ID, &pr.CreatedAt)
	if err != nil {
		return err
	}

	for _, it := range items {
		subtotal := it.UnitPrice * float64(it.Qty)
		itemQuery := `
			INSERT INTO purchase_requisition_items (pr_id, item_id, qty, unit_price, subtotal)
			VALUES ($1, $2, $3, $4, $5)
		`
		_, err = tx.Exec(itemQuery, pr.ID, it.ItemID, it.Qty, it.UnitPrice, subtotal)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *InventoryRepository) FindPRs() ([]domain.PurchaseRequisition, error) {
	var prs []domain.PurchaseRequisition
	query := `
		SELECT pr.id, pr.pr_number, pr.requested_by, pr.request_date, pr.department, pr.status,
		       pr.total_estimated, COALESCE(pr.notes, '') as notes, pr.created_at, pr.updated_at,
		       COALESCE(u.full_name, 'Staff') as requester_name
		FROM purchase_requisitions pr
		LEFT JOIN users u ON pr.requested_by = u.id
		ORDER BY pr.created_at DESC
	`
	err := r.db.Select(&prs, query)
	return prs, err
}

// Goods Receipt Note (GRN)
func (r *InventoryRepository) GenerateGRNNumber() (string, error) {
	dateStr := time.Now().Format("20060102")
	var count int
	query := `SELECT COUNT(*) FROM goods_receipt_notes WHERE grn_number LIKE $1`
	err := r.db.Get(&count, query, fmt.Sprintf("GRN-%s-%%", dateStr))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("GRN-%s-%03d", dateStr, count+1), nil
}

func (r *InventoryRepository) CreateGRN(grn *domain.GoodsReceiptNote, receiverID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if grn.GRNNumber == "" {
		noGRN, err := r.GenerateGRNNumber()
		if err != nil {
			return err
		}
		grn.GRNNumber = noGRN
	}

	query := `
		INSERT INTO goods_receipt_notes (grn_number, po_id, received_date, receiver_name, delivery_order_number, status, notes, created_at)
		VALUES ($1, $2, CURRENT_DATE, $3, $4, 'RECEIVED', $5, NOW())
		RETURNING id, created_at
	`
	err = tx.QueryRow(query, grn.GRNNumber, grn.POID, grn.ReceiverName, grn.DeliveryOrderNumber, grn.Notes).
		Scan(&grn.ID, &grn.CreatedAt)
	if err != nil {
		return err
	}

	// Update inventory items and create mutation from linked PO
	var items []domain.PurchaseOrderItem
	err = tx.Select(&items, "SELECT item_id, qty, unit_price FROM purchase_order_items WHERE po_id = $1", grn.POID)
	if err == nil {
		for _, it := range items {
			_, err = tx.Exec("UPDATE inventory_items SET total_stock = total_stock + $1, updated_at = NOW() WHERE id = $2", it.Qty, it.ItemID)
			if err != nil {
				return err
			}

			_, _ = tx.Exec(`
				INSERT INTO stock_mutations (item_id, transaction_type, qty, source_depo, dest_depo, notes, created_by, created_at)
				VALUES ($1, 'IN_GRN', $2, 'SUPPLIER', 'GUDANG_UTAMA', $3, $4, NOW())
			`, it.ItemID, it.Qty, "Penerimaan GRN: "+grn.GRNNumber, receiverID)
		}
	}

	// Update PO status to RECEIVED
	_, _ = tx.Exec("UPDATE purchase_orders SET status = 'RECEIVED', updated_at = NOW() WHERE id = $1", grn.POID)

	return tx.Commit()
}

func (r *InventoryRepository) FindGRNs() ([]domain.GoodsReceiptNote, error) {
	var grns []domain.GoodsReceiptNote
	query := `
		SELECT grn.id, grn.grn_number, grn.po_id, grn.received_date, grn.receiver_name,
		       COALESCE(grn.delivery_order_number, '') as delivery_order_number,
		       grn.status, COALESCE(grn.notes, '') as notes, grn.created_at,
		       COALESCE(po.po_number, '') as po_number,
		       COALESCE(po.supplier_name, '') as supplier_name
		FROM goods_receipt_notes grn
		LEFT JOIN purchase_orders po ON grn.po_id = po.id
		ORDER BY grn.created_at DESC
	`
	err := r.db.Select(&grns, query)
	return grns, err
}
