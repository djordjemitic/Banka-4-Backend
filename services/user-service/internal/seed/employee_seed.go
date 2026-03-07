package seed

import (
	"time"
	"user-service/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var positions = []string{"Manager", "Developer", "HR"}

var employees = []struct {
	FirstName   string
	LastName    string
	Gender      string
	DateOfBirth string
	Email       string
	PhoneNumber string
	Address     string
	Username    string
	Password    string
	Active      uint16
	Department  string
	Position    string // koristi naziv pozicije, ne ID
}{
	{"Dimitrije", "Mijailovic", "M", "1985-05-01", "dimitrije@raf.rs", "123456789", "Street 1", "dimitrije", "pass123", 1, "IT", "Developer"},
	{"Petar", "Petrovic", "M", "1990-08-12", "petar@raf.rs", "987654321", "Street 2", "petar", "pass123", 1, "HR", "HR"},
}

func Run(db *gorm.DB) error {
	// Seed Positions
	positionMap := make(map[string]uint)
	for _, title := range positions {
		var pos model.Position
		if err := db.Where("title = ?", title).First(&pos).Error; err != nil {
			pos = model.Position{Title: title}
			if err := db.Create(&pos).Error; err != nil {
				return err
			}
		}
		positionMap[title] = pos.PositionID
	}

	// Seed Employees
	for _, e := range employees {
		var existing model.Employee
		if err := db.Where("email = ?", e.Email).First(&existing).Error; err == nil {
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(e.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		dob, _ := time.Parse("2006-01-02", e.DateOfBirth)

		employee := model.Employee{
			FirstName:   e.FirstName,
			LastName:    e.LastName,
			Gender:      e.Gender,
			DateOfBirth: dob,
			Email:       e.Email,
			PhoneNumber: e.PhoneNumber,
			Address:     e.Address,
			Username:    e.Username,
			Password:    string(hash),
			Active:      e.Active,
			Department:  e.Department,
			PositionID:  positionMap[e.Position], // uzima pravi ID iz baze
		}

		if err := db.Create(&employee).Error; err != nil {
			return err
		}
	}

	return nil
}
