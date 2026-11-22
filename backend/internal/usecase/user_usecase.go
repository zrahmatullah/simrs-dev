package usecase

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"simrs-asoy/backend/internal/domain"
	"simrs-asoy/backend/internal/repository"
)

type UserUsecase struct {
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
}

func NewUserUsecase(userRepo *repository.UserRepository, auditRepo *repository.AuditRepository) *UserUsecase {
	return &UserUsecase{userRepo: userRepo, auditRepo: auditRepo}
}

func (u *UserUsecase) GetUsers(search string, roleID, page, limit int) ([]domain.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.userRepo.FindAll(search, roleID, page, limit)
}

func (u *UserUsecase) GetUserByID(id string) (*domain.User, error) {
	return u.userRepo.FindByID(id)
}

func (u *UserUsecase) CreateUser(req domain.CreateUserRequest, currentUserID, currentUsername string) (*domain.User, error) {
	if req.Username == "" || req.Password == "" || req.FullName == "" || req.RoleID == 0 {
		return nil, errors.New("username, password, nama lengkap, dan role wajib diisi")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		NIP:          req.NIP,
		Phone:        req.Phone,
		RoleID:       req.RoleID,
		PoliID:       req.PoliID,
		StatusAktif:  true,
	}

	err = u.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "CREATE",
		EntityType: "user",
		EntityID:   user.ID,
		Details:    `{"username": "` + user.Username + `", "role_id": ` + string(rune(user.RoleID)) + `}`,
	})

	return user, nil
}

func (u *UserUsecase) UpdateUser(id string, req domain.UpdateUserRequest, currentUserID, currentUsername string) error {
	user, err := u.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	user.Email = req.Email
	user.FullName = req.FullName
	user.NIP = req.NIP
	user.Phone = req.Phone
	user.RoleID = req.RoleID
	user.PoliID = req.PoliID
	user.StatusAktif = req.StatusAktif

	err = u.userRepo.Update(user)
	if err != nil {
		return err
	}

	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		if err == nil {
			_ = u.userRepo.UpdatePassword(id, string(hashed))
		}
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "UPDATE",
		EntityType: "user",
		EntityID:   id,
		Details:    `{"username": "` + user.Username + `"}`,
	})

	return nil
}

func (u *UserUsecase) DeleteUser(id string, currentUserID, currentUsername string) error {
	err := u.userRepo.Delete(id, currentUserID)
	if err != nil {
		return err
	}

	_ = u.auditRepo.Log(&domain.AuditLog{
		UserID:     &currentUserID,
		Username:   currentUsername,
		Action:     "DELETE",
		EntityType: "user",
		EntityID:   id,
		Details:    `{"deleted": true}`,
	})

	return nil
}
