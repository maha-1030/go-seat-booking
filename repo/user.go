package repo

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type User struct {
	ID          int       `gorm:"primary_key"`
	Name        string    `gorm:"name"`
	PhoneNumber string    `gorm:"phone_number"`
	Email       string    `gorm:"email"`
	Bookings    []Booking `gorm:"foreignKey:UserID;references:ID"`
}

func (u *User) Validate() error {
	if u.PhoneNumber == "" && u.Email == "" {
		return fmt.Errorf("either parameter 'PhoneNumber' or parameter 'Email' is needed")
	}

	return nil
}

func (u *User) Save() error {
	if err := db.Debug().Create(u).Error; err != nil {
		return err
	}

	return nil
}

func (u *User) GetByPhoneNumber(phoneNumber string) (*User, error) {
	if err := db.Debug().Model(u).Where("phone_number = ?", phoneNumber).First(u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return u, nil
}

func (u *User) GetByUserIdentifier(identifier string) (*User, error) {
	if err := db.Debug().Model(u).Where("phone_number = ? or email = ?", identifier, identifier).
		Preload("Bookings").Preload("Bookings.Seats").Find(u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return u, nil
}
