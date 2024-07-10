package service

import (
	"document/models"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
)

func AddHakAkses(addForm models.FormHA, infoHA []models.AddInfoHA, ha models.HA, isPublished bool, userID int, username string, signatories []models.Signatory) error {
	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()
	appID := currentTimestamp + int64(uniqueID)
	uuidObj := uuid.New()
	uuidString := uuidObj.String()

	formStatus := "Draft"
	if isPublished {
		formStatus = "Published"
	}

	var documentID int64
	err := db.Get(&documentID, "SELECT document_id FROM document_ms WHERE document_uuid = $1", addForm.DocumentUUID)
	if err != nil {
		log.Println("Error getting document_id:", err)
		return err
	}
	formData, err := json.Marshal(ha)
	if err != nil {
		log.Println("Error marshaling ITCM struct:", err)
		return err
	}
	_, err = db.NamedExec("INSERT INTO form_ms (form_id, form_uuid, document_id, user_id, project_id, form_number, form_ticket, form_status, form_data, created_by) VALUES (:form_id, :form_uuid, :document_id, :user_id, :project_id, :form_number, :form_ticket, :form_status, :form_data, :created_by)", map[string]interface{}{
		"form_id":     appID,
		"form_uuid":   uuidString,
		"document_id": documentID,
		"user_id":     userID,
		"project_id":  nil,
		"form_number": " ",
		"form_ticket": " ",
		"form_status": formStatus,
		"form_data":   formData, // Convert JSON to string
		"created_by":  username,
	})

	if err != nil {
		return err
	}
	personalNames, err := GetAllPersonalName() // Mengambil daftar semua personal name
	if err != nil {
		log.Println("Error getting personal names:", err)
		return err
	}

	for _, info := range infoHA {
		uuidString := uuid.New().String()

		_, err := db.NamedExec("INSERT INTO hak_akses_info (info_uuid, form_id, name, instansi, position, username, password, scope, created_by) VALUES (:info_uuid, :form_id, :name, :instansi, :position, :username, :password, :scope, :created_by)", map[string]interface{}{
			"info_uuid":  uuidString,
			"form_id":    appID,
			"name":       info.Name,
			"instansi":   info.Instansi,
			"position":   info.Position,
			"username":   info.Username,
			"password":   info.Password,
			"scope":      info.Scope,
			"created_by": username,
		})
		if err != nil {
			return err
		}
	}

	for _, signatory := range signatories {
		uuidString := uuid.New().String()

		// Mencari user_id yang sesuai dengan personal_name yang dipilih
		var userID string
		for _, personal := range personalNames {
			if personal.PersonalName == signatory.Name {
				userID = personal.UserID
				break
			}
		}

		// Memastikan user_id ditemukan untuk personal_name yang dipilih
		if userID == "" {
			log.Printf("User ID not found for personal name: %s\n", signatory.Name)
			continue
		}

		_, err := db.NamedExec("INSERT INTO sign_form (sign_uuid, form_id, user_id, name, position, role_sign, created_by) VALUES (:sign_uuid, :form_id, :user_id, :name, :position, :role_sign, :created_by)", map[string]interface{}{
			"sign_uuid":  uuidString,
			"user_id":    userID,
			"form_id":    appID,
			"name":       signatory.Name,
			"position":   signatory.Position,
			"role_sign":  signatory.Role,
			"created_by": username,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func GetAllHakAkses() ([]models.FormsHA, error) {
	rows, err := db.Query(`SELECT
		f.form_uuid, f.form_status,                                                                                               
		d.document_name,
		f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'form_name')::text AS form_name
	FROM
		form_ms f
	LEFT JOIN
		document_ms d ON f.document_id = d.document_id
	WHERE
		d.document_code = 'HA' AND f.deleted_at IS NULL
	`)
	var forms []models.FormsHA
	//rows, err := db.Query(&forms, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var form models.FormsHA
		err := rows.Scan(
			&form.FormUUID,
			&form.FormStatus,
			&form.DocumentName,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.FormName,
		)
		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}

	return forms, nil
}

type FormWithSignatories struct {
	Form        models.FormsHAAll     `json:"form"`
	Info        []models.HakAksesInfo `json:"hak_akses_info"`
	Signatories []models.SignatoryHA  `json:"signatories"`
}

func GetSpecAllHA(id string) (*FormWithSignatories, error) {
	var formWithSignatories FormWithSignatories

	// Ambil data form
	err := db.Get(&formWithSignatories.Form, `
        SELECT 
            f.form_uuid,
            f.form_status,
            d.document_name,
            f.created_by,
            f.created_at,
            f.updated_by,
            f.updated_at,
            f.deleted_by,
            f.deleted_at,
            (f.form_data->>'form_name')::text AS form_name
        FROM
            form_ms f
        LEFT JOIN 
            document_ms d ON f.document_id = d.document_id
        WHERE
            f.form_uuid = $1 AND d.document_code = 'HA' AND f.deleted_at IS NULL
    `, id)
	if err != nil {
		return nil, err
	}

	// Ambil data hak akses info
	err = db.Select(&formWithSignatories.Info, `
        SELECT 
            info_uuid,
            name AS info_name,
            instansi,
            position,
            username,
            password,
            scope
        FROM
            hak_akses_info
        WHERE
            form_id IN (
                SELECT form_id FROM form_ms WHERE form_uuid = $1 AND deleted_at IS NULL
            )
    `, id)
	if err != nil {
		return nil, err
	}

	// Ambil data signatories
	err = db.Select(&formWithSignatories.Signatories, `
        SELECT 
            sign_uuid,
            name AS signatory_name,
            position AS signatory_position,
            role_sign,
            is_sign
        FROM
            sign_form
        WHERE
            form_id IN (
                SELECT form_id FROM form_ms WHERE form_uuid = $1 AND deleted_at IS NULL
            )
    `, id)
	if err != nil {
		return nil, err
	}

	return &formWithSignatories, nil
}

func GetSpecHakAkses(id string) (models.FormsHA, error) {
	var specHA models.FormsHA

	err := db.Get(&specHA, `SELECT 
	f.form_uuid,
	f.form_status,
	d.document_name,
	f.created_by,
	f.created_at,
	f.updated_by,
	f.updated_at,
	f.deleted_by,
	f.deleted_at,
	(f.form_data->>'form_name')::text AS form_name
FROM
	form_ms f
LEFT JOIN 
	document_ms d ON f.document_id = d.document_id
WHERE
	f.form_uuid = $1 AND d.document_code = 'HA' AND f.deleted_at IS NULL
	`, id)
	if err != nil {
		return models.FormsHA{}, err
	}

	return specHA, nil

}

func UpdateHakAkses(id string, username string, ha models.HA) error {

	formData, err := json.Marshal(ha)
	if err != nil {
		log.Println("Error marshaling ITCM struct:", err)
		return err
	}
	_, err = db.Exec("UPDATE form_ms SET form_data = $1, updated_at = NOW(), updated_by = $2, WHERE form_uuid = $3", formData, username, id)
	if err != nil {
		return err
	}
	return nil
}

func GetInfoHA(id string) ([]models.HakAksesInfo, error) {
	var infoHA []models.HakAksesInfo
	err := db.Select(&infoHA, `SELECT 
	info_uuid,
	name AS info_name,
	instansi,
	position,
	username,
	password,
	scope
FROM
	hak_akses_info
WHERE
	form_id IN (
		SELECT form_id FROM form_ms WHERE form_uuid = $1 AND deleted_at IS NULL
	)
`, id)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return infoHA, nil
}

func MyFormHA(userID int) ([]models.FormsHA, error) {
	rows, err := db.Query(`SELECT
		f.form_uuid, f.form_status,                                                                                               
		d.document_name,
		f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'form_name')::text AS form_name
	FROM
		form_ms f
	LEFT JOIN
		document_ms d ON f.document_id = d.document_id
	WHERE
	f.user_id = $1 AND d.document_code = 'HA' AND  f.deleted_at IS NULL
	`)
	var forms []models.FormsHA
	//rows, err := db.Query(&forms, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var form models.FormsHA
		err := rows.Scan(
			&form.FormUUID,
			&form.FormStatus,
			&form.DocumentName,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.FormName,
		)
		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}

	return forms, nil
}

func GetFormsByAdmin() ([]models.FormsHA, error) {
	rows, err := db.Query(`SELECT
		f.form_uuid, f.form_status,                                                                                               
		d.document_name,
		f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'form_name')::text AS form_name
	FROM
		form_ms f
	LEFT JOIN
		document_ms d ON f.document_id = d.document_id
	WHERE
		d.document_code = 'HA' AND f.deleted_at IS NULL
	`)
	var forms []models.FormsHA
	//rows, err := db.Query(&forms, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var form models.FormsHA
		err := rows.Scan(
			&form.FormUUID,
			&form.FormStatus,
			&form.DocumentName,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.FormName,
		)
		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}

	return forms, nil
}
