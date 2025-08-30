package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"ticketing/config"
	"ticketing/entity"

	"ticketing/repository"
	"ticketing/routes"
	"ticketing/service"
)

func CreateTableIfNotExists(db *gorm.DB, model interface{}) error {
	if !db.Migrator().HasTable(model) {
		if err := db.Migrator().CreateTable(model); err != nil {
			return err
		}
	}
	return nil
}

func MigrateTables(db *gorm.DB) error {
	if err := CreateTableIfNotExists(db, &entity.Users{}); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&entity.Visitors{},
	); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&entity.EventCategories{},
		&entity.Events{},
		&entity.EventImages{},
	); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&entity.TicketTypes{},
		&entity.Bookings{},
		&entity.Tickets{},
	); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&entity.Payments{},
	); err != nil {
		return err
	}
	return nil
}

func MigrateDatabase(db *gorm.DB) {
	err := MigrateTables(db)
	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
	log.Println("Database migration successful")
}

func main() {
	// load env
	_ = godotenv.Load()
	config.Init()
	config.InitCloud()
	// connect ke database, hasil connection object (*gorm.DB) namanya db
	db := config.ConnectDatabase()
	if err := CreateTableIfNotExists(db, &entity.Users{}); err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	// Migrate db, inject connection sbg context db target
	MigrateDatabase(db)

	bookingRepo := repository.NewBookingRepository(db)
	ticketTypeRepo := repository.NewTicketTypeRepository(db)
	tickeRepo := repository.NewTicketRepository(db)
	eventsRepo := repository.NewEventRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	ticketTypeSvc := service.NewTicketTypeService(ticketTypeRepo, eventsRepo)
	paymentSvc := service.NewPaymentService(paymentRepo, bookingRepo, tickeRepo, ticketTypeSvc, db)

	xenditAPIKey := os.Getenv("XENDIT_API_KEY")
	// Panggil routes.SetupRouter yg di dalamnya daftar endpoint
	// Route diarahkan ke controller, controller panggil service, service panggil repository.
	r := routes.SetupRouter(db, xenditAPIKey)

	// Background Job → Auto cancel pending payment
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // check every 1 hour
		for range ticker.C {
			log.Println("Running auto-cancel pending payments...")
			paymentSvc.AutoCancelPendingPayments()
		}
	}()
	// Start server (nyalain Gin HTTP)
	log.Println("listening on :" + config.C.AppPort)
	if err := r.Run(":" + config.C.AppPort); err != nil {
		log.Fatal(err)
	}
}
