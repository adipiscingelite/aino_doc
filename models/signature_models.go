package models

import (
	"database/sql"
	"time"
)

type Signatory struct {
	UUID     string `json:"sign_uuid" db:"sign_uuid"`
	Name     string `json:"name" db:"name"`
	Position string `json:"position" db:"position"`
	Role     string `json:"role_sign" db:"role_sign"`
}

type AddSignInfo struct {
	FormUUID string `json:"form_uuid" db:"form_uuid" validate:"required"`
	UserID   int    `json:"user_id" db:"user_id"`
	UUID     string `json:"sign_uuid" db:"sign_uuid"`
	Name     string `json:"name" db:"name" validate:"required"`
	Position string `json:"position" db:"position" validate:"required"`
	Role     string `json:"role_sign" db:"role_sign" validate:"required"`
}

type UpdateSignForm struct {
	UserID   int    `json:"user_id" db:"user_id"`
	UUID     string `json:"sign_uuid" db:"sign_uuid"`
	Name     string `json:"name" db:"name" validate:"required"`
	Position string `json:"position" db:"position" validate:"required"`
	Role     string `json:"role_sign" db:"role_sign" validate:"required"`
}

type Signatories struct {
	UUID       string         `json:"sign_uuid" db:"sign_uuid"`
	Name       string         `json:"name" db:"name"`
	Position   string         `json:"position" db:"position"`
	Role       string         `json:"role_sign" db:"role_sign"`
	IsSign     bool           `json:"is_sign" db:"is_sign"`
	Created_by sql.NullString `json:"created_by" db:"created_by"`
	Created_at time.Time      `json:"created_at" db:"created_at"`
	Updated_by sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type Signatorie struct {
	UUID       string         `json:"sign_uuid" db:"sign_uuid"`
	Name       string         `json:"name" db:"name"`
	Position   string         `json:"position" db:"position"`
	Role       string         `json:"role_sign" db:"role_sign"`
	IsSign     bool           `json:"is_sign" db:"is_sign"`
	Created_by sql.NullString `json:"created_by" db:"created_by"`
	Created_at time.Time      `json:"created_at" db:"created_at"`
	Updated_by sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type UpdateSign struct {
	IsSign     bool      `json:"is_sign" db:"is_sign" validate:"required"`
	Image      string    `json:"sign_img" db:"sign_img"`
	Updated_by string    `json:"updated_by" db:"updated_by"`
	Updated_at time.Time `json:"updated_at" db:"updated_at"`
}

type AddApproval struct {
	IsApproval bool      `json:"is_approve" db:"is_approve"`
	Reason     string    `json:"reason" db:"reason"`
	Updated_by string    `json:"updated_by" db:"updated_by"`
	Updated_at time.Time `json:"updated_at" db:"updated_at"`
}

type UserIDSign struct {
	UserID   int    `json:"user_id" db:"user_id"`
	SignUUID string `json:"sign_uuid" db:"sign_uuid"`
}

type SignatoryHA struct {
	SignUUID          string         `json:"sign_uuid" db:"sign_uuid"`
	SignatoryName     string         `json:"signatory_name" db:"signatory_name"`
	SignatoryPosition string         `json:"signatory_position" db:"signatory_position"`
	RoleSign          string         `json:"role_sign" db:"role_sign"`
	IsSign            bool           `json:"is_sign" db:"is_sign"`
	SignImg           sql.NullString `json:"sign_img" db:"sign_img"`
	Updated_at        sql.NullTime   `json:"updated_at" db:"updated_at"`
}

type Notif struct {
	FormUUID     string       `json:"form_uuid" db:"form_uuid"`
	FormNumber   string       `json:"form_number" db:"form_number"`
	FormTicket   string       `json:"form_ticket" db:"form_ticket" validate:"required"`
	FormStatus   string       `json:"form_status" db:"form_status"`
	DocumentCode string       `json:"document_code" db:"document_code"`
	DocumentName string       `json:"document_name" db:"document_name"`
	RoleSign     string       `json:"role_sign" db:"role_sign"`
	IsSign       bool         `json:"is_sign" db:"is_sign"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    sql.NullTime `json:"updated_at" db:"updated_at"`
	DeletedAt    sql.NullTime `json:"deleted_at" db:"deleted_at"`
}

type NotifApproval struct {
	FormUUID     string       `json:"form_uuid" db:"form_uuid"`
	FormNumber   string       `json:"form_number" db:"form_number"`
	FormTicket   string       `json:"form_ticket" db:"form_ticket" validate:"required"`
	IsApprove    string       `json:"is_approve" db:"is_approve"`
	DocumentCode string       `json:"document_code" db:"document_code"`
	DocumentName string       `json:"document_name" db:"document_name"`
	RoleSign     string       `json:"role_sign" db:"role_sign"`
	IsSign       bool         `json:"is_sign" db:"is_sign"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    sql.NullTime `json:"updated_at" db:"updated_at"`
	DeletedAt    sql.NullTime `json:"deleted_at" db:"deleted_at"`
}
