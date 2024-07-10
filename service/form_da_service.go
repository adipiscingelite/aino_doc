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

func generateFormNumberDA(documentID int64, recursionCount int) (string, error) {
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
	formNumberWithDivision := fmt.Sprintf("%s/%s/%s/%s/%d", formNumberString, "PED", "F", romanMonth, year)

	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM form_ms WHERE form_number = $1 and document_id = $2", formNumberString, documentID)
	if err != nil {
		return "", fmt.Errorf("Error checking existing form number: %v", err)
	}
	if count > 0 {
		// If the form number already exists, recursively call the function again
		return generateFormNumberDA(documentID, recursionCount+1)
	}

	return formNumberWithDivision, nil
}

func AddDA(addDA models.Form, isPublished bool, username string, userID int, divisionCode string, recrusionCount int, data models.DampakAnalisa, signatories []models.Signatory) error {
	var documentCode string
	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()

	app_id := currentTimestamp + int64(uniqueID)

	uuidObj := uuid.New()
	uuidString := uuidObj.String()

	formStatus := "Draft"
	if isPublished {
		formStatus = "Published"
	}
	var documentID int64
	err := db.Get(&documentID, "SELECT document_id FROM document_ms WHERE document_uuid = $1", addDA.DocumentUUID)
	if err != nil {
		log.Println("Error getting document_id:", err)
		return err
	}

	err = db.Get(&documentCode, "SELECT document_code FROM document_ms WHERE document_uuid = $1", addDA.DocumentUUID)
	if err != nil {
		log.Println("Error getting document code:", err)
		return err
	}

	var projectID int64
	err = db.Get(&projectID, "SELECT project_id FROM project_ms WHERE project_uuid = $1", addDA.ProjectUUID)
	if err != nil {
		log.Println("Error getting project_id:", err)
		return err
	}
	// Generate form number based on document code
	formNumberDA, err := generateFormNumberDA(documentID, recrusionCount+1)
	if err != nil {
		// Handle error
		log.Println("Error generating form number:", err)
		return err
	}
	// Marshal ITCM struct to JSON
	daJSON, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshaling ITCM struct:", err)
		return err
	}

	_, err = db.NamedExec("INSERT INTO form_ms (form_id, form_uuid, document_id, user_id,form_number, form_ticket, form_status, form_data, project_id, created_by) VALUES (:form_id, :form_uuid, :document_id, :user_id,  :form_number, :form_ticket, :form_status, :form_data, :project_id, :created_by)", map[string]interface{}{
		"form_id":     app_id,
		"form_uuid":   uuidString,
		"document_id": documentID,
		"user_id":     userID,
		"form_number": formNumberDA,
		"form_ticket": addDA.FormTicket,
		"form_status": formStatus,
		"form_data":   daJSON,
		"project_id":  projectID,
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
			"form_id":    app_id,
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

func GetAllFormDA() ([]models.Formss, error) {
	rows, err := db.Query(`
		SELECT 
			f.form_uuid,  f.form_number, f.form_ticket, f.form_status, 
			d.document_name,
			p.project_name,
			f.reason, f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
			(f.form_data->>'nama_analis')::text AS nama_analis,
			(f.form_data->>'jabatan')::text AS jabatan,
			(f.form_data->>'departemen')::text AS departemen,
			(f.form_data->>'jenis_perubahan')::text AS jenis_perubahan,
			(f.form_data->>'detail_dampak_perubahan')::text AS detail_dampak_perubahan,
			(f.form_data->>'rencana_pengembangan_perubahan')::text AS rencana_pengembangan_perubahan,
			(f.form_data->>'rencana_pengujian_perubahan_sistem')::text AS rencana_pengujian_perubahan_sistem,
			(f.form_data->>'rencana_rilis_perubahan_dan_implementasi')::text AS rencana_rilis_perubahan_dan_implementasi,
			CASE
			WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
			WHEN f.is_approve = false THEN 'Tidak Disetujui'
			WHEN f.is_approve = true THEN 'Disetujui'
		END AS ApprovalStatus -- Alias the CASE expression as ApprovalStatus
			FROM 
			form_ms f
		LEFT JOIN 
			document_ms d ON f.document_id = d.document_id
		LEFT JOIN 
			project_ms p ON f.project_id = p.project_id
			WHERE
			d.document_code = 'DA' AND f.deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold all form data
	var forms []models.Formss

	// Iterate through the rows
	for rows.Next() {
		// Scan the row into the Forms struct
		var form models.Formss
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormTicket,
			&form.FormStatus,
			&form.DocumentName,
			&form.ProjectName,
			&form.Reason,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.NamaAnalis,
			&form.Jabatan,
			&form.Departemen,
			&form.JenisPerubahan,
			&form.DetailDampakPerubahan,
			&form.RencanaPengembanganPerubahan,
			&form.RencanaPengujianPerubahanSistem,
			&form.RencanaRilisPerubahanDanImplementasi,
			&form.ApprovalStatus,
		)
		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}
	// Return the forms as JSON response
	return forms, nil
}

func GetAllDAbyUserID(userID int) ([]models.Formss, error) {
	rows, err := db.Query(`
	SELECT 
		f.form_uuid, f.form_number, f.form_ticket, f.form_status, 
		d.document_name,
		p.project_name,
		f.reason, f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'nama_analis')::text AS nama_analis,
		(f.form_data->>'jabatan')::text AS jabatan,
		(f.form_data->>'departemen')::text AS departemen,
		(f.form_data->>'jenis_perubahan')::text AS jenis_perubahan,
		(f.form_data->>'detail_dampak_perubahan')::text AS detail_dampak_perubahan,
		(f.form_data->>'rencana_pengembangan_perubahan')::text AS rencana_pengembangan_perubahan,
		(f.form_data->>'rencana_pengujian_perubahan_sistem')::text AS rencana_pengujian_perubahan_sistem,
		(f.form_data->>'rencana_rilis_perubahan_dan_implementasi')::text AS rencana_rilis_perubahan_dan_implementasi,
		CASE
		WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
		WHEN f.is_approve = false THEN 'Tidak Disetujui'
		WHEN f.is_approve = true THEN 'Disetujui'
	END AS ApprovalStatus -- Alias the CASE expression as ApprovalStatus
		FROM 
		form_ms f
	LEFT JOIN 
		document_ms d ON f.document_id = d.document_id
	LEFT JOIN 
		project_ms p ON f.project_id = p.project_id
		WHERE
		f.user_id = $1 AND d.document_code = 'DA'  AND f.deleted_at IS NULL
`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold all form data
	var forms []models.Formss

	// Iterate through the rows
	for rows.Next() {
		// Scan the row into the Forms struct
		var form models.Formss
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormTicket,
			&form.FormStatus,
			&form.DocumentName,
			&form.ProjectName,
			&form.Reason,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.NamaAnalis,
			&form.Jabatan,
			&form.Departemen,
			&form.JenisPerubahan,
			&form.DetailDampakPerubahan,
			&form.RencanaPengembanganPerubahan,
			&form.RencanaPengujianPerubahanSistem,
			&form.RencanaRilisPerubahanDanImplementasi,
			&form.ApprovalStatus,
		)
		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}
	// Return the forms as JSON response
	return forms, nil
}

func GetAllDAbyAdmin() ([]models.Formss, error) {
	rows, err := db.Query(`
	SELECT 
		f.form_uuid, f.form_number, f.form_ticket, f.form_status, 
		d.document_name,
		p.project_name,
		f.reason, f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'nama_analis')::text AS nama_analis,
		(f.form_data->>'jabatan')::text AS jabatan,
		(f.form_data->>'departemen')::text AS departemen,
		(f.form_data->>'jenis_perubahan')::text AS jenis_perubahan,
		(f.form_data->>'detail_dampak_perubahan')::text AS detail_dampak_perubahan,
		(f.form_data->>'rencana_pengembangan_perubahan')::text AS rencana_pengembangan_perubahan,
		(f.form_data->>'rencana_pengujian_perubahan_sistem')::text AS rencana_pengujian_perubahan_sistem,
		(f.form_data->>'rencana_rilis_perubahan_dan_implementasi')::text AS rencana_rilis_perubahan_dan_implementasi,
		CASE
		WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
		WHEN f.is_approve = false THEN 'Tidak Disetujui'
		WHEN f.is_approve = true THEN 'Disetujui'
	END AS ApprovalStatus -- Alias the CASE expression as ApprovalStatus
		FROM 
		form_ms f
	LEFT JOIN 
		document_ms d ON f.document_id = d.document_id
	LEFT JOIN 
		project_ms p ON f.project_id = p.project_id
		WHERE
		d.document_code = 'DA'  AND f.deleted_at IS NULL
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold all form data
	var forms []models.Formss

	// Iterate through the rows
	for rows.Next() {
		// Scan the row into the Forms struct
		var form models.Formss
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormTicket,
			&form.FormStatus,
			&form.DocumentName,
			&form.ProjectName,
			&form.Reason,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.NamaAnalis,
			&form.Jabatan,
			&form.Departemen,
			&form.JenisPerubahan,
			&form.DetailDampakPerubahan,
			&form.RencanaPengembanganPerubahan,
			&form.RencanaPengujianPerubahanSistem,
			&form.RencanaRilisPerubahanDanImplementasi,
			&form.ApprovalStatus,
		)
		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}
	// Return the forms as JSON response
	return forms, nil
}

func GetSpecDA(id string) (models.Formss, error) {
	var specDA models.Formss

	err := db.Get(&specDA, `SELECT 
	f.form_uuid, f.form_number, f.form_ticket, f.form_status,
	d.document_name,
	p.project_name,
	CASE
		WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
		WHEN f.is_approve = false THEN 'Tidak Disetujui'
		WHEN f.is_approve = true THEN 'Disetujui'
	END AS ApprovalStatus,
	f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
	(f.form_data->>'nama_analis')::text AS nama_analis,
	(f.form_data->>'jabatan')::text AS jabatan,
	(f.form_data->>'departemen')::text AS departemen,
	(f.form_data->>'jenis_perubahan')::text AS jenis_perubahan,
	(f.form_data->>'detail_dampak_perubahan')::text AS detail_dampak_perubahan,
	(f.form_data->>'rencana_pengembangan_perubahan')::text AS rencana_pengembangan_perubahan,
	(f.form_data->>'rencana_pengujian_perubahan_sistem')::text AS rencana_pengujian_perubahan_sistem,
	(f.form_data->>'rencana_rilis_perubahan_dan_implementasi')::text AS rencana_rilis_perubahan_dan_implementasi
FROM 
	form_ms f
LEFT JOIN 
	document_ms d ON f.document_id = d.document_id
LEFT JOIN 
	project_ms p ON f.project_id = p.project_id
WHERE
	f.form_uuid = $1 and d.document_code = 'DA'  AND f.deleted_at IS NULL
`, id)

	if err != nil {
		return models.Formss{}, err
	}

	return specDA, nil
}

func GetSpecAllDA(id string) ([]models.FormsDAAll, error) {
	var signatories []models.FormsDAAll

	err := db.Select(&signatories, `SELECT 
	f.form_uuid, f.form_number, f.form_ticket, f.form_status,
	d.document_name,
	p.project_name,
	CASE
		WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
		WHEN f.is_approve = false THEN 'Tidak Disetujui'
		WHEN f.is_approve = true THEN 'Disetujui'
	END AS ApprovalStatus,
	f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
	(f.form_data->>'nama_analis')::text AS nama_analis,
	(f.form_data->>'jabatan')::text AS jabatan,
	(f.form_data->>'departemen')::text AS departemen,
	(f.form_data->>'jenis_perubahan')::text AS jenis_perubahan,
	(f.form_data->>'detail_dampak_perubahan')::text AS detail_dampak_perubahan,
	(f.form_data->>'rencana_pengembangan_perubahan')::text AS rencana_pengembangan_perubahan,
	(f.form_data->>'rencana_pengujian_perubahan_sistem')::text AS rencana_pengujian_perubahan_sistem,
	(f.form_data->>'rencana_rilis_perubahan_dan_implementasi')::text AS rencana_rilis_perubahan_dan_implementasi,
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
    f.form_uuid = $1 AND d.document_code = 'DA'  AND f.deleted_at IS NULL
	`, id)

	if err != nil {
		return nil, err
	}
	return signatories, nil
}

type FormDAWithSignatories struct {
	Form        models.Formss        `json:"form"`
	Signatories []models.SignatoryHA `json:"signatories"`
}

func GetSpecAllDAa(id string) (*FormDAWithSignatories, error) {
	var formWithSignatories FormDAWithSignatories

	// Ambil data form
	err := db.Get(&formWithSignatories.Form, `
        SELECT 
            f.form_uuid,
            f.form_number,
            f.form_ticket,
            f.form_status,
            d.document_name,
            p.project_name,
            CASE
                WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
                WHEN f.is_approve = false THEN 'Tidak Disetujui'
                WHEN f.is_approve = true THEN 'Disetujui'
            END AS ApprovalStatus,
            f.created_by,
            f.created_at,
            f.updated_by,
            f.updated_at,
            f.deleted_by,
            f.deleted_at,
            (f.form_data->>'nama_analis')::text AS nama_analis,
            (f.form_data->>'jabatan')::text AS jabatan,
            (f.form_data->>'departemen')::text AS departemen,
            (f.form_data->>'jenis_perubahan')::text AS jenis_perubahan,
            (f.form_data->>'detail_dampak_perubahan')::text AS detail_dampak_perubahan,
            (f.form_data->>'rencana_pengembangan_perubahan')::text AS rencana_pengembangan_perubahan,
            (f.form_data->>'rencana_pengujian_perubahan_sistem')::text AS rencana_pengujian_perubahan_sistem,
            (f.form_data->>'rencana_rilis_perubahan_dan_implementasi')::text AS rencana_rilis_perubahan_dan_implementasi
        FROM
            form_ms f
        LEFT JOIN 
            document_ms d ON f.document_id = d.document_id
        LEFT JOIN 
            project_ms p ON f.project_id = p.project_id
        WHERE
            f.form_uuid = $1 AND d.document_code = 'DA' AND f.deleted_at IS NULL
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

// func GetSpecAllDA(id string) ([]models.FormsDAAll, error) {
// 	var signatories []models.FormsDAAll

// 	err := db.Select(&signatories, `SELECT
// 	f.form_uuid, f.form_number, f.form_ticket, f.form_status,
// 	d.document_name,
// 	p.project_name,
// 	CASE
// 		WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
// 		WHEN f.is_approve = false THEN 'Tidak Disetujui'
// 		WHEN f.is_approve = true THEN 'Disetujui'
// 	END AS ApprovalStatus,
// 	f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
// 	(f.form_data->>'nama_analis')::text AS nama_analis,
// 	(f.form_data->>'jabatan')::text AS jabatan,
// 	(f.form_data->>'departemen')::text AS departemen,
// 	(f.form_data->>'jenis_perubahan')::text AS jenis_perubahan,
// 	(f.form_data->>'detail_dampak_perubahan')::text AS detail_dampak_perubahan,
// 	(f.form_data->>'rencana_pengembangan_perubahan')::text AS rencana_pengembangan_perubahan,
// 	(f.form_data->>'rencana_pengujian_perubahan_sistem')::text AS rencana_pengujian_perubahan_sistem,
// 	(f.form_data->>'rencana_rilis_perubahan_dan_implementasi')::text AS rencana_rilis_perubahan_dan_implementasi,
// 	sf.sign_uuid AS sign_uuid,
//     sf.name AS name,
//     sf.position AS position,
//     sf.role_sign AS role_sign,
// 	sf.is_sign AS is_sign
// FROM
//     form_ms f
// LEFT JOIN
//     document_ms d ON f.document_id = d.document_id
// LEFT JOIN
//     project_ms p ON f.project_id = p.project_id
// LEFT JOIN
//     sign_form sf ON f.form_id = sf.form_id
// WHERE
//     f.form_uuid = $1 AND d.document_code = 'DA'  AND f.deleted_at IS NULL
// 	`, id)

// 	if err != nil {
// 		return nil, err
// 	}
// 	return signatories, nil
//}

func UpdateFormDA(updateDA models.Form, data models.DampakAnalisa, username string, userID int, isPublished bool, id string) (models.Form, error) {
	currentTime := time.Now()
	formStatus := "Draft"
	if isPublished {
		formStatus = "Published"
	}

	var projectID int64
	err := db.Get(&projectID, "SELECT project_id FROM project_ms WHERE project_uuid = $1", updateDA.ProjectUUID)
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

	_, err = db.NamedExec("UPDATE form_ms SET user_id = :user_id,  form_ticket = :form_ticket, form_status = :form_status, form_data = :form_data, updated_by = :updated_by, updated_at = :updated_at WHERE form_uuid = :id AND form_status = 'Draft'", map[string]interface{}{
		"user_id":     userID,
		"form_ticket": updateDA.FormTicket,
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
	return updateDA, nil
}

func SignatureUser(userID int) ([]models.Formss, error) {
	rows, err := db.Query(`
	SELECT 
		f.form_uuid, f.form_number, f.form_ticket, f.form_status, 
		d.document_name,
		p.project_name,
		f.reason, f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
		(f.form_data->>'nama_analis')::text AS nama_analis,
		(f.form_data->>'jabatan')::text AS jabatan,
		(f.form_data->>'departemen')::text AS departemen,
		(f.form_data->>'jenis_perubahan')::text AS jenis_perubahan,
		(f.form_data->>'detail_dampak_perubahan')::text AS detail_dampak_perubahan,
		(f.form_data->>'rencana_pengembangan_perubahan')::text AS rencana_pengembangan_perubahan,
		(f.form_data->>'rencana_pengujian_perubahan_sistem')::text AS rencana_pengujian_perubahan_sistem,
		(f.form_data->>'rencana_rilis_perubahan_dan_implementasi')::text AS rencana_rilis_perubahan_dan_implementasi,
		CASE
		WHEN f.is_approve IS NULL THEN 'Menunggu Disetujui'
		WHEN f.is_approve = false THEN 'Tidak Disetujui'
		WHEN f.is_approve = true THEN 'Disetujui'
	END AS ApprovalStatus -- Alias the CASE expression as ApprovalStatus
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
	var forms []models.Formss

	// Iterate through the rows
	for rows.Next() {
		// Scan the row into the Forms struct
		var form models.Formss
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormTicket,
			&form.FormStatus,
			&form.DocumentName,
			&form.ProjectName,
			&form.Reason,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.NamaAnalis,
			&form.Jabatan,
			&form.Departemen,
			&form.JenisPerubahan,
			&form.DetailDampakPerubahan,
			&form.RencanaPengembanganPerubahan,
			&form.RencanaPengujianPerubahanSistem,
			&form.RencanaRilisPerubahanDanImplementasi,
			&form.ApprovalStatus,
		)
		if err != nil {
			return nil, err
		}

		forms = append(forms, form)
	}
	// Return the forms as JSON response
	return forms, nil
}

func GetDACode() (models.DocCodeName, error) {
	var documentCode models.DocCodeName

	err := db.Get(&documentCode, "SELECT document_uuid FROM document_ms WHERE document_code = 'DA'")

	if err != nil {
		return models.DocCodeName{}, err
	}
	return documentCode, nil
}
