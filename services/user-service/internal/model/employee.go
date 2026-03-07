package model

import (
	"time"
)

type Position struct {
	PositionID uint   `gorm:"primaryKey"`
	Title      string `gorm:"size:100;not null"`
	// Employees  []Employee `gorm:"foreignKey:PositionID"` // 1M veza
}

type Employee struct {
	EmployeeID  uint   `gorm:"primaryKey"`
	FirstName   string `gorm:"size:20;not null"`
	LastName    string `gorm:"size:100;not null"`
	Gender      string `gorm:"size:10"`
	DateOfBirth time.Time
	Email       string `gorm:"size:100;uniqueIndex"`
	PhoneNumber string `gorm:"size:20"`
	Address     string `gorm:"size:255"`
	Username    string `gorm:"size:50;uniqueIndex"`
	Password    string `gorm:"size:255"`
	Active      uint16
	Department  string `gorm:"size:100"`
	PositionID  uint   `gorm:""`
	Position    Position
}
