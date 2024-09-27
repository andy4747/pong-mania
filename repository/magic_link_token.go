package repository

import (
	"database/sql"
	"fmt"
	"pong-htmx/models"
	"time"
)

type MagicLinkTokenRepository struct {
	db *sql.DB
}

func NewMagicLinkTokenRepository(db *sql.DB) *MagicLinkTokenRepository {
	return &MagicLinkTokenRepository{
		db: db,
	}
}

func (r *MagicLinkTokenRepository) CreateMagicLinkToken(email string, token string, expiry time.Time) error {
	query := `INSERT INTO magic_link_tokens (email, token, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, email, token, expiry)
	return err
}

func (r *MagicLinkTokenRepository) GetMagicLinkToken(token string) (models.MagicLinkToken, error) {
	query := `SELECT id, email, token, expires_at FROM magic_link_tokens WHERE token = $1`
	row := r.db.QueryRow(query, token)

	var magicLinkRecord models.MagicLinkToken
	err := row.Scan(&magicLinkRecord.ID, &magicLinkRecord.Email, &magicLinkRecord.Token, &magicLinkRecord.ExpiresAt)
	if err != nil {
		return models.MagicLinkToken{}, err
	}
	return magicLinkRecord, nil
}

func (r *MagicLinkTokenRepository) GetByToken(token string) (models.MagicLinkToken, error) {
	var magicLinkRecord models.MagicLinkToken
	err := r.db.QueryRow("SELECT id, email, token, expires_at FROM magic_link_tokens WHERE token = $1", token).
		Scan(&magicLinkRecord.ID, &magicLinkRecord.Email, &magicLinkRecord.Token, &magicLinkRecord.ExpiresAt)
	if err != nil {
		fmt.Println(err)
		return models.MagicLinkToken{}, err
	}
	return magicLinkRecord, nil
}

func (r *MagicLinkTokenRepository) DeleteMagicLinkRecord(token string) error {
	_, err := r.db.Exec(`DELETE FROM magic_link_tokens WHERE token=$1`, token)
	if err != nil {
		return err
	}
	return nil
}
