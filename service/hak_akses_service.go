package service

import (
	"database/sql"
	"document/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

func generateFormNumberHA(documentID int64, divisionCode string, recursionCount int) (string, error) {
	const maxRecursionCount = 1000

	// Check if the maximum recursion count is reached
	if recursionCount > maxRecursionCount {
		return "", errors.New("Maximum recursion count exceeded")
	}

	documentCode, err := GetDocumentCode(documentID)
	if err != nil {
		return "", fmt.Errorf("Failed to get document code: %v", err)
	}

	// Get the latest form number for the given document ID
	var latestFormNumber sql.NullString
	err = db.Get(&latestFormNumber, "SELECT MAX(form_number) FROM form_ms WHERE document_id = $1", documentID)
	if err != nil {
		return "", fmt.Errorf("Error getting latest form number: %v", err)
	}

	// Initialize formNumber to 1 if latestFormNumber is NULL
	formNumber := 1
	if latestFormNumber.Valid {
		// Parse the latest form number
		var latestFormNumberInt int
		_, err := fmt.Sscanf(latestFormNumber.String, "%d", &latestFormNumberInt)
		if err != nil {
			return "", fmt.Errorf("Error parsing latest form number: %v", err)
		}
		// Increment the latest form number
		formNumber = latestFormNumberInt + 1
	}

	// Get current year and month
	year := time.Now().Year()
	month := time.Now().Month()

	// Convert month to Roman numeral
	romanMonth, err := convertToRoman(int(month))
	if err != nil {
		return "", fmt.Errorf("Error converting month to Roman numeral: %v", err)
	}

	fmt.Println("latest", latestFormNumber)
	fmt.Println("document code", documentCode)

	// Format the form number according to the specified format
	formNumberString := fmt.Sprintf("%04d", formNumber)
	formNumberWithDivision := fmt.Sprintf("%s/%s/%s/%s/%d", formNumberString, divisionCode, documentCode, romanMonth, year)
	// formNumberWithDivision := fmt.Sprintf("%s/%s/%s/%s/%d", formNumberString, "PED", "F", romanMonth, year)

	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM form_ms WHERE form_number = $1 and document_id = $2", formNumberString, documentID)
	if err != nil {
		return "", fmt.Errorf("Error checking existing form number: %v", err)
	}
	if count > 0 {
		// If the form number already exists, recursively call the function again
		return generateFormNumberHA(documentID, divisionCode, recursionCount+1)
	}

	fmt.Println(formNumberWithDivision)
	return formNumberWithDivision, nil
}

func AddHakAkses(addForm models.FormHA, infoHA []models.AddInfoHA, ha models.HA, isPublished bool, userID int, divisionCode string, recrusionCount int, username string, signatories []models.Signatory) error {
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

	formNumberHA, err := generateFormNumberHA(documentID, divisionCode, recrusionCount+1)
	if err != nil {
		// Handle error
		log.Println("Error generating form number:", err)
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
		"form_number": formNumberHA,
		"form_ticket": addForm.FormTicket,
		"form_status": formStatus,
		"form_data":   formData, // Convert JSON to string
		"created_by":  username,
	})

	// fmt.Println(formNumberHA)

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
		f.form_uuid,
		f.form_number,
		f.form_ticket,
		f.form_status,
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
			&form.FormNumber,
			&form.FormTicket,
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
	Form        models.FormsHA        `json:"form"`
	InfoHA      []models.HakAksesInfo `json:"hak_akses_info"`
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
	err = db.Select(&formWithSignatories.InfoHA, `
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

func UpdateHakAkses(id string, username string, ha models.HA, signatories []models.Signatory, info_ha []models.HakAksesInfo) error {
	log.Printf("Updating HA with ID: %s, User: %s, Data: %+v", id, username, ha)
	currentTime := time.Now()

	fmt.Println("Info ha", info_ha)

	// Marshaling data HA menjadi JSON
	formData, err := json.Marshal(ha)
	if err != nil {
		log.Println("Error marshaling HA struct:", err)
		return err
	}
	log.Println("HA JSON:", string(formData))

	// Menjalankan query UPDATE
	result, err := db.Exec("UPDATE form_ms SET form_data = $1, updated_at = $2, updated_by = $3 WHERE form_uuid = $4",
		formData, currentTime, username, id)
	if err != nil {
		log.Println("Error executing query:", err)
		return err
	}

	var formID string
	err = db.Get(&formID, "SELECT form_id FROM form_ms WHERE form_uuid = $1", id)
	if err != nil {
		log.Println("Error getting form_id:", err)
		return err // Return only the error
	}

	_, err = db.Exec("DELETE FROM sign_form WHERE form_id = $1", formID)
	if err != nil {
		log.Println("Error deleting sign_form records:", err)
		return err // Return only the error
	}

	personalNames, err := GetAllPersonalName()
	if err != nil {
		log.Println("Error getting personal names:", err)
		return err // Return only the error
	}

	for _, signatory := range signatories {
		uuidString := uuid.New().String()

		log.Printf("Processing signatory: %+v\n", signatory)
		var userID string
		for _, personal := range personalNames {
			if personal.PersonalName == signatory.Name {
				userID = personal.UserID
				break
			}
		}

		if userID == "" {
			log.Printf("User ID not found for personal name: %s\n", signatory.Name)
			continue
		}

		_, err := db.NamedExec("INSERT INTO sign_form (sign_uuid, form_id, user_id, name, position, role_sign, created_by) VALUES (:sign_uuid, :form_id, :user_id, :name, :position, :role_sign, :created_by)", map[string]interface{}{
			"sign_uuid":  uuidString,
			"user_id":    userID,
			"form_id":    formID,
			"name":       signatory.Name,
			"position":   signatory.Position,
			"role_sign":  signatory.Role,
			"created_by": username,
		})
		if err != nil {
			return err // Return only the error
		}
	}

	_, err = db.Exec("DELETE FROM hak_akses_info WHERE form_id = $1", formID)
	if err != nil {
		log.Println("Error deleting sign_form records:", err)
		return err // Return only the error
	}

	for _, info_ha := range info_ha {
		uuidString := uuid.New().String()
		_, err := db.NamedExec(`INSERT INTO hak_akses_info (info_uuid, form_id, name, instansi, position, username, password, scope, created_by, created_at, updated_by, updated_at) VALUES (:info_uuid, :form_id, :name, :instansi, :position, :username, :password, :scope, :created_by, :created_at, :updated_by, :updated_at)`,
			map[string]interface{}{
				"info_uuid":  uuidString,
				"form_id":    formID,
				"name":       info_ha.InfoName,
				"instansi":   info_ha.Instansi,
				"position":   info_ha.Position,
				"username":   info_ha.Username,
				"password":   info_ha.Password,
				"scope":      info_ha.Scope,
				"created_by": username,
				"created_at": time.Now(),
				"updated_by": username,
				"updated_at": time.Now(),
			})
		if err != nil {
			log.Println("Error inserting record:", err)
			return err
		}

		if err != nil {
			log.Println("Error inserting record:", err)
			return err
		}
	}

	// Memeriksa jumlah baris yang diperbarui
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Error getting rows affected:", err)
		return err
	}
	if rowsAffected == 0 {
		log.Println("No rows updated. Check if the ID exists and is correct.")
	}

	return nil // Return nil on success
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
		f.form_uuid,
		f.form_number,
		f.form_ticket,
		f.form_status,
		d.document_name,
		f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'form_name')::text AS form_name
	FROM
		form_ms f
	LEFT JOIN
		document_ms d ON f.document_id = d.document_id
	WHERE
	f.user_id = $1 AND d.document_code = 'HA' AND  f.deleted_at IS NULL
	`, userID)
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
			&form.FormNumber,
			&form.FormTicket,
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
	var forms []models.FormsHA
	query := `SELECT
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
	`

	// Assuming 'db' is an *sqlx.DB instance
	err := db.Select(&forms, query)
	if err != nil {
		return nil, err
	}

	return forms, nil
}

// menampilkan form berdasar user/ milik dia sendiri
func SignatureUserHA(userID int) ([]models.FormsHA, error) {
	rows, err := db.Query(`SELECT
		f.form_uuid,
		f.form_number,
		f.form_status,
		d.document_name,
		CASE
			WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
			WHEN f.is_approve = false THEN 'Tidak Disetujui'
			WHEN f.is_approve = true THEN 'Disetujui'
		END AS ApprovalStatus,
		f.reason, f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'form_name')::text AS form_name
		FROM 
		form_ms f
	LEFT JOIN 
		document_ms d ON f.document_id = d.document_id
	LEFT JOIN 
		sign_form sf ON f.form_id = sf.form_id
	WHERE
			sf.user_id = $1 AND d.document_code = 'HA' AND f.deleted_at IS NULL
			`, userID)
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
			&form.FormNumber,
			&form.FormStatus,
			&form.DocumentName,
			&form.ApprovalStatus,
			&form.Reason,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.FormName, // pastikan urutan form_name disesuaikan dengan urutan yang ada
		)

		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}

	return forms, nil

}

func GetHACode() (models.DocCodeName, error) {
	var documentCode models.DocCodeName

	err := db.Get(&documentCode, "SELECT document_uuid FROM document_ms WHERE document_code = 'HA'")

	if err != nil {
		return models.DocCodeName{}, err
	}
	return documentCode, nil
}

func FormHAByDivision(divisionCode string) ([]models.FormsHA, error) {
	var form []models.FormsHA

	// Now use the retrieved documentID in the query
	errSelect := db.Select(&form, `
			SELECT 
			f.form_uuid,
			f.form_number,
			f.form_ticket,
			f.form_status,
			f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
			d.document_name,
			(f.form_data->>'form_name')::text AS form_name,
			CASE
				WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
				WHEN f.is_approve = false THEN 'Tidak Disetujui'
				WHEN f.is_approve = true THEN 'Disetujui'
			END AS approval_status -- Alias the CASE expression as ApprovalStatus
			FROM 
			form_ms f
		LEFT JOIN 
			document_ms d ON f.document_id = d.document_id
			WHERE
			d.document_code = 'HA' AND f.deleted_at IS NULL AND SPLIT_PART(f.form_number, '/', 2) = $1
	`, divisionCode)

	if errSelect != nil {
		log.Print(errSelect)
		return nil, errSelect
	}

	if len(form) == 0 {
		return nil, sql.ErrNoRows
	}

	return form, nil
}
