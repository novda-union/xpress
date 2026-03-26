package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/xpressgo/server/internal/middleware"
	"github.com/xpressgo/server/internal/model"
	"github.com/xpressgo/server/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	storeRepo *repository.StoreRepo
	staffRepo *repository.StaffRepo
	userRepo  *repository.UserRepo
	jwtSecret string
	botToken  string
}

func NewAuthService(storeRepo *repository.StoreRepo, staffRepo *repository.StaffRepo, userRepo *repository.UserRepo, jwtSecret, botToken string) *AuthService {
	return &AuthService{
		storeRepo: storeRepo,
		staffRepo: staffRepo,
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		botToken:  botToken,
	}
}

// AdminLogin validates store_code + staff_code + password and returns JWT
func (s *AuthService) AdminLogin(ctx context.Context, storeCode, staffCode, password string) (string, *model.Staff, error) {
	store, err := s.storeRepo.GetByCode(ctx, storeCode)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	if !store.IsActive {
		return "", nil, errors.New("store is not active")
	}

	staff, err := s.staffRepo.GetByStoreAndCode(ctx, store.ID, staffCode)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	claims := &middleware.Claims{
		StoreID: store.ID,
		StaffID: staff.ID,
		Role:    staff.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenStr, staff, nil
}

// ValidateTelegramAuth validates Telegram Mini App initData and returns JWT
func (s *AuthService) ValidateTelegramAuth(ctx context.Context, initData string) (string, *model.User, error) {
	// Parse initData
	params, err := url.ParseQuery(initData)
	if err != nil {
		return "", nil, errors.New("invalid init data")
	}

	// Validate hash
	hash := params.Get("hash")
	if hash == "" {
		return "", nil, errors.New("missing hash")
	}

	// Build data check string
	params.Del("hash")
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params.Get(k)))
	}
	dataCheckString := strings.Join(parts, "\n")

	// Compute expected hash
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(s.botToken))
	secret := secretKey.Sum(nil)

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(hash), []byte(expectedHash)) {
		return "", nil, errors.New("invalid hash")
	}

	// Extract user info
	telegramIDStr := params.Get("user_id")
	if telegramIDStr == "" {
		// Try parsing from user JSON — for now, use a simpler approach
		return "", nil, errors.New("missing user_id")
	}

	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		return "", nil, errors.New("invalid user_id")
	}

	// Upsert user
	user := &model.User{
		TelegramID: telegramID,
		FirstName:  params.Get("first_name"),
		LastName:   params.Get("last_name"),
		Username:   params.Get("username"),
	}

	if err := s.userRepo.Upsert(ctx, user); err != nil {
		return "", nil, fmt.Errorf("failed to upsert user: %w", err)
	}

	// Generate JWT
	claims := &middleware.Claims{
		UserID:     user.ID,
		TelegramID: telegramID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenStr, user, nil
}

// GenerateDevToken creates a JWT for development/testing without Telegram validation
func (s *AuthService) GenerateDevToken(ctx context.Context, telegramID int64) (string, *model.User, error) {
	user, err := s.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return "", nil, errors.New("user not found")
	}

	claims := &middleware.Claims{
		UserID:     user.ID,
		TelegramID: telegramID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenStr, user, nil
}
