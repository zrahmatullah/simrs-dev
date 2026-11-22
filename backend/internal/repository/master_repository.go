package repository

import (
	"github.com/jmoiron/sqlx"
	"simrs-asoy/backend/internal/domain"
)

type MasterRepository struct {
	db *sqlx.DB
}

func NewMasterRepository(db *sqlx.DB) *MasterRepository {
	return &MasterRepository{db: db}
}

func (r *MasterRepository) GetRoles() ([]domain.Role, error) {
	var list []domain.Role
	err := r.db.Select(&list, "SELECT id, code, name, description, created_at, updated_at FROM roles ORDER BY id ASC")
	return list, err
}

func (r *MasterRepository) GetPoliUnits() ([]domain.PoliUnit, error) {
	var list []domain.PoliUnit
	err := r.db.Select(&list, "SELECT id, code, name, description, status_aktif, created_at FROM poli_units WHERE status_aktif = TRUE ORDER BY id ASC")
	return list, err
}

func (r *MasterRepository) GetGuarantors() ([]domain.Guarantor, error) {
	var list []domain.Guarantor
	err := r.db.Select(&list, "SELECT id, code, name, category, status_aktif, created_at FROM guarantors WHERE status_aktif = TRUE ORDER BY id ASC")
	return list, err
}

func (r *MasterRepository) GetICD10(search string) ([]domain.ICD10, error) {
	var list []domain.ICD10
	var err error
	if search != "" {
		query := `SELECT id, code, name_id, name_en, status_aktif FROM icd10_codes WHERE (code ILIKE $1 OR name_id ILIKE $1) LIMIT 20`
		err = r.db.Select(&list, query, "%"+search+"%")
	} else {
		err = r.db.Select(&list, "SELECT id, code, name_id, name_en, status_aktif FROM icd10_codes LIMIT 20")
	}
	return list, err
}

func (r *MasterRepository) GetICD9(search string) ([]domain.ICD9, error) {
	var list []domain.ICD9
	var err error
	if search != "" {
		query := `SELECT id, code, name_id, tariff, status_aktif FROM icd9_codes WHERE (code ILIKE $1 OR name_id ILIKE $1) LIMIT 20`
		err = r.db.Select(&list, query, "%"+search+"%")
	} else {
		err = r.db.Select(&list, "SELECT id, code, name_id, tariff, status_aktif FROM icd9_codes LIMIT 20")
	}
	return list, err
}

func (r *MasterRepository) GetCategories() ([]domain.InventoryCategory, error) {
	var list []domain.InventoryCategory
	err := r.db.Select(&list, "SELECT id, name, description FROM inventory_categories ORDER BY id ASC")
	return list, err
}

// Promos Master
func (r *MasterRepository) GetPromos() ([]domain.Promo, error) {
	var list []domain.Promo
	query := `
		SELECT id, code, name, discount_type, discount_value, min_transaction,
		       TO_CHAR(start_date, 'YYYY-MM-DD') AS start_date,
		       TO_CHAR(end_date, 'YYYY-MM-DD') AS end_date,
		       status_aktif, created_at, updated_at
		FROM promos
		ORDER BY id DESC
	`
	err := r.db.Select(&list, query)
	return list, err
}

func (r *MasterRepository) CreatePromo(p *domain.Promo) error {
	query := `
		INSERT INTO promos (code, name, discount_type, discount_value, min_transaction, start_date, end_date, status_aktif, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(query, p.Code, p.Name, p.DiscountType, p.DiscountValue, p.MinTransaction, p.StartDate, p.EndDate, p.StatusAktif).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *MasterRepository) GetPromoByCode(code string) (*domain.Promo, error) {
	var p domain.Promo
	query := `
		SELECT id, code, name, discount_type, discount_value, min_transaction,
		       TO_CHAR(start_date, 'YYYY-MM-DD') AS start_date,
		       TO_CHAR(end_date, 'YYYY-MM-DD') AS end_date,
		       status_aktif, created_at, updated_at
		FROM promos
		WHERE code = $1 AND status_aktif = TRUE AND CURRENT_DATE BETWEEN start_date AND end_date
		LIMIT 1
	`
	err := r.db.Get(&p, query, code)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Master Suppliers
func (r *MasterRepository) GetSuppliers(search string) ([]domain.Supplier, error) {
	var list []domain.Supplier
	var err error
	if search != "" {
		query := `
			SELECT id, code, name, COALESCE(pic_name, '') as pic_name, COALESCE(phone, '') as phone,
			       COALESCE(email, '') as email, COALESCE(address, '') as address, COALESCE(city, '') as city,
			       status_aktif, created_at
			FROM suppliers
			WHERE (code ILIKE $1 OR name ILIKE $1 OR city ILIKE $1 OR pic_name ILIKE $1)
			ORDER BY name ASC
		`
		err = r.db.Select(&list, query, "%"+search+"%")
	} else {
		query := `
			SELECT id, code, name, COALESCE(pic_name, '') as pic_name, COALESCE(phone, '') as phone,
			       COALESCE(email, '') as email, COALESCE(address, '') as address, COALESCE(city, '') as city,
			       status_aktif, created_at
			FROM suppliers
			ORDER BY name ASC
		`
		err = r.db.Select(&list, query)
	}
	return list, err
}

func (r *MasterRepository) CreateSupplier(s *domain.Supplier) error {
	query := `
		INSERT INTO suppliers (code, name, pic_name, phone, email, address, city, status_aktif, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, s.Code, s.Name, s.PicName, s.Phone, s.Email, s.Address, s.City, s.StatusAktif).
		Scan(&s.ID, &s.CreatedAt)
}

// Master Poli Unit
func (r *MasterRepository) CreatePoli(p *domain.PoliUnit) error {
	query := `
		INSERT INTO poli_units (code, name, description, status_aktif, created_at, updated_at)
		VALUES ($1, $2, $3, TRUE, NOW(), NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, p.Code, p.Name, p.Description).
		Scan(&p.ID, &p.CreatedAt)
}

// Master Faskes / RS Rujukan
func (r *MasterRepository) GetFaskes(search string) ([]domain.FaskesRujukan, error) {
	var list []domain.FaskesRujukan
	var err error
	if search != "" {
		query := `
			SELECT id, code, name, tipe, COALESCE(alamat, '') AS alamat,
			       COALESCE(kota, '') AS kota, COALESCE(no_telp, '') AS no_telp,
			       status_aktif, created_at, updated_at
			FROM master_faskes_rujukan
			WHERE (code ILIKE $1 OR name ILIKE $1 OR kota ILIKE $1 OR tipe ILIKE $1)
			ORDER BY name ASC
		`
		err = r.db.Select(&list, query, "%"+search+"%")
	} else {
		query := `
			SELECT id, code, name, tipe, COALESCE(alamat, '') AS alamat,
			       COALESCE(kota, '') AS kota, COALESCE(no_telp, '') AS no_telp,
			       status_aktif, created_at, updated_at
			FROM master_faskes_rujukan
			ORDER BY name ASC
		`
		err = r.db.Select(&list, query)
	}
	return list, err
}

func (r *MasterRepository) CreateFaskes(f *domain.FaskesRujukan) error {
	query := `
		INSERT INTO master_faskes_rujukan (code, name, tipe, alamat, kota, no_telp, status_aktif, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(query, f.Code, f.Name, f.Tipe, f.Alamat, f.Kota, f.NoTelp, f.StatusAktif).
		Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt)
}

// Master Dokter Spesialis & Subspesialis Rujukan
func (r *MasterRepository) GetSpesialis(search string) ([]domain.MasterSpesialis, error) {
	var list []domain.MasterSpesialis
	var err error
	if search != "" {
		query := `
			SELECT id, code, nama_dokter, spesialisasi, COALESCE(rumah_sakit, '') AS rumah_sakit,
			       status_aktif, created_at, updated_at
			FROM master_spesialis
			WHERE (code ILIKE $1 OR nama_dokter ILIKE $1 OR spesialisasi ILIKE $1 OR rumah_sakit ILIKE $1)
			ORDER BY nama_dokter ASC
		`
		err = r.db.Select(&list, query, "%"+search+"%")
	} else {
		query := `
			SELECT id, code, nama_dokter, spesialisasi, COALESCE(rumah_sakit, '') AS rumah_sakit,
			       status_aktif, created_at, updated_at
			FROM master_spesialis
			ORDER BY nama_dokter ASC
		`
		err = r.db.Select(&list, query)
	}
	return list, err
}

func (r *MasterRepository) CreateSpesialis(s *domain.MasterSpesialis) error {
	query := `
		INSERT INTO master_spesialis (code, nama_dokter, spesialisasi, rumah_sakit, status_aktif, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(query, s.Code, s.NamaDokter, s.Spesialisasi, s.RumahSakit, s.StatusAktif).
		Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}
