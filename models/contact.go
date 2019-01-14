package models

import (
	"github.com/jinzhu/gorm"
)

/*
Contact A user's contact
*/
type Contact struct {
	gorm.Model
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	UserID string `json:"user_id"`
}
