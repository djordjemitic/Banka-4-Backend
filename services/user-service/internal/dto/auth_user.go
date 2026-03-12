package dto

import (
	"common/pkg/permission"
	"user-service/internal/model"
)

type AuthUser struct {
	ID          uint                    `json:"id"`
	FirstName   string                  `json:"first_name"`
	LastName    string                  `json:"last_name"`
	Email       string                  `json:"email"`
	Username    string                  `json:"username"`
	Permissions []permission.Permission `json:"permissions"`
}

func NewAuthUser(employee *model.Employee) *AuthUser {
	if employee == nil {
		return nil
	}

	return &AuthUser{
		ID:          employee.EmployeeID,
		FirstName:   employee.FirstName,
		LastName:    employee.LastName,
		Email:       employee.Email,
		Username:    employee.Username,
		Permissions: employee.RawPermissions(),
	}
}
