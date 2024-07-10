package models

import (
	"database/sql"
	"time"
)

type Personal struct {
	UserID       string `json:"user_id" db:"user_id"`
	PersonalName string `json:"personal_name" db:"personal_name"`
}

type ITCM struct {
	NoDA          string `json:"no_da" validate:"required"`
	NamaPemohon   string `json:"nama_pemohon" validate:"required"`
	Intansi       string `json:"instansi" validate:"required"`
	Tanggal       string `json:"tanggal" validate:"required"`
	PerubahanAset string `json:"perubahan_aset" validate:"required"`
	Deskripsi     string `json:"deskripsi" validate:"required"`
}

type FormsITCM struct {
	FormUUID       string         `json:"form_uuid" db:"form_uuid"`
	FormNumber     string         `json:"form_number" db:"form_number"`
	FormTicket     string         `json:"form_ticket" db:"form_ticket"`
	FormStatus     string         `json:"form_status" db:"form_status"`
	DocumentName   string         `json:"document_name" db:"document_name"`
	ProjectName    string         `json:"project_name" db:"project_name"`
	ProjectManager string         `json:"project_manager" db:"project_manager"`
	ApprovalStatus string         `json:"approval_status" db:"approvalstatus"`
	Reason         sql.NullString `json:"reason" db:"reason"`
	CreatedBy      string         `json:"created_by" db:"created_by"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedBy      sql.NullString `json:"updated_by" db:"updated_by"`
	UpdatedAt      sql.NullTime   `json:"updated_at" db:"updated_at"`
	DeletedBy      sql.NullString `json:"deleted_by" db:"deleted_by"`
	DeletedAt      sql.NullTime   `json:"deleted_at" db:"deleted_at"`
	NoDA           string         `json:"no_da" db:"no_da"`
	NamaPemohon    string         `json:"nama_pemohon" db:"nama_pemohon"`
	Intansi        string         `json:"instansi" db:"instansi"`
	Tanggal        string         `json:"tanggal" db:"tanggal"`
	PerubahanAset  string         `json:"perubahan_aset" db:"perubahan_aset"`
	Deskripsi      string         `json:"deskripsi" db:"deskripsi"`
}

type FormITCM struct {
	FormUUID            string         `json:"form_uuid" db:"form_uuid"`
	FormattedFormNumber string         `json:"formatted_form_number" db:"formatted_form_number"`
	FormTicket          string         `json:"form_ticket" db:"form_ticket"`
	FormStatus          string         `json:"form_status" db:"form_status"`
	DocumentName        string         `json:"document_name" db:"document_name"`
	ProjectName         string         `json:"project_name" db:"project_name"`
	ProjectManager      string         `json:"project_manager" db:"project_manager"`
	ApprovalStatus      string         `json:"approval_status" db:"approvalstatus"`
	Reason              sql.NullString `json:"reason" db:"reason"`
	CreatedBy           string         `json:"created_by" db:"created_by"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedBy           sql.NullString `json:"updated_by" db:"updated_by"`
	UpdatedAt           sql.NullTime   `json:"updated_at" db:"updated_at"`
	DeletedBy           sql.NullString `json:"deleted_by" db:"deleted_by"`
	DeletedAt           sql.NullTime   `json:"deleted_at" db:"deleted_at"`
	NoDA                string         `json:"no_da" db:"no_da"`
	NamaPemohon         string         `json:"nama_pemohon" db:"nama_pemohon"`
	Intansi             string         `json:"instansi" db:"instansi"`
	Tanggal             string         `json:"tanggal" db:"tanggal"`
	PerubahanAset       string         `json:"perubahan_aset" db:"perubahan_aset"`
	Deskripsi           string         `json:"deskripsi" db:"deskripsi"`
}

type FormITCMAll struct {
	FormUUID            string         `json:"form_uuid" db:"form_uuid"`
	FormattedFormNumber string         `json:"formatted_form_number" db:"formatted_form_number"`
	FormTicket          string         `json:"form_ticket" db:"form_ticket"`
	FormStatus          string         `json:"form_status" db:"form_status"`
	DocumentName        string         `json:"document_name" db:"document_name"`
	ProjectName         string         `json:"project_name" db:"project_name"`
	ProjectManager      string         `json:"project_manager" db:"project_manager"`
	ApprovalStatus      string         `json:"approval_status" db:"approvalstatus"`
	Reason              sql.NullString `json:"reason" db:"reason"`
	CreatedBy           string         `json:"created_by" db:"created_by"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedBy           sql.NullString `json:"updated_by" db:"updated_by"`
	UpdatedAt           sql.NullTime   `json:"updated_at" db:"updated_at"`
	DeletedBy           sql.NullString `json:"deleted_by" db:"deleted_by"`
	DeletedAt           sql.NullTime   `json:"deleted_at" db:"deleted_at"`
	NoDA                string         `json:"no_da" db:"no_da"`
	NamaPemohon         string         `json:"nama_pemohon" db:"nama_pemohon"`
	Intansi             string         `json:"instansi" db:"instansi"`
	Tanggal             string         `json:"tanggal" db:"tanggal"`
	PerubahanAset       string         `json:"perubahan_aset" db:"perubahan_aset"`
	Deskripsi           string         `json:"deskripsi" db:"deskripsi"`
	UUID                string         `json:"sign_uuid" db:"sign_uuid"`
	Name                string         `json:"name" db:"name"`
	Position            string         `json:"position" db:"position"`
	Role                string         `json:"role_sign" db:"role_sign"`
	IsSign              bool           `json:"is_sign" db:"is_sign"`
}

type DetailITCM struct {
	FormUUID            string         `json:"form_uuid" db:"form_uuid"`
	FormattedFormNumber string         `json:"formatted_form_number" db:"formatted_form_number"`
	FormTicket          string         `json:"form_ticket" db:"form_ticket"`
	FormStatus          string         `json:"form_status" db:"form_status"`
	DocumentName        string         `json:"document_name" db:"document_name"`
	ProjectName         string         `json:"project_name" db:"project_name"`
	ProjectManager      string         `json:"project_manager" db:"project_manager"`
	ApprovalStatus      string         `json:"approval_status" db:"approvalstatus"`
	Reason              sql.NullString `json:"reason" db:"reason"`
	CreatedBy           string         `json:"created_by" db:"created_by"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedBy           sql.NullString `json:"updated_by" db:"updated_by"`
	UpdatedAt           sql.NullTime   `json:"updated_at" db:"updated_at"`
	DeletedBy           sql.NullString `json:"deleted_by" db:"deleted_by"`
	DeletedAt           sql.NullTime   `json:"deleted_at" db:"deleted_at"`
	NoDA                string         `json:"no_da" db:"no_da"`
	NamaPemohon         string         `json:"nama_pemohon" db:"nama_pemohon"`
	Intansi             string         `json:"instansi" db:"instansi"`
	Tanggal             string         `json:"tanggal" db:"tanggal"`
	PerubahanAset       string         `json:"perubahan_aset" db:"perubahan_aset"`
	Deskripsi           string         `json:"deskripsi" db:"deskripsi"`
	Signers             []Signer       `json:"signers"`
	// SignUUID            string         `json:"sign_uuid" db:"sign_uuid"`
	// Name                string         `json:"name" db:"name"`
	// Position            string         `json:"position" db:"position"`
	// RoleSign            string         `json:"role_sign" db:"role_sign"`
}

type Signer struct {
	SignUUID string `json:"sign_uuid" db:"sign_uuid"`
	Name     string `json:"name" db:"name"`
	Position string `json:"position" db:"position"`
	RoleSign string `json:"role_sign" db:"role_sign"`
}
