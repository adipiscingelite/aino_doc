package models

import (
	"database/sql"
	"time"
)

type Project struct {
	UUID           string         `json:"project_uuid" db:"project_uuid"`
	ProductUUID    string         `json:"product_uuid" db:"product_uuid"`
	ProductID      int            `json:"product_id" db:"product_id"`
	ProjectName    string         `json:"project_name" db:"project_name"`
	ProjectCode    string         `json:"project_code" db:"project_code"`
	ProjectManager string         `json:"project_manager" db:"project_manager"`
	Order          int            `json:"project_order" db:"project_order"`
	Created_by     string         `json:"created_by" db:"created_by"`
	Created_at     time.Time      `json:"created_at" db:"created_at"`
	Updated_by     sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at     sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by     sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at     sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}

type Projects struct {
	UUID           string         `json:"project_uuid" db:"project_uuid"`
	ProductName    string         `json:"product_name" db:"product_name"`
	ProjectName    string         `json:"project_name" db:"project_name"`
	ProjectCode    string         `json:"project_code" db:"project_code"`
	ProjectManager string         `json:"project_manager" db:"project_manager"`
	Order          int            `json:"project_order" db:"project_order"`
	Created_by     string         `json:"created_by" db:"created_by"`
	Created_at     time.Time      `json:"created_at" db:"created_at"`
	Updated_by     sql.NullString `json:"updated_by" db:"updated_by"`
	Updated_at     sql.NullTime   `json:"updated_at" db:"updated_at"`
	Deleted_by     sql.NullString `json:"deleted_by" db:"deleted_by"`
	Deleted_at     sql.NullTime   `json:"deleted_at" db:"deleted_at"`
}
type ProjectCodeName struct {
	UUID        string `json:"project_uuid" db:"project_uuid"`
	ProjectCode string `json:"project_code" db:"project_code"`
	ProjectName string `json:"project_name" db:"project_name"`
}
