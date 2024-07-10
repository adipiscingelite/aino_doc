package models

import (
	"database/sql"
	"time"
)

type FormHA struct {
	UUID         string         `json:"form_uuid" db:"form_uuid"`
	DocumentUUID string         `json:"document_uuid" db:"document_uuid"`
	DocumentID   int64          `json:"document_id" db:"document_id"`
	UserID       int            `json:"user_id" db:"user_id" validate:"required"`
	FormNumber   string         `json:"form_number" db:"form_number"`
	FormTicket   string         `json:"form_ticket" db:"form_ticket"`
	FormStatus   string         `json:"form_status" db:"form_status"`
	Created_by   string         `json:"created_by" db:"created_by"`
	Created_at   time.Time      `json:"created_at" db:"created_at"`
	Updated_by   sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at   sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by   sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at   sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type AddInfoHA struct {
	UUID     string `json:"form_uuid" db:"form_uuid"`
	Name     string `json:"name" db:"name"`
	Instansi string `json:"instansi" db:"instansi"`
	Position string `json:"position" db:"position"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Scope    string `json:"scope" db:"scope"`
}

type FormsHA struct {
	FormUUID     string         `json:"form_uuid" db:"form_uuid"`
	DocumentName string         `json:"document_name" db:"document_name"`
	FormName     string         `json:"form_name" db:"form_name"`
	FormStatus   string         `json:"form_status" db:"form_status"`
	CreatedBy    string         `json:"created_by" db:"created_by"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedBy    sql.NullString `json:"updated_by" db:"updated_by"`
	UpdatedAt    sql.NullTime   `json:"updated_at" db:"updated_at"`
	DeletedBy    sql.NullString `json:"deleted_by" db:"deleted_by"`
	DeletedAt    sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type HA struct {
	FormName string `json:"form_name" db:"form_name"`
}
type FormsHAAll struct {
	FormUUID      string         `json:"form_uuid" db:"form_uuid"`
	FormStatus    string         `json:"form_status" db:"form_status"`
	DocumentName  string         `json:"document_name" db:"document_name"`
	CreatedBy     string         `json:"created_by" db:"created_by"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedBy     sql.NullString `json:"updated_by" db:"updated_by"`
	UpdatedAt     sql.NullTime   `json:"updated_at" db:"updated_at"`
	DeletedBy     sql.NullString `json:"deleted_by" db:"deleted_by"`
	DeletedAt     sql.NullTime   `json:"deleted_at" db:"deleted_at"`
	FormName      string         `json:"form_name" db:"form_name"`
	InfoUUID      string         `json:"info_uuid" db:"info_uuid"`
	InfoName      string         `json:"info_name" db:"info_name"` // Ubah nama field dari "name" menjadi "info_name"
	Instansi      string         `json:"instansi" db:"instansi"`
	InfoPosition  string         `json:"position" db:"position"`
	Username      string         `json:"username" db:"username"`
	Password      string         `json:"password" db:"password"`
	Scope         string         `json:"scope" db:"scope"`
	UUID          string         `json:"sign_uuid" db:"sign_uuid"`
	SignatoryName string         `json:"signatory_name" db:"signatory_name"`         // Ubah nama field dari "name" menjadi "signatory_name"
	Position      string         `json:"signatory_position" db:"signatory_position"` // Ubah nama field dari "position" menjadi "signatory_position"
	Role          string         `json:"role_sign" db:"role_sign"`
	IsSign        bool           `json:"is_sign" db:"is_sign"`
}

type HakAksesInfo struct {
	InfoUUID string `json:"info_uuid" db:"info_uuid"`
	InfoName string `json:"info_name" db:"info_name"`
	Instansi string `json:"instansi" db:"instansi"`
	Position string `json:"position" db:"position"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Scope    string `json:"scope" db:"scope"`
}
