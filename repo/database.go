package repo

import (
	"fmt"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

var (
	db *gorm.DB
)

func InitializeDB() {
	var (
		err           error
		connectionURL string
	)

	defer panicIfHasError(err, "Connecting Database")

	//Load .env file
	if err = godotenv.Load(); err != nil {
		fmt.Printf("Error occured while reading .env file, err: %v\n", err)

		return
	}

	username := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	dbName := os.Getenv("POSTGRES_DB")

	connectionURL = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", host, port, username, dbName, password)

	db, err = gorm.Open("postgres", connectionURL)
	if err != nil {
		fmt.Printf("Error occured while connecting to database, err: %v\n", err)

		return
	}

	if err = db.Debug().DB().Ping(); err != nil {
		fmt.Printf("Error occured while pinging the database, err: %v", err)

		return
	}

	fmt.Printf("\nSuccessfully connected to the %v database in postgres server\n", dbName)
	autoMigrate()
}

func autoMigrate() {
	var err error

	defer panicIfHasError(err, "Automigrating Models")

	if err = db.Debug().AutoMigrate(&Seat{}, &SeatPrice{}, &Booking{}, &User{}).Error; err != nil {
		fmt.Printf("\nError occured while automigrating models into the database, err: %v\n", err)

		return
	}
}

func PopulateData() error {
	var (
		err                       error
		seatsCSV, seatPricingsCSV *os.File
		seats                     []*Seat
		seatPricings              []*SeatPrice
	)

	defer panicIfHasError(err, "Populating Data")

	if seatsCSV, err = os.Open("./data/seats.csv"); err != nil {
		fmt.Printf("\nError occured while opening data file of seats, err: %v\n", err)

		return fmt.Errorf("error occured while opening data file of seats")
	}

	if seatPricingsCSV, err = os.Open("./data/seatpricings.csv"); err != nil {
		fmt.Printf("\nError occured while opening data file of seatpricings, err: %v\n", err)

		return fmt.Errorf("error occured while opening data file of seatpricings")
	}

	if err = gocsv.Unmarshal(seatsCSV, &seats); err != nil {
		fmt.Printf("\nError occured while unmarshaling seats data into Seat struct format, err: %v\n", err)

		return fmt.Errorf("error occured while unmarshaling seats data into Seat struct format")
	}

	if err = gocsv.Unmarshal(seatPricingsCSV, &seatPricings); err != nil {
		fmt.Printf("\nError occured while unmarshaling seatpricings data into SeatPrice struct format, err: %v\n", err)

		return fmt.Errorf("error occured while unmarshaling seatpricings data into SeatPrice struct format")
	}

	tx := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Debug().Delete(&Seat{}).Error; err != nil {
		fmt.Printf("\nError occured while deleting existing data from seats table, err: %v\n", err)

		return fmt.Errorf("error occured while deleting existing data from seats table")
	}

	if err = tx.Debug().Delete(&SeatPrice{}).Error; err != nil {
		fmt.Printf("\nError occured while deleting existing data from seat_prices table, err: %v\n", err)

		return fmt.Errorf("error occured while deleting existing data from seat_prices table")
	}

	for i, s := range seats {
		if err = tx.Create(s).Error; err != nil {
			fmt.Printf("\nError occured while inserting records into seats table, record no: %v, err: %v\n", i+1, err)

			return fmt.Errorf("error occured while inserting records into seats table, record no: %v, err: %v\n", i+1, err)
		}
	}

	for i, sp := range seatPricings {
		if err = tx.Create(sp).Error; err != nil {
			fmt.Printf("\nError occured while inserting records into seat_prices table, record no: %v, err: %v\n", i+1, err)

			return fmt.Errorf("error occured while inserting records into seat_prices table, record no: %v, err: %v\n", i+1, err)
		}
	}

	if err = tx.Commit().Error; err != nil {
		fmt.Printf("\n Error occured while committing the changes, err: %v\n", err)

		return fmt.Errorf("unable to commit changes into the db")
	}

	fmt.Printf("\nSuccessfully truncated tables and inserted %v records into seats table & %v records into seat_prices table\n", len(seats), len(seatPricings))

	return nil
}

func panicIfHasError(err error, functionalityWithError string) {
	if err != nil {
		panic("\nERROR!!! " + functionalityWithError + "\n")
	}
}
