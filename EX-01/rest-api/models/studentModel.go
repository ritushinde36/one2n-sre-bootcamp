package models

import (
	"gorm.io/gorm"
)

type Student struct {
	gorm.Model
	Name       string `json:"name"`
	Email      string `gorm:"unique" json:"email"`
	Age        int    `json:"age"`
	Class      string `json:"class"`
	Department string `json:"department"`
}
