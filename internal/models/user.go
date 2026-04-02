package models

import "time"

type User struct {
	ID        string
	Name      string
	Email     string
	Password  string // hashed
	Phone     string
	CreatedAt time.Time
}
