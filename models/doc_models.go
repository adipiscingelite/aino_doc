package models

import (
	"database/sql"
	"time"
)

type Document struct {
	UUID       string         `json:"document_uuid" db:"document_uuid"`
	Code       string         `json:"document_code" db:"document_code" validate:"required"`
	Name       string         `json:"document_name" db:"document_name" validate:"required"`
	Order      int            `json:"document_order" db:"document_order"`
	Created_by string         `json:"created_by" db:"created_by"`
	Created_at time.Time      `json:"created_at" db:"created_at"`
	Updated_by sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type DocCodeName struct {
	UUID string `json:"document_uuid" db:"document_uuid"`
	Code string `json:"document_code" db:"document_code"`
	Name string `json:"document_name" db:"document_name"`
}
