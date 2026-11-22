package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"simrs-asoy/backend/internal/domain"
)

type AuditRepository struct {
	db *sqlx.DB
}

func NewAuditRepository(db *sqlx.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Log(a *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, username, action, entity_type, entity_id, details, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`
	detailsStr := a.Details
	if detailsStr == "" {
		detailsStr = "{}"
	}
	_, err := r.db.Exec(query, a.UserID, a.Username, a.Action, a.EntityType, a.EntityID, detailsStr, a.IPAddress, a.UserAgent)
	return err
}

func (r *AuditRepository) FindLogs(search, action, entityType string, page, limit int) ([]domain.AuditLog, int, error) {
	offset := (page - 1) * limit
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(username ILIKE $%d OR entity_id ILIKE $%d OR details::text ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if action != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, action)
		argIndex++
	}

	if entityType != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("entity_type = $%d", argIndex))
		args = append(args, entityType)
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM audit_logs WHERE %s`, whereSQL)
	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, username, action, entity_type, entity_id, details::text AS details,
		       COALESCE(ip_address, '') AS ip_address, COALESCE(user_agent, '') AS user_agent, created_at
		FROM audit_logs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, limit, offset)
	var list []domain.AuditLog
	err = r.db.Select(&list, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
