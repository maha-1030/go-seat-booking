package repo

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

const (
	SEAT_CLASS_FIELD = "seat_class"
	ASCENDING_ORDER  = "ASC"
)

type Seat struct {
	ID             int    `gorm:"primary_key" csv:"id"`
	SeatIdentifier string `gorm:"seat_identifier" csv:"seat_identifier"`
	SeatClass      string `gorm:"seat_class" csv:"seat_class"`
	BookingID      *int   `gorm:"booking_id"`
}

func (s *Seat) Validate() error {
	if s.SeatIdentifier == "" {
		return fmt.Errorf("parameter 'SeatIdentifier' is missing")
	}

	if s.SeatClass == "" {
		return fmt.Errorf("parameter 'SeatClass' is missing")
	}

	return nil
}

func (s *Seat) Save() error {
	if err := db.Debug().Create(s).Error; err != nil {
		return err
	}

	return nil
}

func (s *Seat) Get(field, order string) ([]Seat, error) {
	seats := make([]Seat, 0)
	query := db.Debug().Model(s)

	if field != "" {
		if order != "" {
			query = query.Order(field + " " + order)
		} else {
			query = query.Order(field)
		}
	}

	if err := query.Scan(&seats).Error; err != nil {
		return nil, err
	}

	return seats, nil
}

func (s *Seat) GetByID(id int) (*Seat, error) {
	if err := db.Debug().First(s, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return s, nil
}

func (s *Seat) GetSeatsCount(class string) (occupied, total int, err error) {
	query := db.Debug().Model(s)

	if class != "" {
		query = query.Where("seat_class = ?", class)
	}

	if err = query.Count(&total).Error; err != nil {
		return 0, 0, err
	}

	if err = query.Where("booking_id IS NOT NULL").Count(&occupied).Error; err != nil {
		return 0, 0, err
	}

	return occupied, total, nil
}

func (s *Seat) IsAvailableForBooking(ids []int) (available bool, err error) {
	var unavailableSeatsCount int

	if err = db.Debug().Model(s).Where(ids).Where("booking_id IS NOT NULL").
		Count(&unavailableSeatsCount).Error; err != nil {
		return false, err
	}

	if unavailableSeatsCount > 0 {
		return false, nil
	}

	return true, nil
}
