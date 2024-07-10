package models

import (
	"database/sql"
	"time"
)

type DampakAnalisa struct {
	//NamaProyek                           string    `json:"nama_proyek"`
	NamaAnalis                           string `json:"nama_analis"`
	Jabatan                              string `json:"jabatan"`
	Departemen                           string `json:"departemen"`
	JenisPerubahan                       string `json:"jenis_perubahan"`
	DetailDampakPerubahan                string `json:"detail_dampak_perubahan"`
	RencanaPengembanganPerubahan         string `json:"rencana_pengembangan_perubahan"`
	RencanaPengujianPerubahanSistem      string `json:"rencana_pengujian_perubahan_sistem"`
	RencanaRilisPerubahanDanImplementasi string `json:"rencana_rilis_perubahan_dan_implementasi"`
}

type Formss struct {
	FormUUID                             string         `json:"form_uuid" db:"form_uuid"`
	FormNumber                           string         `json:"form_number" db:"form_number"`
	FormTicket                           string         `json:"form_ticket" db:"form_ticket"`
	FormStatus                           string         `json:"form_status" db:"form_status"`
	DocumentName                         string         `json:"document_name" db:"document_name"`
	ProjectName                          string         `json:"project_name" db:"project_name"`
	ApprovalStatus                       string         `json:"approval_status"`
	Reason                               sql.NullString `json:"reason" db:"reason"`
	CreatedBy                            string         `json:"created_by" db:"created_by"`
	CreatedAt                            time.Time      `json:"created_at" db:"created_at"`
	UpdatedBy                            sql.NullString `json:"updated_by" db:"updated_by"`
	UpdatedAt                            sql.NullTime   `json:"updated_at" db:"updated_at"`
	DeletedBy                            sql.NullString `json:"deleted_by" db:"deleted_by"`
	DeletedAt                            sql.NullTime   `json:"deleted_at" db:"deleted_at"`
	NamaAnalis                           string         `json:"nama_analis" db:"nama_analis"`
	Jabatan                              string         `json:"jabatan" db:"jabatan"`
	Departemen                           string         `json:"departemen" db:"departemen"`
	JenisPerubahan                       string         `json:"jenis_perubahan" db:"jenis_perubahan"`
	DetailDampakPerubahan                string         `json:"detail_dampak_perubahan" db:"detail_dampak_perubahan"`
	RencanaPengembanganPerubahan         string         `json:"rencana_pengembangan_perubahan" db:"rencana_pengembangan_perubahan"`
	RencanaPengujianPerubahanSistem      string         `json:"rencana_pengujian_perubahan_sistem" db:"rencana_pengujian_perubahan_sistem"`
	RencanaRilisPerubahanDanImplementasi string         `json:"rencana_rilis_perubahan_dan_implementasi" db:"rencana_rilis_perubahan_dan_implementasi"`
}

type FormsDAAll struct {
	FormUUID                             string         `json:"form_uuid" db:"form_uuid"`
	FormNumber                           string         `json:"form_number" db:"form_number"`
	FormTicket                           string         `json:"form_ticket" db:"form_ticket"`
	FormStatus                           string         `json:"form_status" db:"form_status"`
	DocumentName                         string         `json:"document_name" db:"document_name"`
	ProjectName                          string         `json:"project_name" db:"project_name"`
	ApprovalStatus                       string         `json:"approval_status"`
	Reason                               sql.NullString `json:"reason" db:"reason"`
	CreatedBy                            string         `json:"created_by" db:"created_by"`
	CreatedAt                            time.Time      `json:"created_at" db:"created_at"`
	UpdatedBy                            sql.NullString `json:"updated_by" db:"updated_by"`
	UpdatedAt                            sql.NullTime   `json:"updated_at" db:"updated_at"`
	DeletedBy                            sql.NullString `json:"deleted_by" db:"deleted_by"`
	DeletedAt                            sql.NullTime   `json:"deleted_at" db:"deleted_at"`
	NamaAnalis                           string         `json:"nama_analis" db:"nama_analis"`
	Jabatan                              string         `json:"jabatan" db:"jabatan"`
	Departemen                           string         `json:"departemen" db:"departemen"`
	JenisPerubahan                       string         `json:"jenis_perubahan" db:"jenis_perubahan"`
	DetailDampakPerubahan                string         `json:"detail_dampak_perubahan" db:"detail_dampak_perubahan"`
	RencanaPengembanganPerubahan         string         `json:"rencana_pengembangan_perubahan" db:"rencana_pengembangan_perubahan"`
	RencanaPengujianPerubahanSistem      string         `json:"rencana_pengujian_perubahan_sistem" db:"rencana_pengujian_perubahan_sistem"`
	RencanaRilisPerubahanDanImplementasi string         `json:"rencana_rilis_perubahan_dan_implementasi" db:"rencana_rilis_perubahan_dan_implementasi"`
	UUID                                 string         `json:"sign_uuid" db:"sign_uuid"`
	Name                                 string         `json:"name" db:"name"`
	Position                             string         `json:"position" db:"position"`
	Role                                 string         `json:"role_sign" db:"role_sign"`
	IsSign                               bool           `json:"is_sign" db:"is_sign"`
}
