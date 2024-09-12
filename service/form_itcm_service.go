package service

import (
	"database/sql"
	"document/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/google/uuid"
)

func generateFormNumberITCM(documentID int64, divisionCode string, recursionCount int) (string, error) {
	const maxRecursionCount = 1000

	// Check if the maximum recursion count is reached
	if recursionCount > maxRecursionCount {
		return "", errors.New("Maximum recursion count exceeded")
	}

	// Get the latest form number for the given document ID
	var latestFormNumber sql.NullString
	err := db.Get(&latestFormNumber, "SELECT MAX(form_number) FROM form_ms WHERE document_id = $1", documentID)
	if err != nil {
		return "", fmt.Errorf("Error getting latest form number: %v", err)
	}

	documentCode, err := GetDocumentCode(documentID)
	if err != nil {
		return "", fmt.Errorf("failed to get document code: %v", err)
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

	// Format the form number according to the specified format
	formNumberString := fmt.Sprintf("%04d", formNumber)
	formNumberWithDivision := fmt.Sprintf("%s/%s/%s/%s/%d", formNumberString, divisionCode, documentCode, romanMonth, year)
	// formNumberWithDivision := fmt.Sprintf("%s/%s/%s/%s/%d", formNumberString, "F", documentCode, romanMonth, year)

	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM form_ms WHERE form_number = $1 and document_id = $2", formNumberString, documentID)
	if err != nil {
		return "", fmt.Errorf("Error checking existing form number: %v", err)
	}
	if count > 0 {
		// If the form number already exists, recursively call the function again
		return generateFormNumberITCM(documentID, divisionCode, recursionCount+1)
	}

	return formNumberWithDivision, nil
}

func AddITCM(addForm models.Form, itcm models.ITCM, isPublished bool, userID int, username string, divisionCode string, recursionCount int, signatories []models.Signatory) error {
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

	formNumber, err := generateFormNumberITCM(documentID, divisionCode, recursionCount+1)
	if err != nil {
		log.Println("Error generating project form number:", err)
		return err
	}

	itcmJSON, err := json.Marshal(itcm)
	if err != nil {
		log.Println("Error marshaling ITCM struct:", err)
		return err
	}

	_, err = db.NamedExec("INSERT INTO form_ms (form_id, form_uuid, document_id, user_id, project_id, form_number, form_ticket, form_status, form_data, created_by) VALUES (:form_id, :form_uuid, :document_id, :user_id, :project_id,:form_number, :form_ticket, :form_status, :form_data, :created_by)", map[string]interface{}{
		"form_id":     appID,
		"form_uuid":   uuidString,
		"document_id": documentID,
		"user_id":     userID,
		"project_id":  projectID,
		"form_number": formNumber,
		"form_ticket": addForm.FormTicket,
		"form_status": formStatus,
		"form_data":   itcmJSON,
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

// menampilkan form tanpa token
func GetAllFormITCM() ([]models.FormsITCM, error) {
	rows, err := db.Query(`SELECT
		f.form_uuid,
		REPLACE(f.form_number, '/ITCM/', '/') AS formatted_form_number,
		f.form_ticket, f.form_status,
		d.document_name,
		p.project_name,
		p.project_manager,
		CASE
			WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
			WHEN f.is_approve = false THEN 'Tidak Disetujui'
			WHEN f.is_approve = true THEN 'Disetujui'
		END AS ApprovalStatus,
		f.reason, f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'no_da')::text AS no_da,
		(f.form_data->>'nama_pemohon')::text AS nama_pemohon,
		(f.form_data->>'instansi')::text AS instansi,
		(f.form_data->>'tanggal')::text AS tanggal,
		(f.form_data->>'perubahan_aset')::text AS perubahan_aset,
		(f.form_data->>'deskripsi')::text AS deskripsi
	FROM
		form_ms f
	LEFT JOIN
		document_ms d ON f.document_id = d.document_id
	LEFT JOIN
		project_ms p ON f.project_id = p.project_id
	WHERE
		d.document_code = 'ITCM' AND f.deleted_at IS NULL
	`)
	var forms []models.FormsITCM
	//rows, err := db.Query(&forms, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var form models.FormsITCM
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormTicket,
			&form.FormStatus,
			&form.DocumentName,
			&form.ProjectName,
			&form.ProjectManager,
			&form.ApprovalStatus,
			&form.Reason,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.NoDA,
			&form.NamaPemohon,
			&form.Intansi,
			&form.Tanggal,
			&form.PerubahanAset,
			&form.Deskripsi,
		)
		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}

	return forms, nil
}

// menampilkan form berdasar user/ milik dia sendiri
func GetAllITCMbyUserID(userID int) ([]models.FormsITCM, error) {
	rows, err := db.Query(`SELECT
		f.form_uuid,
		REPLACE(f.form_number, '/ITCM/', '/') AS formatted_form_number,
		f.form_ticket, f.form_status,
		d.document_name,
		p.project_name,
		p.project_manager,
		CASE
			WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
			WHEN f.is_approve = false THEN 'Tidak Disetujui'
			WHEN f.is_approve = true THEN 'Disetujui'
		END AS ApprovalStatus,
		f.reason, f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'no_da')::text AS no_da,
		(f.form_data->>'nama_pemohon')::text AS nama_pemohon,
		(f.form_data->>'instansi')::text AS instansi,
		(f.form_data->>'tanggal')::text AS tanggal,
		(f.form_data->>'perubahan_aset')::text AS perubahan_aset,
		(f.form_data->>'deskripsi')::text AS deskripsi
		FROM 
			form_ms f
		LEFT JOIN 
			document_ms d ON f.document_id = d.document_id
		LEFT JOIN 
			project_ms p ON f.project_id = p.project_id
		WHERE
			f.user_id = $1 AND d.document_code = 'ITCM' AND f.deleted_at IS NULL
			`, userID)
	var forms []models.FormsITCM
	//rows, err := db.Query(&forms, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var form models.FormsITCM
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormTicket,
			&form.FormStatus,
			&form.DocumentName,
			&form.ProjectName,
			&form.ProjectManager,
			&form.ApprovalStatus,
			&form.Reason,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.NoDA,
			&form.NamaPemohon,
			&form.Intansi,
			&form.Tanggal,
			&form.PerubahanAset,
			&form.Deskripsi,
		)
		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}

	return forms, nil

}

// menampilkan form dari admin
func GetAllFormITCMAdmin() ([]models.FormsITCM, error) {
	rows, err := db.Query(`SELECT
		f.form_uuid,
		REPLACE(f.form_number, '/ITCM/', '/') AS formatted_form_number,
		f.form_ticket, f.form_status,
		d.document_name,
		p.project_name,
		p.project_manager,
		CASE
			WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
			WHEN f.is_approve = false THEN 'Tidak Disetujui'
			WHEN f.is_approve = true THEN 'Disetujui'
		END AS ApprovalStatus,
		f.reason, f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'no_da')::text AS no_da,
		(f.form_data->>'nama_pemohon')::text AS nama_pemohon,
		(f.form_data->>'instansi')::text AS instansi,
		(f.form_data->>'tanggal')::text AS tanggal,
		(f.form_data->>'perubahan_aset')::text AS perubahan_aset,
		(f.form_data->>'deskripsi')::text AS deskripsi
	FROM
		form_ms f
	LEFT JOIN
		document_ms d ON f.document_id = d.document_id
	LEFT JOIN
		project_ms p ON f.project_id = p.project_id
	WHERE
		d.document_code = 'ITCM' AND f.deleted_at IS NULL
	`)
	var forms []models.FormsITCM
	//rows, err := db.Query(&forms, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var form models.FormsITCM
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormTicket,
			&form.FormStatus,
			&form.DocumentName,
			&form.ProjectName,
			&form.ProjectManager,
			&form.ApprovalStatus,
			&form.Reason,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.NoDA,
			&form.NamaPemohon,
			&form.Intansi,
			&form.Tanggal,
			&form.PerubahanAset,
			&form.Deskripsi,
		)
		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}

	return forms, nil
}

// untuk mengambil data spesifik hanya data form_ms
func GetSpecITCM(id string) (models.FormITCM, error) {
	var specITCM models.FormITCM

	err := db.Get(&specITCM, `SELECT 
	f.form_uuid, 
	REPLACE(f.form_number, '/ITCM/', '/') AS formatted_form_number,
	f.form_ticket, f.form_status,
	d.document_name,
	p.project_name,
	p.project_manager,
	CASE
		WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
		WHEN f.is_approve = false THEN 'Tidak Disetujui'
		WHEN f.is_approve = true THEN 'Disetujui'
	END AS ApprovalStatus,
	f.reason, f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
	(f.form_data->>'no_da')::text AS no_da,
	(f.form_data->>'nama_pemohon')::text AS nama_pemohon,
	(f.form_data->>'instansi')::text AS instansi,
	(f.form_data->>'tanggal')::text AS tanggal,
	(f.form_data->>'perubahan_aset')::text AS perubahan_aset,
	(f.form_data->>'deskripsi')::text AS deskripsi
FROM 
	form_ms f
LEFT JOIN 
	document_ms d ON f.document_id = d.document_id
LEFT JOIN 
	project_ms p ON f.project_id = p.project_id
WHERE
    f.form_uuid = $1 AND d.document_code = 'ITCM'  AND f.deleted_at IS NULL
	`, id)
	if err != nil {
		return models.FormITCM{}, err
	}

	return specITCM, nil

}

type FormITCMWithSignatories struct {
	Form        models.FormITCMAll   `json:"form"`
	Signatories []models.SignatoryHA `json:"signatories"`
}

// untuk mengambil data spesifik dari frm_ms dan sign_form
func GetSpecAllITCM(id string) (*FormITCMWithSignatories, error) {
	var formITCMWithSignatories FormITCMWithSignatories

	err := db.Get(&formITCMWithSignatories.Form, `
        SELECT
            f.form_uuid,
            REPLACE(f.form_number, '/ITCM/', '/') AS formatted_form_number,
            f.form_ticket,
            f.form_status,
            d.document_name,
            p.project_name,
            p.project_manager,
            CASE
                WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
                WHEN f.is_approve = false THEN 'Tidak Disetujui'
                WHEN f.is_approve = true THEN 'Disetujui'
            END AS ApprovalStatus,
            f.reason,
            f.created_by,
            f.created_at,
            f.updated_by,
            f.updated_at,
            f.deleted_by,
            f.deleted_at,
            (f.form_data->>'no_da')::text AS no_da,
            (f.form_data->>'nama_pemohon')::text AS nama_pemohon,
            (f.form_data->>'instansi')::text AS instansi,
            (f.form_data->>'tanggal')::text AS tanggal,
            (f.form_data->>'perubahan_aset')::text AS perubahan_aset,
            (f.form_data->>'deskripsi')::text AS deskripsi
        FROM
            form_ms f
        LEFT JOIN 
            document_ms d ON f.document_id = d.document_id
        LEFT JOIN 
            project_ms p ON f.project_id = p.project_id
        WHERE
            f.form_uuid = $1 
            AND d.document_code = 'ITCM'  
            AND f.deleted_at IS NULL
    `, id)

	if err != nil {
		return nil, err
	}

	err = db.Select(&formITCMWithSignatories.Signatories, `
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
                SELECT form_id 
                FROM form_ms 
                WHERE form_uuid = $1 
                AND deleted_at IS NULL
            )
    `, id)
	if err != nil {
		return nil, err
	}

	return &formITCMWithSignatories, nil
}

func DetailITCM(id string) ([]byte, error) {
	var formsWithSigners []models.DetailITCM
	type signerData struct {
		Signers []models.Signer `json:"signers"`
	}
	err := db.Select(&formsWithSigners, `
	SELECT
    f.form_uuid,
    REPLACE(f.form_number, '/ITCM/', '/') AS formatted_form_number,
    f.form_ticket,
    f.form_status,
    d.document_name,
    p.project_name,
    p.project_manager,
    CASE
        WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
        WHEN f.is_approve = false THEN 'Tidak Disetujui'
        WHEN f.is_approve = true THEN 'Disetujui'
    END AS ApprovalStatus,
    f.reason,
    f.created_by,
    f.created_at,
    f.updated_by,
    f.updated_at,
    f.deleted_by,
    f.deleted_at,
    (f.form_data->>'no_da')::text AS no_da,
    (f.form_data->>'nama_pemohon')::text AS nama_pemohon,
    (f.form_data->>'instansi')::text AS instansi,
    (f.form_data->>'tanggal')::text AS tanggal,
    (f.form_data->>'perubahan_aset')::text AS perubahan_aset,
    (f.form_data->>'deskripsi')::text AS deskripsi,
    array_agg(json_build_object(
        'sign_uuid', sf.sign_uuid,
        'name', sf.name,
        'position', sf.position,
        'role_sign', sf.role_sign
    )) AS signers
FROM
    form_ms f
LEFT JOIN 
    document_ms d ON f.document_id = d.document_id
LEFT JOIN 
    project_ms p ON f.project_id = p.project_id
LEFT JOIN
    sign_form sf ON f.form_id = sf.form_id
WHERE
    f.form_uuid = $1 AND d.document_code = 'ITCM'  AND f.deleted_at IS NULL
GROUP BY
f.form_uuid, f.form_number, f.form_ticket, f.form_status,  d.document_name,
p.project_name,
p.project_manager,
f.reason,
ApprovalStatus,
f.created_by,
f.created_at,
f.updated_by,
f.updated_at,
f.deleted_by,
f.deleted_at,
(f.form_data->>'no_da')::text,
(f.form_data->>'nama_pemohon')::text,
(f.form_data->>'instansi')::text,
(f.form_data->>'tanggal')::text,
(f.form_data->>'perubahan_aset')::text,
(f.form_data->>'deskripsi')::text
	`, id)

	if err != nil {
		return nil, err
	}

	// Loop through the results and unmarshal the signer data
	for i := range formsWithSigners {
		var signerDataObj signerData
		signerDataBytes, err := json.Marshal(formsWithSigners[i].Signers)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(signerDataBytes, &signerDataObj)
		if err != nil {
			return nil, err
		}

		// Set the unmarshaled signer data back to the formsWithSigners slice
		formsWithSigners[i].Signers = signerDataObj.Signers
	}

	jsonOutput, err := json.Marshal(formsWithSigners)
	if err != nil {
		return nil, err
	}

	return jsonOutput, nil
}

func UpdateFormITCM(updateITCM models.Form, data models.ITCM, username string, userID int, isPublished bool, id string, signatories []models.Signatory) (models.Form, error) {
	currentTime := time.Now()
	formStatus := "Draft"
	if isPublished {
		formStatus = "Published"
	}

	var projectID int64
	err := db.Get(&projectID, "SELECT project_id FROM project_ms WHERE project_uuid = $1", updateITCM.ProjectUUID)
	if err != nil {
		log.Println("Error getting project_id:", err)
		return models.Form{}, err
	}

	daJSON, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshaling DampakAnalisa struct:", err)
		return models.Form{}, err
	}
	log.Println("ITCM JSON:", string(daJSON))

	_, err = db.NamedExec("UPDATE form_ms SET user_id = :user_id, form_ticket = :form_ticket, form_status = :form_status, form_data = :form_data, updated_by = :updated_by, updated_at = :updated_at WHERE form_uuid = :id AND form_status = 'Draft'", map[string]interface{}{
		"user_id":     userID,
		"form_ticket": updateITCM.FormTicket,
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

	var formID string
	err = db.Get(&formID, "SELECT form_id FROM form_ms WHERE form_uuid = $1", id)
	if err != nil {
		log.Println("Error getting form_id:", err)
		return models.Form{}, err
	}

	_, err = db.Exec("DELETE FROM sign_form WHERE form_id = $1", formID)
	if err != nil {
		log.Println("Error deleting sign_form records:", err)
		return models.Form{}, err
	}

	personalNames, err := GetAllPersonalName()
	if err != nil {
		log.Println("Error getting personal names:", err)
		return models.Form{}, err
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
			"form_id":    formID, // Adjusted to use documentID
			"name":       signatory.Name,
			"position":   signatory.Position,
			"role_sign":  signatory.Role,
			"created_by": username,
		})
		if err != nil {
			return models.Form{}, err
		}
	}

	return updateITCM, nil

}

// gakepake itcma nya
func UpdateFormITCMa(updateITCM models.Form, data models.ITCM, username string, userID int, isPublished bool, id string, signatories []models.UpdateSignForm) (models.Form, error) {
	currentTime := time.Now()
	formStatus := "Draft"
	if isPublished {
		formStatus = "Published"
	}

	var projectID int64
	err := db.Get(&projectID, "SELECT project_id FROM project_ms WHERE project_uuid = $1", updateITCM.ProjectUUID)
	if err != nil {
		log.Println("Error getting project_id:", err)
		return models.Form{}, err
	}

	daJSON, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshaling DampakAnalisa struct:", err)
		return models.Form{}, err
	}
	log.Println("ITCM JSON:", string(daJSON))

	_, err = db.NamedExec("UPDATE form_ms SET user_id = :user_id, form_ticket = :form_ticket, form_status = :form_status, form_data = :form_data, updated_by = :updated_by, updated_at = :updated_at WHERE form_uuid = :id AND form_status = 'Draft'", map[string]interface{}{
		"user_id":     userID,
		"form_ticket": updateITCM.FormTicket,
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

	// Update or insert sign_form
	personalNames, err := GetAllPersonalName() // Mengambil daftar semua personal name
	if err != nil {
		log.Println("Error getting personal names:", err)
		return models.Form{}, err
	}

	for _, signatory := range signatories {

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

		// Mendapatkan daftar UUID yang baru dari signatories
		var uuids []string
		for _, signatory := range signatories {
			uuids = append(uuids, signatory.UUID)
		}
		// Mendapatkan form_id dari form_uuid
		var formID int64
		err = db.Get(&formID, "SELECT form_id FROM form_ms WHERE form_uuid = $1", id)
		if err != nil {
			return models.Form{}, err
		}
		// Lakukan penghapusan tanda tangan yang tidak relevan
		_, err = db.Exec("DELETE FROM sign_form WHERE form_id = $1 AND sign_uuid NOT IN ($2, $3)", formID, uuids[0], uuids[1])
		if err != nil {
			return models.Form{}, err
		}
	}
	return updateITCM, nil
}

func FormITCMByDivision(divisionCode string) ([]models.FormsITCM, error) {
	var form []models.FormsITCM

	// Now use the retrieved documentID in the query

	// REPLACE(f.form_number, '/ITCM/', '/') AS formatted_form_number,
	errSelect := db.Select(&form, `
		SELECT
		f.form_uuid,
		f.form_number,
		f.form_ticket, f.form_status,
		d.document_name,
		p.project_name,
		p.project_manager,
		CASE
			WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
			WHEN f.is_approve = false THEN 'Tidak Disetujui'
			WHEN f.is_approve = true THEN 'Disetujui'
		END AS ApprovalStatus,
		f.reason, f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'no_da')::text AS no_da,
		(f.form_data->>'nama_pemohon')::text AS nama_pemohon,
		(f.form_data->>'instansi')::text AS instansi,
		(f.form_data->>'tanggal')::text AS tanggal,
		(f.form_data->>'perubahan_aset')::text AS perubahan_aset,
		(f.form_data->>'deskripsi')::text AS deskripsi
	FROM
		form_ms f
	LEFT JOIN
		document_ms d ON f.document_id = d.document_id
	LEFT JOIN
		project_ms p ON f.project_id = p.project_id
	WHERE
		d.document_code = 'ITCM' AND f.deleted_at IS NULL AND SPLIT_PART(f.form_number, '/', 2) = $1
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

// menampilkan form berdasar user/ milik dia sendiri
func SignatureUserITCM(userID int) ([]models.FormsITCM, error) {
	rows, err := db.Query(`SELECT
		f.form_uuid,
		REPLACE(f.form_number, '/ITCM/', '/') AS formatted_form_number,
		f.form_ticket, f.form_status,
		d.document_name,
		p.project_name,
		p.project_manager,
		CASE
			WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
			WHEN f.is_approve = false THEN 'Tidak Disetujui'
			WHEN f.is_approve = true THEN 'Disetujui'
		END AS ApprovalStatus,
		f.reason, f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'no_da')::text AS no_da,
		(f.form_data->>'nama_pemohon')::text AS nama_pemohon,
		(f.form_data->>'instansi')::text AS instansi,
		(f.form_data->>'tanggal')::text AS tanggal,
		(f.form_data->>'perubahan_aset')::text AS perubahan_aset,
		(f.form_data->>'deskripsi')::text AS deskripsi
		FROM 
		form_ms f
	LEFT JOIN 
		document_ms d ON f.document_id = d.document_id
	LEFT JOIN 
		project_ms p ON f.project_id = p.project_id
	LEFT JOIN 
		sign_form sf ON f.form_id = sf.form_id
	WHERE
			sf.user_id = $1 AND d.document_code = 'ITCM' AND f.deleted_at IS NULL
			`, userID)
	var forms []models.FormsITCM
	//rows, err := db.Query(&forms, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var form models.FormsITCM
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormTicket,
			&form.FormStatus,
			&form.DocumentName,
			&form.ProjectName,
			&form.ProjectManager,
			&form.ApprovalStatus,
			&form.Reason,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.NoDA,
			&form.NamaPemohon,
			&form.Intansi,
			&form.Tanggal,
			&form.PerubahanAset,
			&form.Deskripsi,
		)
		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}

	return forms, nil

}

func GetITCMCode() (models.DocCodeName, error) {
	var documentCode models.DocCodeName

	err := db.Get(&documentCode, "SELECT document_uuid FROM document_ms WHERE document_code = 'ITCM'")

	if err != nil {
		return models.DocCodeName{}, err
	}
	return documentCode, nil
}
