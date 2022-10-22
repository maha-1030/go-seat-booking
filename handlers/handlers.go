package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/maha-1030/go-seat-booking/repo"
)

func ResetDBWithCSVDataHandler(w http.ResponseWriter, r *http.Request) {
	if err := repo.PopulateData(); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())

		return
	}

	respondWithJson(w, http.StatusOK, "success")
}

func GetSeatsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		seats    []repo.Seat
		response []struct {
			ID             int
			SeatIdentifier string
			SeatClass      string
			IsBooked       bool
		}
		err error
	)

	if seats, err = (&repo.Seat{}).Get(repo.SEAT_CLASS_FIELD, repo.ASCENDING_ORDER); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())

		return
	}

	for i := range seats {
		response = append(response, struct {
			ID             int
			SeatIdentifier string
			SeatClass      string
			IsBooked       bool
		}{
			ID:             seats[i].ID,
			SeatIdentifier: seats[i].SeatIdentifier,
			SeatClass:      seats[i].SeatClass,
			IsBooked:       seats[i].BookingID != nil,
		})
	}

	respondWithJson(w, http.StatusOK, response)
}

func GetSeatPricingHandler(w http.ResponseWriter, r *http.Request) {
	var (
		seat     *repo.Seat = &repo.Seat{}
		response struct {
			ID             int
			SeatIdentifier string
			SeatClass      string
			IsBooked       bool
			Price          string
		}
	)

	vars := mux.Vars(r)

	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "missing path param 'id'")

		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		respondWithError(w, http.StatusBadRequest, "invalid path param 'id', it should be a positive interger")

		return
	}

	seat, err = seat.GetByID(id)
	if err != nil {
		fmt.Printf("\nError occured while getting seat with id: %v, err: %v\n", id, err)
		respondWithError(w, http.StatusInternalServerError, "internal server error")

		return
	} else if seat == nil {
		respondWithError(w, http.StatusNotFound, "no seat found with id: "+idStr)

		return
	}

	response.ID = seat.ID
	response.SeatClass = seat.SeatClass
	response.SeatIdentifier = seat.SeatIdentifier
	response.IsBooked = seat.BookingID != nil

	if response.Price, err = getPriceOfSeat(seat.SeatClass); err != nil {
		respondWithError(w, http.StatusInternalServerError, "internal server error")

		return
	} else if response.Price == "" {
		respondWithError(w, http.StatusInternalServerError, "pricing for class: "+seat.SeatClass+" is not available")

		return
	}

	respondWithJson(w, http.StatusOK, response)
}

func BookingHandler(w http.ResponseWriter, r *http.Request) {
	var (
		user        *repo.User = &repo.User{}
		seat        *repo.Seat = &repo.Seat{}
		seats       []repo.Seat
		totalAmount string
		request     struct {
			IDs         []int
			Name        string
			PhoneNumber string
		}
		available bool
		err       error
	)

	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		fmt.Println("\nError while decoding the request body in booking request, err: ", err)
		respondWithError(w, http.StatusBadRequest, "unable to decode request body")

		return
	}

	if available, err = seat.IsAvailableForBooking(request.IDs); err != nil {
		fmt.Println("\nError occured while checking for availability of seats, err: ", err)
		respondWithError(w, http.StatusInternalServerError, "internal server error")

		return
	} else if !available {
		respondWithError(w, http.StatusBadRequest, "one or more of the given seats are not available")

		return
	}

	if seats, totalAmount, err = getSeatsAndPrices(request.IDs); err != nil {
		respondWithError(w, http.StatusInternalServerError, "internal server error")

		return
	} else if seats == nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("One or more seats not found with ids: %v", request.IDs))

		return
	} else if totalAmount == "" {
		respondWithError(w, http.StatusInternalServerError, "pricing for any one of the seat class is not available")

		return
	}

	if user, err = user.GetByPhoneNumber(request.PhoneNumber); err != nil {
		fmt.Printf("\nError occured while getting user with phoneNumber: %v, err: %v", request.PhoneNumber, err)
		respondWithError(w, http.StatusInternalServerError, "internal server error")

		return
	} else if user == nil {
		user = &repo.User{
			Name:        request.Name,
			PhoneNumber: request.PhoneNumber,
		}

		if err = user.Save(); err != nil {
			fmt.Printf("\nError occured while creating user with phoneNumber: %v, err: %v", request.PhoneNumber, err)
			respondWithError(w, http.StatusInternalServerError, "internal server error")

			return
		}
	}

	booking := repo.Booking{
		UserID:      user.ID,
		Name:        request.Name,
		TotalAmount: totalAmount,
		Seats:       seats,
	}

	if err = booking.Save(); err != nil {
		fmt.Printf("\nError occured while creating booking with phoneNumber: %v, err: %v", request.PhoneNumber, err)
		respondWithError(w, http.StatusInternalServerError, "internal server error")

		return
	}

	respondWithJson(w, http.StatusOK, booking)
}

func GetUserBookingsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		user *repo.User = &repo.User{}
	)
	userIdentifier := r.URL.Query().Get("userIdentifier")

	if userIdentifier == "" {
		respondWithError(w, http.StatusBadRequest, "userIdentifier is not provided")

		return
	}

	if user, err := user.GetByUserIdentifier(userIdentifier); err != nil {
		fmt.Printf("\nError occured while getting user with userIdentifier: %v, err: %v\n", userIdentifier, err)
		respondWithError(w, http.StatusInternalServerError, "internal server error")

		return
	} else if user == nil {
		respondWithError(w, http.StatusNotFound, "No user found with given userIdentifier: "+userIdentifier)

		return
	}

	respondWithJson(w, http.StatusOK, user)
}

func getPriceOfSeat(class string) (seatPrice string, err error) {
	seat := &repo.Seat{}
	price := &repo.SeatPrice{}

	occupied, total, err := seat.GetSeatsCount(class)
	if err != nil {
		fmt.Printf("\nError occured while getting seat count of class: %v, err: %v\n", seat.SeatClass, err)

		return "", err
	}

	price, err = price.GetByClass(class)
	if err != nil {
		fmt.Printf("\nError occured while getting seat price of class: %v, err: %v\n", seat.SeatClass, err)

		return "", err
	} else if price == nil {
		return "", nil
	}

	occupancy := float64(occupied) / float64(total) * 100

	if occupancy < 40 && price.MinPrice != "" {
		return price.MinPrice, nil
	}

	if occupancy >= 60 && price.MaxPrice != "" {
		return price.MaxPrice, nil
	}

	return price.NormalPrice, nil
}

func getSeatsAndPrices(seatIDs []int) (seats []repo.Seat, totalPrice string, err error) {
	seats = make([]repo.Seat, 0)

	var (
		seat           *repo.Seat     = &repo.Seat{}
		seatClassCount map[string]int = make(map[string]int)
		price          string
		prices         []string
	)

	for i := range seatIDs {
		if seat, err = seat.GetByID(seatIDs[i]); err != nil {
			fmt.Printf("\nError occured while getting the seat with id: %v, err: %v\n", seatIDs[i], err)

			return nil, "", err
		} else if seat == nil {
			return nil, "", nil
		}

		seatClassCount[seat.SeatClass]++

		seats = append(seats, *seat)
		seat = &repo.Seat{}
	}

	for class, count := range seatClassCount {
		if price, err = getPriceOfSeat(class); err != nil {
			fmt.Printf("\nError occured while getting seat price of class: %v, err: %v\n", class, err)

			return seats, "", err
		} else if price == "" {
			return seats, "", err
		}

		for count > 0 {
			prices = append(prices, price)
			count--
		}
	}

	return seats, sumOfPrices(prices), nil
}

func sumOfPrices(prices []string) (totalPrice string) {
	var totalValue float64

	for i := range prices {
		valueString := strings.TrimPrefix(prices[i], "$")
		value, _ := strconv.ParseFloat(valueString, 64)

		totalValue += value
	}

	return "$" + fmt.Sprintf("%v", totalValue)
}

func respondWithJson(w http.ResponseWriter, statusCode int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(body)
}

func respondWithError(w http.ResponseWriter, statusCode int, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
}
