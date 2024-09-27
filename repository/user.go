package repository

import (
	"database/sql"
	"fmt"
	"pong-htmx/models"
	"time"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(user models.User) (models.User, error) {
	var returnUser models.User

	// Use the RETURNING clause to get the inserted row's id and other fields
	err := r.DB.QueryRow(
		`INSERT INTO users (username, email, provider, image_url, is_active, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, username, email, provider, image_url, is_active, is_verified`,
		user.Username, user.Email, user.Provider, user.ImageUrl, user.IsActive, user.IsVerified,
	).Scan(&returnUser.ID, &returnUser.Username, &returnUser.Email, &returnUser.Provider, &returnUser.ImageUrl, &returnUser.IsActive, &returnUser.IsVerified)

	if err != nil {
		return models.User{}, err
	}

	return returnUser, nil
}

func (r *UserRepository) GetByID(id int64) (models.User, error) {
	var user models.User
	err := r.DB.QueryRow("SELECT id, username, email, provider, image_url, is_active, is_verified, created_at, updated_at FROM users WHERE id = $1", id).
		Scan(&user.ID, &user.Username, &user.Email, &user.Provider, &user.ImageUrl, &user.IsActive, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(email string) (models.User, error) {
	var user models.User
	err := r.DB.QueryRow("SELECT id, username, email, provider, image_url, is_active, is_verified, created_at, updated_at FROM users WHERE email = $1", email).
		Scan(&user.ID, &user.Username, &user.Email, &user.Provider, &user.ImageUrl, &user.IsActive, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(username string) (models.User, error) {
	var user models.User
	err := r.DB.QueryRow("SELECT id, email, username, provider, image_url, is_active, is_verified, created_at, updated_at FROM users WHERE username = $1", username).
		Scan(&user.ID, &user.Email, &user.Username, &user.Provider, &user.ImageUrl, &user.IsActive, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	// Query the database
	rows, err := r.DB.Query("SELECT id, username, email, provider, image_url, is_active, is_verified FROM users")
	if err != nil {
		fmt.Printf("Query Error: %+v\n", err)
		return nil, err
	}
	defer rows.Close()

	// Prepare a slice to hold the users
	var users []models.User

	// Iterate through the result set
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Provider, &user.ImageUrl, &user.IsActive, &user.IsVerified); err != nil {
			fmt.Printf("Scan Error: %+v\n", err)
			return nil, err
		}
		fmt.Println("Fetched User:", user)
		users = append(users, user)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		fmt.Printf("Rows Error: %+v\n", err)
		return nil, err
	}

	fmt.Println("All Users:", users)
	return users, nil
}

func (r *UserRepository) Update(user models.User) error {
	updatedAt := time.Now()
	_, err := r.DB.Exec("UPDATE users SET username=$1, email=$2, provider=$3, image_url=$4, is_active=$5, is_verified=$6, updated_at= $7 WHERE id=$8",
		user.Username, user.Email, user.Provider, user.ImageUrl, user.IsActive, user.IsVerified, updatedAt, user.ID)
	return err
}

func (r *UserRepository) Delete(id int64) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = r.DB.Exec("DELETE FROM sessions WHERE user_id=$1", id)
	if err != nil {
		return err
	}
	_, err = r.DB.Exec("DELETE FROM users WHERE id=$1", id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *UserRepository) UpdateProfileImage(userId int64, imageUrl string, tx *sql.Tx) error {
	updatedAt := time.Now()
	query := `UPDATE users SET image_url = $1, updated_at = $2 WHERE id = $3 `
	if tx != nil {
		_, err := tx.Exec(query, imageUrl, updatedAt, userId)
		return err
	}
	_, err := r.DB.Exec(query, imageUrl, updatedAt, userId)
	return err
}
