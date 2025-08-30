package config

import (
	"log" 
	"os" 
	"fmt" 
	"github.com/cloudinary/cloudinary-go/v2" 
)

func InitCloud() *cloudinary.Cloudinary {
	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to cloud: %v", err)
	}
	fmt.Println("Cloud connected successfully")
	return cld
}
