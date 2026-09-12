package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID  `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Status      int        `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
}

const (
	Category_ID          = "id"
	Category_Code        = "code"
	Category_Name        = "name"
	Category_Description = "description"
	Category_Status      = "status"
)
