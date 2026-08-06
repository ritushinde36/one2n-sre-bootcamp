package models

import (
	"time"

	"gorm.io/gorm"
)

type Student struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Name       string `gorm:"not null;size:255" json:"name" binding:"required"`
	Email      string `gorm:"unique;not null;size:255" json:"email" binding:"required,email"`
	Age        int    `gorm:"not null" json:"age" binding:"required,gt=0"`
	Class      string `gorm:"not null;size:255" json:"class" binding:"required"`
	Department string `gorm:"not null;size:255" json:"department" binding:"required"`
}
