package db

import (
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

var (
	ErrUserExists = errors.New("User already exists")
)

func (db *AuthDatabase) RegisterUser(username, login, hashedPassword string) (string, error) {
	registration := &Registration{
		Username:       username,
		Login:          login,
		HashedPassword: hashedPassword,
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

func (db *AuthDatabase) StoreRefreshToken(userID string, tokenHash []byte) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse user ID")
		return err
	}

	refreshToken := &RefreshToken{
		UserID:    userUUID,
		TokenHash: tokenHash,
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
