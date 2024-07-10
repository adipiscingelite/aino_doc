package service

import (
	"database/sql"
	"document/models"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func GetAllPersonalName() ([]models.Personal, error) {
	getUserAppRole := []models.Personal{}

	// Lakukan query ke database lain
	rows, err := db2.Queryx("SELECT u.user_id, pd.personal_name FROM user_ms u JOIN personal_data_ms pd ON u.user_id = pd.user_id WHERE u.deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		place := models.Personal{}
		err := rows.StructScan(&place)
		if err != nil {
			log.Println("Error scanning row to struct:", err)
			continue
		}
		getUserAppRole = append(getUserAppRole, place)
	}

	return getUserAppRole, nil
}

func GetSignatureForm(id string) ([]models.Signatories, error) {
	var signatories []models.Signatories

	err := db.Select(&signatories, `SELECT 
	sf.sign_uuid, 
	sf.name, 
	sf.position, 
	sf.role_sign, 
	sf.is_sign, 
	sf.created_by, 
	sf.created_at, 
	sf.updated_by, 
	sf.updated_at, 
	sf.deleted_by, 
	sf.deleted_at
FROM 
	sign_form sf 
	JOIN form_ms fm ON sf.form_id = fm.form_id 
WHERE 
	fm.form_uuid = $1 AND sf.deleted_at IS NULL`, id)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return signatories, nil

}

func GetSpecSignatureByID(id string) (models.Signatorie, error) {
	var signatories models.Signatorie
	err := db.Get(&signatories, "SELECT sign_uuid, name, position, role_sign, is_sign, created_by, created_at, updated_by, updated_at, deleted_by, deleted_at FROM sign_form sf WHERE sign_uuid = $1 AND deleted_at IS NULL", id)
	if err != nil {
		log.Print(err)
		return models.Signatorie{}, err
	}

	return signatories, nil
}

func GetUserIDSign(id string) (models.UserIDSign, error) {
	var userID models.UserIDSign
	err := db.Get(&userID, "SELECT user_id, sign_uuid FROM sign_form WHERE sign_uuid = $1", id)
	if err != nil {
		log.Print(err)
		return models.UserIDSign{}, err
	}
	return userID, nil
}

func UpdateFormSignature(updateSign models.UpdateSign, id string, username string) error {
	currentTime := time.Now()

	_, err := db.NamedExec("UPDATE sign_form SET is_sign = :is_sign, updated_by = :updated_by, updated_at = :updated_at WHERE sign_uuid = :id", map[string]interface{}{
		"is_sign":    updateSign.IsSign,
		"updated_by": username,
		"updated_at": currentTime,
		"id":         id,
	})
	if err != nil {
		return err
	}
	return nil
}

// ga kepake
func GetUserRoleFromDatabase(userID int) (string, error) {
	var role string
	err := db.QueryRow("SELECT role_sign FROM sign_form WHERE user_id = $1", userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("Pengguna tidak ditemukan")
		}
		return "", err
	}
	return role, nil
}

func AddApproval(addApproval models.AddApproval, id string, username string, userID int) error {
	// //userRole, err := GetUserRoleFromDatabase(userID)
	// if err != nil {
	// 	return err
	// }

	// Periksa apakah peran pengguna adalah "Atasan Penerima"
	currentTime := time.Now()
	_, err := db.NamedExec("UPDATE form_ms SET is_approve = :is_approve, reason = :reason, updated_by = :updated_by, updated_at = :updated_at WHERE form_uuid = :id", map[string]interface{}{
		"is_approve": addApproval.IsApproval,
		"reason":     addApproval.Reason,
		"updated_by": username,
		"updated_at": currentTime,
		"id":         id,
	})
	if err != nil {
		return err
	}
	return nil

}

func UpdateSignInfo(signatory models.UpdateSignForm, id string, username string) (models.UpdateSignForm, error) {
	currentTime := time.Now()
	personalName := signatory.Name

	personalNames, err := GetAllPersonalName()
	if err != nil {
		log.Println("Error getting personal names:", err)
		return models.UpdateSignForm{}, err
	}
	var userID int
	for _, personal := range personalNames {
		if personal.PersonalName == personalName {
			userID, err = strconv.Atoi(personal.UserID)
			if err != nil {
				return models.UpdateSignForm{}, err
			}
			break
		}
	}

	if userID == 0 {
		log.Printf("User ID not found for personal name: %s\n", personalName)
		return models.UpdateSignForm{}, errors.New("User ID not found for personal name")
	}

	_, err = db.NamedExec("UPDATE sign_form SET user_id = :user_id, name = :name, position = :position, role_sign = :role_sign, updated_by = :updated_by, updated_at = :updated_at WHERE sign_uuid = :sign_uuid", map[string]interface{}{
		"user_id":    userID,
		"name":       personalName,
		"position":   signatory.Position,
		"role_sign":  signatory.Role,
		"updated_by": username,
		"updated_at": currentTime,
		"sign_uuid":  id,
	})
	if err != nil {
		return models.UpdateSignForm{}, err
	}

	return signatory, nil
}

func AddSignInfo(signatory models.AddSignInfo, username string) error {
	currentTime := time.Now()
	uuidObj := uuid.New()
	uuidString := uuidObj.String()

	var formID int
	err := db.Get(&formID, "SELECT form_id FROM form_ms WHERE form_uuid = $1", signatory.FormUUID)
	if err != nil {
		log.Println("Error getting form_id:", err)
		return err
	}
	personalName := signatory.Name
	personalNames, err := GetAllPersonalName()
	if err != nil {
		log.Println("Error getting personal names:", err)
		return err
	}
	var userID int
	for _, personal := range personalNames {
		if personal.PersonalName == personalName {
			userID, err = strconv.Atoi(personal.UserID)
			if err != nil {
				return err
			}
			break
		}
	}

	if userID == 0 {
		log.Printf("User ID not found for personal name: %s\n", personalName)
		return errors.New("User ID not found for personal name")
	}
	_, err = db.NamedExec("INSERT INTO sign_form (sign_uuid, form_id, user_id, name, position, role_sign, created_by) VALUES (:sign_uuid, :form_id, :user_id, :name, :position, :role_sign, :created_by)", map[string]interface{}{
		"sign_uuid":  uuidString,
		"user_id":    userID,
		"form_id":    formID,
		"name":       signatory.Name,
		"position":   signatory.Position,
		"role_sign":  signatory.Role,
		"created_by": username,
		"created_at": currentTime,
	})
	if err != nil {
		return err
	}

	return nil
}

func DeleteSignInfo(id, username string) error {
	currentTime := time.Now()
	result, err := db.NamedExec("UPDATE sign_form SET deleted_by = :deleted_by, deleted_at = :deleted_at WHERE sign_uuid = :id", map[string]interface{}{
		"deleted_by": username,
		"deleted_at": currentTime,
		"id":         id,
	})

	if err != nil {
		log.Print(err)
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound // Mengembalikan error jika tidak ada rekaman yang cocok
	}

	return nil
}
