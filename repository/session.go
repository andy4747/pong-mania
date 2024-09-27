package repository

import (
	"database/sql"
	"pong-htmx/models"
)

type SessionsRepository struct {
	db *sql.DB
}

func NewSessionsRepository(db *sql.DB) *SessionsRepository {
	return &SessionsRepository{db: db}
}

func (r *SessionsRepository) Create(session models.Session) error {
	_, err := r.db.Exec(`
	INSERT INTO sessions (user_id, token, expires_at)
	VALUES ($1, $2, $3)
	`, session.UserID, session.Token, session.ExpiresAt)
	return err
}

func (r *SessionsRepository) GetByToken(token string) (*models.Session, error) {
	var session models.Session
	err := r.db.QueryRow(`
	SELECT id, user_id, token, expires_at, created_at
	FROM sessions
	WHERE token = $1
	`, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionsRepository) Update(session models.Session) error {
	_, err := r.db.Exec("UPDATE sessions SET user_id=$1, token=$2, expires_at=$3 WHERE id=$4",
		session.UserID, session.Token, session.ExpiresAt, session.ID)
	return err
}

func (r *SessionsRepository) Delete(token string) error {
	query := `DELETE FROM sessions WHERE token = ?`
	_, err := r.db.Exec(query, token)
	return err
}

func (r *SessionsRepository) CleanupExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`
	_, err := r.db.Exec(query)
	return err
}
