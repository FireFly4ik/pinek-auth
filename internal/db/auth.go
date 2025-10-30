package db

import (
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserExists         = errors.New("User already exists")
	ErrInvalidCredentials = errors.New("Invalid credentials")
	ErrTokenNotFound      = errors.New("Token not found")
)

func (db *AuthDatabase) RegisterUser(username, login, hashedPassword, role string) (string, error) {
	registration := &Registration{
		Username:       username,
		Login:          login,
		HashedPassword: hashedPassword,
		Role:           role,
	}

	tx := db.Database.Begin()

	if err := tx.Create(registration).Error; err != nil {
		tx.Rollback()
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", ErrUserExists
		} else {
			log.Error().Err(err).Msg("failed to create token")
			return "", err
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Error().Err(err).Msg("failed to commit transaction")
		tx.Rollback()
		return "", err
	}

	return registration.ID.String(), nil
}

func (db *AuthDatabase) StoreRefreshToken(userID string, tokenHash []byte, expiresAt time.Time) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse user ID")
		return err
	}

	refreshToken := &RefreshToken{
		UserID:    userUUID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	tx := db.Database.Begin()

	if err := tx.Create(refreshToken).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("failed to create token")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		log.Error().Err(err).Msg("failed to commit transaction")
		tx.Rollback()
		return err
	}

	return nil
}

func (db *AuthDatabase) DeleteUser(userID string) error {
	tx := db.Database.Begin()

	if err := tx.Delete(&Registration{}, "id = ?", userID).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("failed to delete user")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		log.Error().Err(err).Msg("failed to commit transaction")
		tx.Rollback()
		return err
	}

	return nil
}

func (db *AuthDatabase) AuthenticateUser(login string) (string, string, string, string, error) {
	var registration Registration
	if err := db.Database.Where("login = ?", login).First(&registration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", "", "", "", ErrInvalidCredentials
		}
		log.Error().Err(err).Msg("failed to authenticate user")
		return "", "", "", "", err
	}

	return registration.ID.String(), registration.Username, registration.HashedPassword, registration.Role, nil
}

func (db *AuthDatabase) CheckRefreshTokenHash(userID string, tokenHash []byte) error {
	var refreshToken RefreshToken
	if err := db.Database.Where("user_id = ? AND token_hash = ?", userID, tokenHash).First(&refreshToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrTokenNotFound
		}
		log.Error().Err(err).Msg("failed to get refresh token hash")
		return err
	}

	return nil
}

func (db *AuthDatabase) DeleteRefreshToken(userID string, tokenHash []byte) error {
	tx := db.Database.Begin()

	if err := tx.Delete(&RefreshToken{}, "user_id = ? AND token_hash = ?", userID, tokenHash).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("failed to delete refresh token")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		log.Error().Err(err).Msg("failed to commit transaction")
		tx.Rollback()
		return err
	}

	return nil
}

func (db *AuthDatabase) GetRegistrationByID(userID string) (*Registration, error) {
	var registration Registration
	if err := db.Database.Where("id = ?", userID).First(&registration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrInvalidCredentials
		}
		log.Error().Err(err).Msg("failed to get registration by ID")
		return nil, err
	}

	return &registration, nil
}

func (db *AuthDatabase) RevokeRefreshToken(hashedToken []byte, reason string) error {
	tx := db.Database.Begin()

	if err := tx.Model(&RefreshToken{}).Where("token_hash = ?", hashedToken).Update("revoked_reason", reason).Update("revoked_at", time.Now()).Error; err != nil {
		tx.Rollback()
		log.Error().Err(err).Msg("failed to revoke refresh token")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		log.Error().Err(err).Msg("failed to commit transaction")
		tx.Rollback()
		return err
	}

	return nil
}
