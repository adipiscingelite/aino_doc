package service

import (
	"document/models"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
)

func AddBA(addForm models.Form, ba models.BA, isPublished bool, userID int, username string, divisionCode string, recursionCount int, signatories []models.Signatory) error {
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

	var projectID int64
	err = db.Get(&projectID, "SELECT project_id FROM project_ms WHERE project_uuid = $1", addForm.ProjectUUID)
	if err != nil {
		log.Println("Error getting project_id:", err)
		return err
	}

	var documentCode string
	err = db.Get(&documentCode, "SELECT document_code FROM document_ms WHERE document_uuid = $1", addForm.DocumentUUID)
	if err != nil {
		log.Println("Error getting document code:", err)
		return err
	}

	formNumber, err := generateFormNumber(documentID, divisionCode, recursionCount+1)
	if err != nil {
		log.Println("Error generating project form number:", err)
		return err
	}

	// Marshal ITCM struct to JSON
	baJSON, err := json.Marshal(ba)
	if err != nil {
		log.Println("Error marshaling ITCM struct:", err)
		return err
	}

	_, err = db.NamedExec("INSERT INTO form_ms (form_id, form_uuid, document_id, user_id, project_id, form_number, form_ticket, form_status, form_data, created_by) VALUES (:form_id, :form_uuid, :document_id, :user_id, :project_id, :form_number, :form_ticket, :form_status, :form_data, :created_by)", map[string]interface{}{
		"form_id":     appID,
		"form_uuid":   uuidString,
		"document_id": documentID,
		"user_id":     userID,
		"project_id":  projectID,
		"form_number": formNumber,
		"form_ticket": addForm.FormTicket,
		"form_status": formStatus,
		"form_data":   baJSON, // Convert JSON to string
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

func GetAllFormBA() ([]models.FormsBA, error) {
	rows, err := db.Query(`
		SELECT 
			f.form_uuid,  f.form_number, f.form_ticket, f.form_status,
			d.document_name,
			p.project_name,
			f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
			(f.form_data->>'judul')::text AS judul,
			(f.form_data->>'tanggal')::text AS tanggal,
			(f.form_data->>'nama_aplikasi')::text AS nama_aplikasi,
			(f.form_data->>'no_da')::text AS no_da,
			(f.form_data->>'no_itcm')::text AS no_itcm,
			(f.form_data->>'dilakukan_oleh')::text AS dilakukan_oleh,
			(f.form_data->>'didampingi_oleh')::text AS didampingi_oleh
			FROM 
			form_ms f
		LEFT JOIN 
			document_ms d ON f.document_id = d.document_id
		LEFT JOIN 
			project_ms p ON f.project_id = p.project_id
			WHERE
			d.document_code = 'BA' AND f.deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold all form data
	var forms []models.FormsBA

	// Iterate through the rows
	for rows.Next() {
		// Scan the row into the Forms struct
		var form models.FormsBA
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormTicket,
			&form.FormStatus,
			&form.DocumentName,
			&form.ProjectName,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.Judul,
			&form.Tanggal,
			&form.AppName,
			&form.NoDA,
			&form.NoITCM,
			&form.DilakukanOleh,
			&form.DidampingiOleh,
		)
		if err != nil {
			return nil, err
		}

		// Append the form data to the slice
		forms = append(forms, form)
	}
	// Return the forms as JSON response
	return forms, nil
}

func GetAllBAbyUserID(userID int) ([]models.FormsBA, error) {
	rows, err := db.Query(`
		SELECT 
			f.form_uuid,  f.form_number, f.form_ticket, f.form_status,
			d.document_name,
			p.project_name,
			f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
			(f.form_data->>'judul')::text AS judul,
			(f.form_data->>'tanggal')::text AS tanggal,
			(f.form_data->>'nama_aplikasi')::text AS nama_aplikasi,
			(f.form_data->>'no_da')::text AS no_da,
			(f.form_data->>'no_itcm')::text AS no_itcm,
			(f.form_data->>'dilakukan_oleh')::text AS dilakukan_oleh,
			(f.form_data->>'didampingi_oleh')::text AS didampingi_oleh
			FROM 
			form_ms f
		LEFT JOIN 
			document_ms d ON f.document_id = d.document_id
		LEFT JOIN 
			project_ms p ON f.project_id = p.project_id
			WHERE
			f.user_id = $1 AND d.document_code = 'BA' AND f.deleted_at IS NULL
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold all form data
	var forms []models.FormsBA

	// Iterate through the rows
	for rows.Next() {
		// Scan the row into the Forms struct
		var form models.FormsBA
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormTicket,
			&form.FormStatus,
			&form.DocumentName,
			&form.ProjectName,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.Judul,
			&form.Tanggal,
			&form.AppName,
			&form.NoDA,
			&form.NoITCM,
			&form.DilakukanOleh,
			&form.DidampingiOleh,
		)
		if err != nil {
			return nil, err
		}

		// Append the form data to the slice
		forms = append(forms, form)
	}
	// Return the forms as JSON response
	return forms, nil
}

func GetAllBAbyAdmin() ([]models.FormsBA, error) {
	rows, err := db.Query(`
		SELECT 
			f.form_uuid, f.form_number, f.form_ticket, f.form_status,
			d.document_name,
			p.project_name,
			f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
			(f.form_data->>'judul')::text AS judul,
			(f.form_data->>'tanggal')::text AS tanggal,
			(f.form_data->>'nama_aplikasi')::text AS nama_aplikasi,
			(f.form_data->>'no_da')::text AS no_da,
			(f.form_data->>'no_itcm')::text AS no_itcm,
			(f.form_data->>'dilakukan_oleh')::text AS dilakukan_oleh,
			(f.form_data->>'didampingi_oleh')::text AS didampingi_oleh
			FROM 
			form_ms f
		LEFT JOIN 
			document_ms d ON f.document_id = d.document_id
		LEFT JOIN 
			project_ms p ON f.project_id = p.project_id
			WHERE
			d.document_code = 'BA' AND f.deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold all form data
	var forms []models.FormsBA

	// Iterate through the rows
	for rows.Next() {
		// Scan the row into the Forms struct
		var form models.FormsBA
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormTicket,
			&form.FormStatus,
			&form.DocumentName,
			&form.ProjectName,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.Judul,
			&form.Tanggal,
			&form.AppName,
			&form.NoDA,
			&form.NoITCM,
			&form.DilakukanOleh,
			&form.DidampingiOleh,
		)
		if err != nil {
			return nil, err
		}

		// Append the form data to the slice
		forms = append(forms, form)
	}
	// Return the forms as JSON response
	return forms, nil
}

func GetSpecBA(id string) (models.FormsBA, error) {
	var specBA models.FormsBA
	err := db.Get(&specBA, `SELECT 
	f.form_uuid,f.form_number, f.form_ticket, f.form_status,
	d.document_name,
	p.project_name,
	f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
	(f.form_data->>'judul')::text AS judul,
	(f.form_data->>'tanggal')::text AS tanggal,
	(f.form_data->>'nama_aplikasi')::text AS nama_aplikasi,
	(f.form_data->>'no_da')::text AS no_da,
	(f.form_data->>'no_itcm')::text AS no_itcm,
	(f.form_data->>'dilakukan_oleh')::text AS dilakukan_oleh,
	(f.form_data->>'didampingi_oleh')::text AS didampingi_oleh
	FROM 
	form_ms f
LEFT JOIN 
	document_ms d ON f.document_id = d.document_id
LEFT JOIN 
	project_ms p ON f.project_id = p.project_id
	WHERE
	f.form_uuid = $1 AND d.document_code = 'BA'  AND f.deleted_at IS NULL
	`, id)

	if err != nil {
		return models.FormsBA{}, err
	}

	return specBA, nil
}
func GetSpecAllBA(id string) ([]models.FormsBAAll, error) {
	var forms []models.FormsBAAll

	err := db.Select(&forms, `SELECT 
	f.form_uuid, f.form_number, f.form_ticket, f.form_status,
	d.document_name,
	p.project_name,
	f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
	(f.form_data->>'judul')::text AS judul,
	(f.form_data->>'tanggal')::text AS tanggal,
	(f.form_data->>'nama_aplikasi')::text AS nama_aplikasi,
	(f.form_data->>'no_da')::text AS no_da,
	(f.form_data->>'no_itcm')::text AS no_itcm,
	(f.form_data->>'dilakukan_oleh')::text AS dilakukan_oleh,
	(f.form_data->>'didampingi_oleh')::text AS didampingi_oleh,
	sf.sign_uuid AS sign_uuid,
    sf.name AS name,
    sf.position AS position,
    sf.role_sign AS role_sign,
	sf.is_sign AS is_sign
	FROM
    form_ms f
LEFT JOIN 
    document_ms d ON f.document_id = d.document_id
LEFT JOIN 
    project_ms p ON f.project_id = p.project_id
LEFT JOIN
    sign_form sf ON f.form_id = sf.form_id
WHERE
    f.form_uuid = $1 AND d.document_code = 'BA'  AND f.deleted_at IS NULL
	`, id)

	if err != nil {
		return nil, err
	}
	return forms, nil
}

func UpdateBA(updateBA models.Form, data models.BA, username string, userID int, isPublished bool, id string) (models.Form, error) {
	currentTime := time.Now()
	formStatus := "Draft"
	if isPublished {
		formStatus = "Published"
	}

	var projectID int64
	err := db.Get(&projectID, "SELECT project_id FROM project_ms WHERE project_uuid = $1", updateBA.ProjectUUID)
	if err != nil {
		log.Println("Error getting project_id:", err)
		return models.Form{}, err
	}

	daJSON, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshaling DampakAnalisa struct:", err)
		return models.Form{}, err
	}
	log.Println("DampakAnalisa JSON:", string(daJSON)) // Periksa hasil marshaling

	_, err = db.NamedExec("UPDATE form_ms SET user_id = :user_id, form_ticket = :form_ticket, form_status = :form_status, form_data = :form_data, updated_by = :updated_by, updated_at = :updated_at WHERE form_uuid = :id AND form_status = 'Draft'", map[string]interface{}{
		"user_id":     userID,
		"form_ticket": updateBA.FormTicket,
		"project_id":  projectID,
		"form_status": formStatus,
		"form_data":   daJSON,
		"updated_by":  username,
		"updated_at":  currentTime,
		"id":          id,
	})
	if err != nil {
		return models.Form{}, err
	}
	return updateBA, nil
}

// menampilkan formulir sesuai dengan nama signature user tersebut. required signature
func SignatureUserBA(userID int) ([]models.FormsBA, error) {
	rows, err := db.Query(`
		SELECT 
			f.form_uuid,  f.form_number, f.form_ticket, f.form_status,
			d.document_name,
			p.project_name,
			f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
			(f.form_data->>'judul')::text AS judul,
			(f.form_data->>'tanggal')::text AS tanggal,
			(f.form_data->>'nama_aplikasi')::text AS nama_aplikasi,
			(f.form_data->>'no_da')::text AS no_da,
			(f.form_data->>'no_itcm')::text AS no_itcm,
			(f.form_data->>'dilakukan_oleh')::text AS dilakukan_oleh,
			(f.form_data->>'didampingi_oleh')::text AS didampingi_oleh
			FROM 
			form_ms f
		LEFT JOIN 
			document_ms d ON f.document_id = d.document_id
		LEFT JOIN 
			project_ms p ON f.project_id = p.project_id
			LEFT JOIN 
		sign_form sf ON f.form_id = sf.form_id
		WHERE
		sf.user_id = $1 AND d.document_code = 'DA'  AND f.deleted_at IS NULL
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold all form data
	var forms []models.FormsBA

	// Iterate through the rows
	for rows.Next() {
		// Scan the row into the Forms struct
		var form models.FormsBA
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormTicket,
			&form.FormStatus,
			&form.DocumentName,
			&form.ProjectName,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.Judul,
			&form.Tanggal,
			&form.AppName,
			&form.NoDA,
			&form.NoITCM,
			&form.DilakukanOleh,
			&form.DidampingiOleh,
		)
		if err != nil {
			return nil, err
		}

		// Append the form data to the slice
		forms = append(forms, form)
	}
	// Return the forms as JSON response
	return forms, nil
}
