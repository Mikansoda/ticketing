package config

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm" 
)

func ConnectDatabase() *gorm.DB {
	db, err := gorm.Open(mysql.Open(C.DBDSN), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	fmt.Println("Database connected successfully")
	return db
}