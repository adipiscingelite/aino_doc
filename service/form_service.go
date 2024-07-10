package service

import (
	"database/sql"
	"document/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

func GetUserIdFromToken(tokenStr string) (int, string, error) {
	var claims JwtCustomClaims
	if err := json.Unmarshal([]byte(tokenStr), &claims); err != nil {
		return 0, "", fmt.Errorf("Gagal mengurai klaim: %v", err)
	}

	userID := claims.UserID
	divisionTitle := claims.DivisionTitle

	log.Println("USER ID : ", userID)
	log.Println("DIVISION TITLE : ", divisionTitle)
	return userID, divisionTitle, nil
}

func GetDivisionCode(tokenStr string) (string, error) {
	var claims JwtCustomClaims
	if err := json.Unmarshal([]byte(tokenStr), &claims); err != nil {
		return "", fmt.Errorf("Gagal mengurai klaim: %v", err)
	}
	divisionCode := claims.DivisionCode

	log.Println("Division Code: ", divisionCode)
	return divisionCode, nil
}

func GetDocumentCode(documentID int64) (string, error) {
	var documentCode string
	err := db.Get(&documentCode, "SELECT document_code FROM document_ms WHERE document_id = $1", documentID)

	if err != nil {
		log.Println("Error getting document code:", err)
	}
	return documentCode, nil
}

func GetProjectCode(projectID int64) (string, error) {
	var projectCode string
	err := db.Get(&projectCode, "SELECT project_code FROM project_ms WHERE project_uuid = $1", projectID)
	if err != nil {
		return "", fmt.Errorf("Failed to get project ID: %v", err)
	}
	return projectCode, nil
}
func convertToRoman(num int) (string, error) {
	if num < 1 || num > 12 {
		return "", errors.New("Month out of range")
	}

	// Define Roman numeral representations for each digit
	romans := []struct {
		value   int
		numeral string
	}{
		{10, "X"},
		{9, "IX"},
		{5, "V"},
		{4, "IV"},
		{1, "I"},
	}

	var result strings.Builder

	for _, r := range romans {
		for num >= r.value {
			result.WriteString(r.numeral)
			num -= r.value
		}
	}

	return result.String(), nil
}
func generateFormNumber(documentID int64, divisionCode string, recursionCount int) (string, error) {
	const maxRecursionCount = 1000

	// Check if the maximum recursion count is reached
	if recursionCount > maxRecursionCount {
		return "", errors.New("Maximum recursion count exceeded")
	}

	// Get document code
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

	// Format the form number according to the specified format
	formNumberString := fmt.Sprintf("%04d", formNumber)
	formNumberWithDivision := fmt.Sprintf("%s/%s/%s/%s/%d", formNumberString, divisionCode, documentCode, romanMonth, year)

	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM form_ms WHERE form_number = $1", formNumberString)
	if err != nil {
		return "", fmt.Errorf("Error checking existing form number: %v", err)
	}
	if count > 0 {
		// If the form number already exists, recursively call the function again
		return generateFormNumber(documentID, divisionCode, recursionCount+1)
	}

	return formNumberWithDivision, nil
}

func generateProjectFormNumber(documentID int64, projectID int64, recursionCount int) (string, error) {
	const maxRecursionCount = 1000

	// Check if the maximum recursion count is reached
	if recursionCount > maxRecursionCount {
		return "", errors.New("Maximum recursion count exceeded")
	}

	documentCode, err := GetDocumentCode(documentID)
	if err != nil {
		return "", fmt.Errorf("failed to get document code: %v", err)
	}

	// Get document code
	projectCode, err := GetProjectCode(projectID)
	if err != nil {
		return "", fmt.Errorf("failed to get document code: %v", err)
	}

	// Get the latest form number for the given document ID
	var latestFormNumber sql.NullString
	err = db.Get(&latestFormNumber, "SELECT MAX(form_number) FROM form_ms WHERE document_id = $1", documentID)
	if err != nil {
		return "", fmt.Errorf("error getting latest form number: %v", err)
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
	formNumberWithDivision := fmt.Sprintf("%s/%s/%s/%s/%d", formNumberString, projectCode, documentCode, romanMonth, year)

	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM form_ms WHERE form_number = $1", formNumberString)
	if err != nil {
		return "", fmt.Errorf("Error checking existing form number: %v", err)
	}
	if count > 0 {
		// If the form number already exists, recursively call the function again
		return generateFormNumber(documentID, projectCode, recursionCount+1)
	}

	return formNumberWithDivision, nil
}

func AddForm(addFrom models.Form, isPublished bool, username string, userID int, divisionCode string, recursionCount int) error {

	var documentCode string
	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()

	app_id := currentTimestamp + int64(uniqueID)

	uuid := uuid.New()
	uuidString := uuid.String()

	formStatus := "Draft"
	if isPublished {
		formStatus = "Published"
	}
	var documentID int64
	err := db.Get(&documentID, "SELECT document_id FROM document_ms WHERE document_uuid = $1", addFrom.DocumentUUID)
	if err != nil {
		log.Println("Error getting document_id:", err)
		return err
	}

	err = db.Get(&documentCode, "SELECT document_code FROM document_ms WHERE document_uuid = $1", addFrom.DocumentUUID)
	if err != nil {
		log.Println("Error getting document code:", err)
		return err
	}

	// Generate form number based on document code
	formNumber, err := generateFormNumber(documentID, divisionCode, recursionCount+1)
	if err != nil {
		// Handle error
		log.Println("Error generating form number:", err)
		return err
	}

	_, err = db.NamedExec("INSERT INTO form_ms (form_id, form_uuid, document_id, user_id,form_number, form_ticket, form_status, created_by) VALUES (:form_id, :form_uuid, :document_id, :user_id,:form_number, :form_ticket, :form_status, :created_by)", map[string]interface{}{
		"form_id":     app_id,
		"form_uuid":   uuidString,
		"document_id": documentID,
		"user_id":     userID,
		"form_number": formNumber,
		"form_ticket": addFrom.FormTicket,
		"form_status": formStatus,
		"created_by":  username,
	})

	if err != nil {
		return err
	}

	// _, err = db.Exec("SET LOCAL my_vars.division_code TO $1", divisionCode)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func GetAllForm() ([]models.Forms, error) {

	form := []models.Forms{}

	rows, errSelect := db.Queryx("SELECT f.form_uuid, f.form_number, f.form_ticket, f.form_status, f.created_by, f.created_at, f.updated_by, f.updated_at, d.document_name FROM form_ms f JOIN  document_ms d ON f.document_id = d.document_id WHERE f.deleted_at IS NULL")
	//rows, errSelect := db.Queryx("select form_uuid, form_number, form_ticket, form_status, document_id, user_id, created_by, created_at, updated_by, updated_at from form_ms WHERE deleted_at IS NULL")
	if errSelect != nil {
		return nil, errSelect
	}

	for rows.Next() {
		place := models.Forms{}
		rows.StructScan(&place)
		form = append(form, place)
	}

	return form, nil
}

func MyForm(userID int) ([]models.Forms, error) {
	var form []models.Forms

	errSelect := db.Select(&form, "SELECT f.form_uuid, f.form_number, f.form_ticket, f.form_status, f.created_by, f.created_at, f.updated_by, f.updated_at, d.document_name FROM form_ms f JOIN  document_ms d ON f.document_id = d.document_id WHERE f.user_id = $1 AND f.deleted_at IS NULL", userID)
	//rows, errSelect := db.Queryx("select form_uuid, form_number, form_ticket, form_status, document_id, user_id, created_by, created_at, updated_by, updated_at from form_ms WHERE deleted_at IS NULL")
	if errSelect != nil {
		log.Print(errSelect)
		return nil, errSelect
	}

	if len(form) == 0 {
		return nil, sql.ErrNoRows
	}

	return form, nil
}

func FormByDivision(divisionCode string) ([]models.Forms, error) {
	var form []models.Forms

	errSelect := db.Select(&form, `
    SELECT f.form_uuid, f.form_number, f.form_ticket, f.form_status, f.created_by, f.created_at, f.updated_by, f.updated_at, d.document_name 
    FROM form_ms f 
    JOIN document_ms d ON f.document_id = d.document_id 
    WHERE f.deleted_at IS NULL AND SPLIT_PART(f.form_number, '/', 2) = $1
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

func ShowFormById(id string) (models.Forms, error) {
	var form models.Forms

	//err := db.Get(&form, "SELECT f.form_uuid, f.form_number, f.form_ticket, f.form_status, f.user_id, f.created_by, f.created_at, f.updated_by, f.updated_at, d.document_name FROM form_ms f JOIN  document_ms d ON f.document_id = d.document_id WHERE f.form_uuid = $1 AND f.deleted_at IS NULL", id)
	err := db.Get(&form, "SELECT f.form_uuid, f.form_number, f.form_ticket, f.form_status, f.created_by, f.created_at, f.updated_by, f.updated_at, d.document_name FROM form_ms f JOIN  document_ms d ON f.document_id = d.document_id WHERE f.form_uuid = $1 AND f.deleted_at IS NULL", id)
	if err != nil {
		return models.Forms{}, err
	}
	return form, nil

}

func GetPreviousDocumentID(formUUID string) (int64, error) {
	var previousDocumentID int64
	err := db.Get(&previousDocumentID, "SELECT document_id FROM form_ms WHERE form_uuid = $1", formUUID)
	if err != nil {
		return 0, err
	}
	return previousDocumentID, nil
}

func GetFormNumber(formUUID string) (string, error) {
	var formNumber string
	err := db.Get(&formNumber, "SELECT form_number FROM form_ms WHERE form_uuid = $1", formUUID)
	if err != nil {
		return "", err
	}
	return formNumber, nil
}

func UpdateForm(updateForm models.Form, id string, isPublished bool, username string, userID int, divisionCode string, recursionCount int) (models.Form, error) {
	currentTime := time.Now()
	formStatus := "Draft"
	if isPublished {
		formStatus = "Published"
	}

	var documentID int64
	err := db.Get(&documentID, "SELECT document_id FROM document_ms WHERE document_uuid = $1", updateForm.DocumentUUID)
	if err != nil {
		log.Println("Error getting document_id:", err)
		return models.Form{}, err
	}

	var documentCode string
	err = db.Get(&documentCode, "SELECT document_code FROM document_ms WHERE document_uuid = $1", updateForm.DocumentUUID)
	if err != nil {
		log.Println("Error getting document code:", err)
		return models.Form{}, err
	}

	previousDocumentID, err := GetPreviousDocumentID(id)
	if err != nil {
		log.Println("Error getting previous document ID:", err)
		return models.Form{}, err
	}

	var formNumber string
	if documentID != previousDocumentID {
		// Jika documentID berbeda dengan previousDocumentID, maka kita perlu menghasilkan form_number baru
		formNumber, err = generateFormNumber(documentID, divisionCode, recursionCount+1)
		if err != nil {
			log.Println("Error generating form number:", err)
			return models.Form{}, err
		}
	} else {
		// Jika documentID sama dengan previousDocumentID, kita perlu mengambil form_number dari form sebelumnya
		formNumber, err = GetFormNumber(id)
		if err != nil {
			log.Println("Error getting form number:", err)
			return models.Form{}, err
		}
	}

	var existingUserID int
	errID := db.Get(&existingUserID, "SELECT user_id FROM form_ms WHERE form_uuid = $1", id)
	if errID != nil {
		log.Println("Error getting user ID:", err)
		return models.Form{}, err
	}
	if existingUserID != userID {
		return models.Form{}, errors.New("You are not authorized to update this form")
	}
	log.Println("EXISTING USER ID : ", existingUserID)

	_, err = db.NamedExec("UPDATE form_ms SET  form_number = :form_number, form_ticket = :form_ticket, form_status = :form_status, document_id = :document_id, user_id = :user_id, updated_by = :updated_by, updated_at = :updated_at WHERE form_uuid = :id and form_status='Draft'", map[string]interface{}{
		"form_number": formNumber,
		"form_ticket": updateForm.FormTicket,
		"form_status": formStatus,
		"document_id": documentID,
		"user_id":     userID,
		"updated_by":  username,
		"updated_at":  currentTime,
		"id":          id,
	})
	if err != nil {
		log.Print(err)
		return models.Form{}, err
	}
	return updateForm, nil
}

// delete formulir
var ErrNotFound = errors.New("form not found")

func DeleteForm(id string, username string) error {
	currentTime := time.Now()

	result, err := db.Exec("UPDATE form_ms SET deleted_by = $1, deleted_at = $2, form_number = '' WHERE form_uuid = $3", username, currentTime, id)
	if err != nil {
		return err
	}

	// Soft delete related sign_form entries
	deleteSignFormQuery := `
		UPDATE sign_form
		SET deleted_by = $1, deleted_at = NOW()
		WHERE form_id = (
			SELECT form_id
			FROM form_ms
			WHERE form_uuid = $2
		)
	`
	_, err = db.Exec(deleteSignFormQuery, username, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil

}
