package models

import "time"

type User struct {
	ID           int64
	Mail         string
	PasswordHash string
	Role         int32
	IsActive     bool
	CreatedAt    time.Time
}
