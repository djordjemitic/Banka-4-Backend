package service

import (
	"common/pkg/jwt"
	"context"
	"fmt"
	"user-service/internal/config"

	"common/pkg/errors"
	_ "common/pkg/jwt"
	"user-service/internal/dto"
	"user-service/internal/model"
	"user-service/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type EmployeeService struct {
	repo repository.EmployeeRepository // <-- no pointer
	cfg  *config.Configuration
}

func NewEmployeeService(repo repository.EmployeeRepository, cfg *config.Configuration) *EmployeeService {
	return &EmployeeService{repo: repo, cfg: cfg}
}

func (s *EmployeeService) Register(ctx context.Context, req *dto.CreateEmployeeRequest) (*model.Employee, error) {

	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	if existing != nil {
		return nil, errors.ConflictErr("email already in use")
	}

	existingByUsername, err := s.repo.FindByUserName(ctx, req.Username)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if existingByUsername != nil {
		return nil, errors.ConflictErr("username already in use")
	}
	employee := &model.Employee{
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Gender:      req.Gender,
		DateOfBirth: req.DateOfBirth,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		Username:    req.Username,
		Department:  req.Department,
		PositionID:  req.PositionID,
		Active:      req.Active,
	}

	if err := s.repo.Create(ctx, employee); err != nil {
		return nil, errors.InternalErr(err)
	}

	return employee, nil
}

func (s *EmployeeService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	//Pronadji zaposlenog po email-u
	employee, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if employee == nil {
		return nil, errors.UnauthorizedErr("invalid credentials")
	}

	//Proveri da li je zaposleni aktivan
	if !employee.Active {
		return nil, errors.ForbiddenErr("account is disabled")
	}

	//Proveri lozinku koristeci bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(employee.Password), bcrypt.DefaultCost)
	fmt.Println(hashedPassword) //DEBUG REMOVE
	err = bcrypt.CompareHashAndPassword([]byte(employee.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.UnauthorizedErr("invalid credentials")
	}

	//Generisi token
	token, err := jwt.GenerateToken(employee.EmployeeID, s.cfg.JWTSecret, s.cfg.JWTExpiry)
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	//Vrati generisani token
	return &dto.LoginResponse{Token: token}, nil
}
