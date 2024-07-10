package models

import (
	"database/sql"
	"time"
)

type Product struct {
	UUID string `json:"product_uuid" db:"product_uuid"`
	//UserID       int            `json:"user_id" db:"user_id" validate:"required"`
	ProductName  string         `json:"product_name" db:"product_name" validate:"required"`
	ProductOwner string         `json:"product_owner" db:"product_owner"`
	Order        int            `json:"product_order" db:"product_order"`
	Created_by   string         `json:"created_by" db:"created_by"`
	Created_at   time.Time      `json:"created_at" db:"created_at"`
	Updated_by   sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at   sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by   sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at   sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type ProductName struct {
	UUID string `json:"product_uuid" db:"product_uuid"`
	Name string `json:"product_name" db:"product_name"`
}
