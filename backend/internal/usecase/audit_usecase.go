package usecase

import (
	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/repository"
)

type AuditUsecase struct {
	auditRepo *repository.AuditRepository
}

func NewAuditUsecase(auditRepo *repository.AuditRepository) *AuditUsecase {
	return &AuditUsecase{auditRepo: auditRepo}
}

func (u *AuditUsecase) GetAuditLogs(search, action, entityType string, page, limit int) ([]domain.AuditLog, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.auditRepo.FindLogs(search, action, entityType, page, limit)
}
