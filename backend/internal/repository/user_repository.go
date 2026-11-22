package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"simrs-asoy/backend/internal/domain"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(username string) (*domain.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, 
		       COALESCE(u.nip, '') AS nip, COALESCE(u.phone, '') AS phone, 
		       u.role_id, r.code AS role_code, r.name AS role_name, 
		       u.poli_id, COALESCE(p.name, '') AS poli_name, u.status_aktif, u.created_at, u.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		LEFT JOIN poli_units p ON u.poli_id = p.id
		WHERE u.username = $1 AND u.deleted_at IS NULL
	`
	var user domain.User
	err := r.db.Get(&user, query, username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id string) (*domain.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, 
		       COALESCE(u.nip, '') AS nip, COALESCE(u.phone, '') AS phone, 
		       u.role_id, r.code AS role_code, r.name AS role_name, 
		       u.poli_id, COALESCE(p.name, '') AS poli_name, u.status_aktif, u.created_at, u.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		LEFT JOIN poli_units p ON u.poli_id = p.id
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`
	var user domain.User
	err := r.db.Get(&user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAll(search string, roleID int, page, limit int) ([]domain.User, int, error) {
	offset := (page - 1) * limit
	whereClauses := []string{"u.deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(u.username ILIKE $%d OR u.full_name ILIKE $%d OR u.nip ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if roleID > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("u.role_id = $%d", argIndex))
		args = append(args, roleID)
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users u WHERE %s", whereSQL)
	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT u.id, u.username, u.email, u.full_name, 
		       COALESCE(u.nip, '') AS nip, COALESCE(u.phone, '') AS phone, 
		       u.role_id, r.code AS role_code, r.name AS role_name, 
		       u.poli_id, COALESCE(p.name, '') AS poli_name, u.status_aktif, u.created_at, u.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		LEFT JOIN poli_units p ON u.poli_id = p.id
		WHERE %s
		ORDER BY u.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, limit, offset)
	var users []domain.User
	err = r.db.Select(&users, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) Create(u *domain.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, nip, phone, role_id, poli_id, status_aktif, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(query, u.Username, u.Email, u.PasswordHash, u.FullName, u.NIP, u.Phone, u.RoleID, u.PoliID, u.StatusAktif).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) Update(u *domain.User) error {
	query := `
		UPDATE users 
		SET email = $1, full_name = $2, nip = $3, phone = $4, role_id = $5, poli_id = $6, status_aktif = $7, updated_at = NOW()
		WHERE id = $8 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(query, u.Email, u.FullName, u.NIP, u.Phone, u.RoleID, u.PoliID, u.StatusAktif, u.ID)
	return err
}

func (r *UserRepository) UpdatePassword(id, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, passwordHash, id)
	return err
}

func (r *UserRepository) Delete(id, deletedBy string) error {
	query := `UPDATE users SET deleted_at = NOW(), updated_by = $1 WHERE id = $2`
	_, err := r.db.Exec(query, deletedBy, id)
	return err
}

func (r *UserRepository) CountDoctors() (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM users u 
		JOIN roles r ON u.role_id = r.id 
		WHERE r.code = 'DOCTOR' AND u.status_aktif = TRUE AND u.deleted_at IS NULL
	`
	err := r.db.Get(&count, query)
	return count, err
}
