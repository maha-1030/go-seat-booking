package repo

import (
	"fmt"
)

type Booking struct {
	ID          int    `gorm:"primary_key"`
	UserID      int    `gorm:"user_id"`
	Name        string `gorm:"name"`
	TotalAmount string `gorm:"total_amount"`
	Seats       []Seat `gorm:"foreignKey:BookingID;references:ID"`
}

func (b *Booking) Validate() error {
	if b.TotalAmount == "" {
		return fmt.Errorf("parameter 'TotalAmount' is missing")
	}

	if !isValidPrice(b.TotalAmount) {
		return fmt.Errorf("invalid parameter 'TotalAmount' should be a positive value and prefixed with '$'")
	}

	return nil
}

func (b *Booking) Save() (err error) {
	if err := db.Debug().Create(b).Error; err != nil {

		return err
	}

	return nil
}
