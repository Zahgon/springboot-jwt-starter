// Package repository holds the persistence access that Spring Data JPA
// generated from the UserRepository interface.
package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/bfwg/springboot-jwt-starter/internal/model"
)

// UserRepository reads and writes user accounts.
type UserRepository struct {
	db *sql.DB
}

// New returns a repository backed by the given database.
func New(database *sql.DB) *UserRepository { return &UserRepository{db: database} }

const selectUserColumns = `
    SELECT id, username, password, first_name, last_name, email, phone_number,
           enabled, last_password_reset_date
      FROM USERS`

// FindByUsername returns the account with the given username, or nil when there
// is none.
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	return r.queryOne(selectUserColumns+" WHERE username = ?", username)
}

// FindByID returns the account with the given id, or nil when there is none.
func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	return r.queryOne(selectUserColumns+" WHERE id = ?", id)
}

// FindAll returns every account, ordered by id ascending.
func (r *UserRepository) FindAll() ([]*model.User, error) {
	rows, err := r.db.Query(selectUserColumns + " ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("repository: listing users: %w", err)
	}
	defer rows.Close()

	users := []*model.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: listing users: %w", err)
	}
	for _, u := range users {
		if err := r.loadAuthorities(u); err != nil {
			return nil, err
		}
	}
	return users, nil
}

// Save persists a modified account.
func (r *UserRepository) Save(u *model.User) error {
	_, err := r.db.Exec(
		`UPDATE USERS
            SET username = ?, password = ?, first_name = ?, last_name = ?,
                email = ?, phone_number = ?, enabled = ?, last_password_reset_date = ?
          WHERE id = ?`,
		u.Username, u.Password, u.FirstName, u.LastName, u.Email, u.PhoneNumber,
		u.Enabled, u.LastPasswordResetDate.String(), u.ID)
	if err != nil {
		return fmt.Errorf("repository: saving user %d: %w", u.ID, err)
	}
	return nil
}

func (r *UserRepository) queryOne(query string, args ...any) (*model.User, error) {
	u, err := scanUser(r.db.QueryRow(query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.loadAuthorities(u); err != nil {
		return nil, err
	}
	return u, nil
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface{ Scan(dest ...any) error }

func scanUser(s scanner) (*model.User, error) {
	var (
		u       model.User
		resetAt sql.NullString
	)
	err := s.Scan(&u.ID, &u.Username, &u.Password, &u.FirstName, &u.LastName,
		&u.Email, &u.PhoneNumber, &u.Enabled, &resetAt)
	if err != nil {
		return nil, err
	}
	if resetAt.Valid {
		ts, err := model.ParseTimestamp(resetAt.String)
		if err != nil {
			return nil, fmt.Errorf("repository: user %d: %w", u.ID, err)
		}
		u.LastPasswordResetDate = ts
	}
	return &u, nil
}

// loadAuthorities fills in the roles joined through USER_AUTHORITY, which the
// original fetched eagerly via @ManyToMany(fetch = FetchType.EAGER).
func (r *UserRepository) loadAuthorities(u *model.User) error {
	rows, err := r.db.Query(
		`SELECT a.id, a.name
           FROM AUTHORITY a
           JOIN USER_AUTHORITY ua ON ua.authority_id = a.id
          WHERE ua.user_id = ?
          ORDER BY a.id`, u.ID)
	if err != nil {
		return fmt.Errorf("repository: loading authorities for user %d: %w", u.ID, err)
	}
	defer rows.Close()

	u.Authorities = []model.Authority{}
	for rows.Next() {
		var a model.Authority
		if err := rows.Scan(&a.ID, &a.Name); err != nil {
			return fmt.Errorf("repository: loading authorities for user %d: %w", u.ID, err)
		}
		u.Authorities = append(u.Authorities, a)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("repository: loading authorities for user %d: %w", u.ID, err)
	}
	return nil
}
