package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"time"

	"ticketing/entity"
	"ticketing/helper"
	"ticketing/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(ctx context.Context, fullName, username, email, password, role string) error
	VerifyOTP(ctx context.Context, email, otp string) error
	Login(ctx context.Context, email, password string) (access string, refresh string, err error)
	Refresh(ctx context.Context, refreshToken string) (access string, newRefresh string, err error)
	Logout(ctx context.Context, accessToken string) error
}

type authService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) AuthService {
	return &authService{repo: repo}
}

func (s *authService) Register(ctx context.Context, fullName, username, email, password, role string) error {
	_, err := s.repo.FindByEmail(ctx, email)
	switch {
	case err == nil:
		return errors.New("email already registered")
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return errors.New("failed to register user: " + err.Error())
	}

	_, err = s.repo.FindByUsername(ctx, username)
	if err == nil {
		return errors.New("username already taken")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("failed to register user: " + err.Error())
	}

	// HASHING W BCRYPT
	pwHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// generate OTP plaintext & hash
	otpPlain := generateNumericOTP(6)
	otpHash, _ := bcrypt.GenerateFromPassword([]byte(otpPlain), bcrypt.DefaultCost)
	exp := time.Now().Add(10 * time.Minute)

	u := &entity.Users{
		ID:           uuid.New(),
		FullName:     fullName,
		Email:        email,
		Username:     username,
		PasswordHash: string(pwHash),
		Role:         entity.RoleUser,
		IsActive:     false,
		OTPHash:      string(otpHash),
		OTPExpiresAt: &exp,
	}

	fmt.Printf("DEBUG: Saving user -> Email: %s, Username: %s, Role: %s\n", u.Email, u.Username, u.Role)
	if err := s.repo.Create(ctx, u); err != nil {
		return errors.New("failed to create user: " + err.Error())
	}

	if err := helper.SendEmailOTP(email, username, otpPlain); err != nil {
		return errors.New("failed to send OTP email: " + err.Error())
	}
	return nil
}

func (s *authService) VerifyOTP(ctx context.Context, email, otp string) error {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return errors.New("failed to verify OTP: user not found")
	}
	if u.OTPExpiresAt == nil || time.Now().After(*u.OTPExpiresAt) {
		return errors.New("failed to verify OTP: otp expired")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.OTPHash), []byte(otp)); err != nil {
		return errors.New("failed to verify OTP: invalid otp")
	}
	u.IsActive = true
	u.OTPHash = ""
	u.OTPExpiresAt = nil
	if err := s.repo.Update(ctx, u); err != nil {
		return errors.New("failed to verify OTP: " + err.Error())
	}
	return nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("failed to login: invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", "", errors.New("failed to login: invalid credentials")
	}
	if !u.IsActive {
		return "", "", errors.New("failed to login: account not active, verify otp")
	}
	access, _, err := helper.GenerateAccessToken(u.ID.String(), u.Email, string(u.Role))
	if err != nil {
		return "", "", errors.New("failed to generate access token: " + err.Error())
	}
	refresh, rexp, err := helper.GenerateRefreshToken(u.ID.String(), u.Email, string(u.Role))
	if err != nil {
		return "", "", errors.New("failed to generate refresh token: " + err.Error())
	}
	rh, _ := bcrypt.GenerateFromPassword([]byte(refresh), bcrypt.DefaultCost)
	u.RefreshTokenHash = string(rh)
	u.RefreshExpiresAt = &rexp
	if err := s.repo.Update(ctx, u); err != nil {
		return "", "", errors.New("failed to save refresh token: " + err.Error())
	}
	return access, refresh, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	claims, email, _, err := helper.ParseRefresh(refreshToken)
	if err != nil {
		return "", "", errors.New("failed to refresh token: invalid refresh token")
	}

	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("failed to refresh token: invalid refresh token")
	}
	if u.RefreshExpiresAt == nil || time.Now().After(*u.RefreshExpiresAt) {
		return "", "", errors.New("failed to refresh token: refresh expired")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.RefreshTokenHash), []byte(refreshToken)); err != nil {
		return "", "", errors.New("failed to refresh token: invalid refresh token")
	}

	access, _, err := helper.GenerateAccessToken(claims.UserID, u.Email, string(u.Role))
	if err != nil {
		return "", "", errors.New("failed to refresh token: " + err.Error())
	}
	newRefresh, rexp, err := helper.GenerateRefreshToken(claims.UserID, u.Email, string(u.Role))
	if err != nil {
		return "", "", errors.New("failed to refresh token: " + err.Error())
	}
	rh, _ := bcrypt.GenerateFromPassword([]byte(newRefresh), bcrypt.DefaultCost)
	u.RefreshTokenHash = string(rh)
	u.RefreshExpiresAt = &rexp
	if err := s.repo.Update(ctx, u); err != nil {
		return "", "", errors.New("failed to save new refresh token: " + err.Error())
	}
	return access, newRefresh, nil
}

// In-memory blacklist
var accessBlacklist = make(map[string]time.Time)

func (s *authService) Logout(ctx context.Context, accessToken string) error {
	// blacklist until expiry claim
	claims, err := helper.ParseAccess(accessToken)
	if err != nil {
		return errors.New("failed to logout: invalid token")
	}
	exp := claims.ExpiresAt.Time
	accessBlacklist[accessToken] = exp
	return nil
}

func generateNumericOTP(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "000000"
	}
	code := strings.ToUpper(base32.StdEncoding.EncodeToString(b))
	digs := make([]rune, 0, n)
	for _, c := range code {
		if c >= '0' && c <= '9' {
			digs = append(digs, c)
			if len(digs) == n {
				break
			}
		}
	}
	for len(digs) < n {
		digs = append(digs, '0')
	}
	return string(digs)
}
