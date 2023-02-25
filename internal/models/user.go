package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Login    string `gorm:"unique" json:"login"`
	Password string `json:"password"`

	Current   float64 `json:"current,omitempty"`
	Withdrawn int64   `json:"withdrawn,omitempty"`
}