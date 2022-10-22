package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/maha-1030/go-seat-booking/handlers"
	"github.com/maha-1030/go-seat-booking/repo"
)

const ()

func main() {
	fmt.Println("Waiting for postgres server to be ready to accept connections")
	time.Sleep(5 * time.Second)
	repo.InitializeDB()

	router := mux.NewRouter().Headers("Content-Type", "application/json").Subrouter()

	router.HandleFunc("/reset-db-with-csv-data", handlers.ResetDBWithCSVDataHandler).Methods(http.MethodGet)
	router.HandleFunc("/seats", handlers.GetSeatsHandler).Methods(http.MethodGet)
	router.HandleFunc("/seats/{id}", handlers.GetSeatPricingHandler).Methods(http.MethodGet)
	router.HandleFunc("/booking", handlers.BookingHandler).Methods(http.MethodPost)
	router.HandleFunc("/bookings", handlers.GetUserBookingsHandler).Methods(http.MethodGet)

	http.ListenAndServe(":8080", router)
}
