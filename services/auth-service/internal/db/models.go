// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

type PasswordResetToken struct {
	ID        int32        `json:"id"`
	UserID    *int32       `json:"user_id"`
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	UsedAt    sql.NullTime `json:"used_at"`
	CreatedAt time.Time    `json:"created_at"`
}

type User struct {
	ID            int32           `json:"id"`
	Email         string          `json:"email"`
	PasswordHash  string          `json:"password_hash"`
	FirstName     string          `json:"first_name"`
	LastName      string          `json:"last_name"`
	Role          string          `json:"role"`
	Permissions   json.RawMessage `json:"permissions"`
	Active        *bool           `json:"active"`
	EmailVerified *bool           `json:"email_verified"`
	LastLogin     sql.NullTime    `json:"last_login"`
	CreatedBy     *int32          `json:"created_by"`
	UpdatedBy     *int32          `json:"updated_by"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     sql.NullTime    `json:"deleted_at"`
}
