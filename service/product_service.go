package service

import (
	"database/sql"
	"document/models"
	"log"
	"time"

	"github.com/google/uuid"
)

func AddProduct(addProduct models.Product, username string) error {
	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()

	app_id := currentTimestamp + int64(uniqueID)

	uuid := uuid.New()
	uuidString := uuid.String()
	_, err := db.NamedExec("INSERT INTO product_ms (product_id, product_uuid, product_name, product_owner, created_by) VALUES (:product_id, :product_uuid, :product_name, :product_owner, :created_by)", map[string]interface{}{
		"product_id":    app_id,
		"product_uuid":  uuidString,
		"product_name":  addProduct.ProductName,
		"product_owner": addProduct.ProductOwner,
		"created_by":    username,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetAllProduct() ([]models.Product, error) {
	products := []models.Product{}

	rows, errSelect := db.Queryx("select product_uuid, product_name, product_owner, product_order, created_by, created_at, updated_by, updated_at from product_ms WHERE deleted_at IS NULL")
	if errSelect != nil {
		return nil, errSelect
	}

	for rows.Next() {
		place := models.Product{}
		rows.StructScan(&place)
		products = append(products, place)
	}

	return products, nil
}

func ShowProductById(id string) (models.Product, error) {
	var product models.Product

	err := db.Get(&product, "SELECT product_uuid, product_name, product_owner, product_order, created_by, created_at, updated_by, updated_at FROM product_ms WHERE product_uuid = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return models.Product{}, err
	}
	return product, nil

}

func IsUniqueProduct(uuid, name string) (bool, error) {
	var count int

	var exsitingProName string
	err := db.Get(&exsitingProName, "SELECT product_name FROM product_ms WHERE product_uuid = $1 AND deleted_at IS NULL", uuid)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	if name == exsitingProName {
		return true, nil
	}

	err = db.Get(&count, "SELECT COUNT(*) FROM product_ms WHERE product_name = $1 AND product_uuid != $2 AND deleted_at IS NULL", name, uuid)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func GetProductIDByName(name string) (int, error) {
	var productID int
	err := db.QueryRow("SELECT product_id FROM product_ms WHERE product_name = $1 AND deleted_at IS NULL", name).Scan(&productID)
	return productID, err
}

func GetProductName(uuid string) (models.ProductName, error) {
	var docCodeName models.ProductName

	err := db.Get(&docCodeName, "SELECT product_name FROM product_ms WHERE product_uuid = $1 AND deleted_at IS NULL", uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			// Tidak ada baris yang sesuai
			log.Println("No rows found for product_uuid:", uuid)
			return models.ProductName{}, err
		}

		// Terjadi kesalahan lain
		log.Println("Error getting product data by product_ms:", err)
		return models.ProductName{}, err
	}

	return docCodeName, nil
}

func UpdateProduct(updateProduct models.Product, id string, username string) (models.Product, error) {
	currentTime := time.Now()

	_, err := db.NamedExec("UPDATE product_ms SET product_name = :product_name, product_owner = :product_owner, updated_by = :updated_by, updated_at = :updated_at WHERE product_uuid = :id", map[string]interface{}{
		"product_name":  updateProduct.ProductName,
		"product_owner": updateProduct.ProductOwner,
		"updated_by":    username,
		"updated_at":    currentTime,
		"id":            id,
	})
	if err != nil {
		log.Print(err)
		return models.Product{}, err
	}
	return updateProduct, nil
}

func DeleteProduct(id, username string) error {
	currentTime := time.Now()
	var productID int64
	err := db.Get(&productID, "SELECT product_id FROM product_ms WHERE product_uuid = $1", id)
	if err != nil {
		log.Println("Error getting product_id:", err)
		return err
	}

	// Melakukan soft delete pada tabel product_ms
	result, err := db.NamedExec("UPDATE product_ms SET deleted_by = :deleted_by, deleted_at = :deleted_at WHERE product_uuid = :id", map[string]interface{}{
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

	// Melakukan soft delete secara cascading pada tabel project_ms
	_, err = db.NamedExec("UPDATE project_ms SET deleted_by = :deleted_by, deleted_at = :deleted_at WHERE product_id = :id", map[string]interface{}{
		"deleted_by": username,
		"deleted_at": currentTime,
		"id":         productID,
	})
	if err != nil {
		log.Print(err)
		return err
	}

	// Melakukan soft delete secara cascading pada tabel form_ms yang terkait dengan project_ms
	_, err = db.NamedExec("UPDATE form_ms SET form_number = :form_number, deleted_by = :deleted_by, deleted_at = :deleted_at WHERE project_id IN (SELECT project_id FROM project_ms WHERE product_id = :id)", map[string]interface{}{
		"form_number": " ",
		"deleted_by":  username,
		"deleted_at":  currentTime,
		"id":          productID,
	})
	if err != nil {
		log.Print(err)
		return err
	}

	_, err = db.NamedExec("UPDATE sign_form SET deleted_by = :deleted_by, deleted_at = :deleted_at WHERE form_id IN (SELECT form_id FROM form_ms WHERE project_id IN (SELECT project_id FROM project_ms WHERE product_id = :id))", map[string]interface{}{
		"deleted_by": username,
		"deleted_at": currentTime,
		"id":         productID,
	})
	if err != nil {
		log.Print(err)
		return err
	}
	return nil
}
