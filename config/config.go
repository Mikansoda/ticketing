package config

import (
	"log"
	"os" 
	"strconv" 
	"fmt" 
)

// storage for app configurations as written
type AppConfig struct {
	AppPort        string 
	DBDSN          string 
	JWTAccessKey   string 
	JWTRefreshKey  string
	AccessTTLMin   int 
	RefreshTTLDays int 
	SMTPHost       string 
	SMTPPort       int 
	SMTPUser       string 
	SMTPPass       string 
	FromEmail      string 
	Env            string 
}

var C AppConfig

// isi struct dengan yang ada di .env
// Init() → loader utama config (isi struct C).
func Init() {
	C = AppConfig{
		AppPort: getenv("APP_PORT", "8080"),
		DBDSN: getenv("DB_DSN", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		)),
		JWTAccessKey:   must("JWT_ACCESS_SECRET"),
		JWTRefreshKey:  must("JWT_REFRESH_SECRET"),
		AccessTTLMin:   atoi(getenv("ACCESS_TTL_MIN", "15")),
		RefreshTTLDays: atoi(getenv("REFRESH_TTL_DAYS", "7")),
		SMTPHost:       getenv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:       atoi(getenv("SMTP_PORT", "587")),
		SMTPUser:       must("SMTP_USER"),
		SMTPPass:       must("SMTP_PASS"),
		FromEmail:      getenv("FROM_EMAIL", "myappondev@gmail.com"),
		Env:            getenv("APP_ENV", "dev"),
	}
}

// ambil nilai env berdasarkan k, mis APP_PORT, DB_USER, dll yg wajib
func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing required env: %s", k)
	}
	return v
}

// yg bisa pake/fallback default
func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

// Dipake buat nilai numeric di env (contoh: ACCESS_TTL_MIN, SMTP_PORT). 
// Convert string ke int
func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}