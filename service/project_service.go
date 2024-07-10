package service

import (
	"database/sql"
	"document/models"
	"log"
	"time"

	"github.com/google/uuid"
)

func AddProject(addForm models.Project, username string) error {
	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()

	app_id := currentTimestamp + int64(uniqueID)

	uuid := uuid.New()
	uuidString := uuid.String()

	var productID int64
	err := db.Get(&productID, "SELECT product_id FROM product_ms WHERE product_uuid = $1", addForm.ProductUUID)
	if err != nil {
		log.Println("Error getting product_id:", err)
		return err
	}
	_, err = db.NamedExec("INSERT INTO project_ms (project_id, product_id, project_uuid, project_name, project_code, project_manager, created_by) VALUES (:project_id, :product_id, :project_uuid, :project_name, :project_code, :project_manager, :created_by)", map[string]interface{}{
		"project_id":      app_id,
		"project_uuid":    uuidString,
		"product_id":      productID,
		"project_name":    addForm.ProjectName,
		"project_code":    addForm.ProjectCode,
		"project_manager": addForm.ProjectManager,
		"created_by":      username,
	})

	if err != nil {
		return err
	}

	return nil
}

func GetAllProject() ([]models.Projects, error) {
	projects := []models.Projects{}
	err := db.Select(&projects, "SELECT p.project_uuid, f.product_name, p.project_order, p.project_name, p.project_code, p.project_manager, p.created_by, p.created_at, p.updated_by, p.updated_at, p.deleted_by, p.deleted_at FROM project_ms p JOIN product_ms f ON p.product_id = f.product_id WHERE p.deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func ShowProjectById(id string) (models.Projects, error) {
	var project models.Projects
	err := db.Get(&project, "SELECT p.project_uuid, f.product_name, p.project_order, p.project_name, p.project_code, p.project_manager, p.created_by, p.created_at, p.updated_by, p.updated_at, p.deleted_by, p.deleted_at FROM project_ms p JOIN product_ms f ON p.product_id = f.product_id WHERE p.project_uuid = $1 AND p.deleted_at IS NULL", id)
	if err != nil {
		return models.Projects{}, err
	}
	return project, nil
}

func UpdateProject(updateForm models.Project, id string, username string) (models.Project, error) {
	var productID int64
	err := db.Get(&productID, "SELECT product_id FROM product_ms WHERE product_uuid = $1", updateForm.ProductUUID)
	if err != nil {
		log.Println("Error getting product_id:", err)
		return models.Project{}, err
	}
	currentTimestamp := time.Now()
	_, err = db.NamedExec("UPDATE project_ms SET product_id = :product_id, project_name = :project_name, project_code = :project_code, project_manager = :project_manager, updated_by = :updated_by, updated_at = :updated_at WHERE project_uuid = :id", map[string]interface{}{
		"product_id":      productID,
		"project_name":    updateForm.ProjectName,
		"project_code":    updateForm.ProjectCode,
		"project_manager": updateForm.ProjectManager,
		"updated_by":      username,
		"updated_at":      currentTimestamp,
		"id":              id,
	})
	if err != nil {
		log.Print(err)
		return models.Project{}, err
	}

	return updateForm, nil

}

func GetProjectCodeName(uuid string) (models.ProjectCodeName, error) {
	var projectCodeName models.ProjectCodeName
	err := db.Get(&projectCodeName, "SELECT project_uuid, project_code, project_name FROM project_ms WHERE project_uuid = $1 AND deleted_at IS NULL", uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No rows found for project_uuid:", uuid)
			return models.ProjectCodeName{}, err
		}
		log.Println("Error getting project data by project_ms:", err)
		return models.ProjectCodeName{}, err
	}
	return projectCodeName, nil
}

func IsUniqueProject(uuid, code, name string) (bool, error) {
	var count int
	var exsitingProCode, exsitingProName string
	err := db.Get(&exsitingProCode, "SELECT project_code FROM project_ms WHERE project_uuid = $1 AND deleted_at IS NULL", uuid)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	err = db.Get(&exsitingProName, "SELECT project_name FROM project_ms WHERE project_uuid = $1 AND deleted_at IS NULL", uuid)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if code == exsitingProCode && name == exsitingProName {
		return true, nil
	}
	err = db.Get(&count, "SELECT COUNT(*) FROM project_ms WHERE (project_code = $1 OR project_name = $2) AND project_uuid != $3 AND deleted_at IS NULL", code, name, uuid)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func GetProjectIDByName(name string) (int, error) {
	var projectID int
	err := db.QueryRow("SELECT project_id FROM project_ms WHERE project_name = $1 AND deleted_at IS NULL", name).Scan(&projectID)
	return projectID, err
}

func GetProjectIDByCode(code string) (int, error) {
	var projectID int
	err := db.QueryRow("SELECT project_id FROM project_ms WHERE project_code = $1 AND deleted_at IS NULL", code).Scan(&projectID)
	return projectID, err
}

func DeleteProject(id, username string) error {
	currentTime := time.Now()
	var projectID int64
	err := db.Get(&projectID, "SELECT project_id FROM project_ms WHERE project_uuid = $1", id)
	if err != nil {
		log.Println("Error getting project_id:", err)
		return err
	}
	_, err = db.NamedExec("UPDATE project_ms SET deleted_by = :deleted_by, deleted_at = :deleted_at WHERE project_uuid = :id", map[string]interface{}{
		"deleted_by": username,
		"deleted_at": currentTime,
		"id":         id,
	})
	if err != nil {
		log.Print(err)
		return err
	}

	// Melakukan soft delete secara cascading pada tabel form_ms yang terkait dengan project_ms
	_, err = db.NamedExec("UPDATE form_ms SET form_number = :form_number, deleted_by = :deleted_by, deleted_at = :deleted_at WHERE project_id = :id", map[string]interface{}{
		"form_number": " ",
		"deleted_by":  username,
		"deleted_at":  currentTime,
		"id":          projectID,
	})
	if err != nil {
		log.Print(err)
		return err
	}

	_, err = db.NamedExec("UPDATE sign_form SET deleted_by = :deleted_by, deleted_at = :deleted_at WHERE form_id IN (SELECT form_id FROM form_ms WHERE project_id = :id)", map[string]interface{}{
		"deleted_by": username,
		"deleted_at": currentTime,
		"id":         projectID,
	})
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}
