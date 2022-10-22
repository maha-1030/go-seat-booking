package repo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
)

type SeatPrice struct {
	ID          int    `gorm:"primary_key" csv:"id"`
	SeatClass   string `gorm:"seat_class" csv:"seat_class"`
	MinPrice    string `gorm:"min_price" csv:"min_price"`
	NormalPrice string `gorm:"normal_price" csv:"normal_price"`
	MaxPrice    string `gorm:"max_price" csv:"max_price"`
}

func (sp *SeatPrice) Validate() error {
	if sp.SeatClass == "" {
		return fmt.Errorf("parameter 'SeatClass' is missing")
	}

	if sp.NormalPrice == "" {
		return fmt.Errorf("parameter 'SeatClass' is missing")
	}

	if !isValidPrice(sp.NormalPrice) {
		return fmt.Errorf("invalid parameter 'NormalPrice' should be a positive value and prefixed with '$'")
	}

	if !isValidPrice(sp.MinPrice) {
		return fmt.Errorf("invalid parameter 'MinPrice' should be a positive value and prefixed with '$'")
	}

	if !isValidPrice(sp.MaxPrice) {
		return fmt.Errorf("invalid parameter 'MaxPrice' should be a positive value and prefixed with '$'")
	}

	return nil
}

func isValidPrice(price string) bool {
	if price == "" {
		return true
	}

	if strings.HasPrefix(price, "$") {
		price = price[1:]
	} else {
		return false
	}

	value, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return false
	}

	if value < 0 {
		return false
	}

	return true
}

func (sp *SeatPrice) Save() error {
	if err := db.Debug().Create(sp).Error; err != nil {
		return err
	}

	return nil
}

func (sp *SeatPrice) GetByClass(class string) (*SeatPrice, error) {
	if err := db.Debug().First(sp, "seat_class = ?", class).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return sp, nil
}
