package service

import (
	"database/sql"
	"document/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
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

func AddBeritaAcara(addForm models.Form, beritaAcara models.BeritaAcara, isPublished bool, userID int, username string, divisionCode string, recursionCount int, signatories []models.Signatory) error {
	currentTime := time.Now()
	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()
	appID := currentTimestamp + int64(uniqueID)
	uuidObj := uuid.New()
	uuidString := uuidObj.String()

	picUUID := uuid.New().String() // Membuat UUID baru untuk PIC

	if beritaAcara.Jenis == "Peminjaman" {
		beritaAcara.PicUUID = picUUID // Menyimpan ke struktur berita acara
	} else {
		beritaAcara.PicUUID = ""
		fmt.Println("Bukan peminjaman")
	}

	// fmt.Println("pic uuid form", beritaAcara.PicUUID)
	// fmt.Println("pic uuid pic", picUUID)

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
	baJSON, err := json.Marshal(beritaAcara)
	if err != nil {
		log.Println("Error marshaling ITCM struct:", err)
		return err
	}

	_, err = db.NamedExec("INSERT INTO form_ms (form_id, form_uuid, form_ticket, document_id, user_id, form_number, form_status, form_data, created_by) VALUES (:form_id, :form_uuid, :form_ticket, :document_id, :user_id, :form_number, :form_status, :form_data, :created_by)", map[string]interface{}{
		"form_id":     appID,
		"form_uuid":   uuidString,
		"form_ticket": "",
		"document_id": documentID,
		"user_id":     userID,
		"form_number": formNumber,
		"form_status": formStatus,
		"form_data":   baJSON, // Convert JSON to string
		"created_by":  username,
	})

	if err != nil {
		return err
	}

	fmt.Println("peliss", beritaAcara.AssetUUID)
	var assetID int64
	err = db.Get(&assetID, "SELECT asset_id FROM assets_ms WHERE asset_uuid = $1", beritaAcara.AssetUUID)
	if err != nil {
		log.Println("Error getting asset_uuid:", err)
		return err
	}

	var assetStatus string
	if beritaAcara.Jenis == "Peminjaman" {
		assetStatus = "Dipinjam"
	} else if beritaAcara.Jenis == "Pengembalian" {
		assetStatus = "Tersedia"
	}
	// Update status asset di database
	_, err = db.NamedExec(`UPDATE assets_ms 
                       SET asset_status = :asset_status, updated_by = :updated_by, updated_at = :updated_at 
                       WHERE asset_id = :asset_id`, map[string]interface{}{
		"asset_status": assetStatus,
		"updated_by":   username,
		"updated_at":   currentTime,
		"asset_id":     assetID,
	})

	if err != nil {
		return err
	}

	if beritaAcara.Jenis == "Peminjaman" {
		var latestPicNumber sql.NullString
		err = db.Get(&latestPicNumber, "SELECT MAX(CAST(REGEXP_REPLACE(pic_description, '[^0-9]', '', 'g') AS INTEGER)) FROM pic_ms WHERE asset_id = $1", assetID)
		if err != nil {
			return fmt.Errorf("Error getting latest form number: %v", err)
		}

		// Initialize picNumber to 1 if latestPicNumber is NULL
		picNumber := 1
		if latestPicNumber.Valid {
			re := regexp.MustCompile(`\d+`) // Regular expression to capture numbers
			match := re.FindString(latestPicNumber.String)

			if match != "" {
				var latestPicNumberInt int
				_, err := fmt.Sscanf(match, "%d", &latestPicNumberInt) // Parsing number
				if err != nil {
					return fmt.Errorf("Error parsing latest pic number: %v", err)
				}
				// Increment the latest pic number
				picNumber = latestPicNumberInt + 1
			} else {
				log.Println("No number found in latest pic description, starting with 1")
			}
		}

		fmt.Println("Latest PIC Number from DB:", latestPicNumber.String)
		fmt.Println("New PIC Number:", picNumber)

		// Ensure PIC description is formatted correctly
		PICWithNumber := fmt.Sprintf("PIC %d", picNumber)
		fmt.Println("New PIC Description for UUID:", PICWithNumber)

		_, err = db.NamedExec("INSERT INTO pic_ms (pic_uuid, asset_id, pic_name, pic_description, created_by) VALUES (:pic_uuid, :asset_id, :pic_name, :pic_description, :created_by)", map[string]interface{}{
			"pic_uuid":        picUUID,
			"asset_id":        assetID,
			"pic_name":        beritaAcara.NamaPIC,
			"pic_description": PICWithNumber,
			"created_by":      username,
		})
		if err != nil {
			return err
		}
	} else {
		log.Println("Jenis Berita Acara bukan Peminjaman, skip proses insert PIC")
	}

	// personalNames, err := GetAllPersonalName() // Mengambil daftar semua personal name
	// if err != nil {
	// 	log.Println("Error getting personal names:", err)
	// 	return err
	// }

	// for _, signatory := range signatories {
	// 	uuidString := uuid.New().String()

	// 	// Mencari user_id yang sesuai dengan personal_name yang dipilih
	// 	var userID string
	// 	for _, personal := range personalNames {
	// 		if personal.PersonalName == signatory.Name {
	// 			userID = personal.UserID
	// 			break
	// 		}
	// 	}

	// 	// Memastikan user_id ditemukan untuk personal_name yang dipilih
	// 	if userID == "" {
	// 		log.Printf("User ID not found for personal name: %s\n", signatory.Name)
	// 		continue
	// 	}

	// 	_, err := db.NamedExec("INSERT INTO sign_form (sign_uuid, form_id, user_id, name, position, role_sign, created_by) VALUES (:sign_uuid, :form_id, :user_id, :name, :position, :role_sign, :created_by)", map[string]interface{}{
	// 		"sign_uuid":  uuidString,
	// 		"user_id":    userID,
	// 		"form_id":    appID,
	// 		"name":       signatory.Name,
	// 		"position":   signatory.Position,
	// 		"role_sign":  signatory.Role,
	// 		"created_by": username,
	// 	})
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

func AddAsset(addAsset models.Asset, pic []models.Pic, userID int, username string, divisionCode string, recursionCount int) error {
	const maxRecursionCount = 1000

	// Check if the maximum recursion count is reached
	if recursionCount > maxRecursionCount {
		return errors.New("Maximum recursion count exceeded")

	}
	currentTimestamp := time.Now().UnixNano() / int64(time.Microsecond)
	uniqueID := uuid.New().ID()
	appID := currentTimestamp + int64(uniqueID)
	uuidObj := uuid.New()
	uuidString := uuidObj.String()

	var err error

	var latestjawa sql.NullString
	err = db.Get(&latestjawa, "SELECT MAX(asset_code) FROM assets_ms")
	if err != nil {
		return fmt.Errorf("Error getting latest form number: %v", err)
	}

	formNumber := 1
	if latestjawa.Valid {
		// Gunakan regex untuk mengekstrak angka terakhir setelah "LP"
		re := regexp.MustCompile(`LP(\d+)$`)
		matches := re.FindStringSubmatch(latestjawa.String)

		if len(matches) == 2 {
			latestFormNumberInt, err := strconv.Atoi(matches[1]) // Konversi dari string ke int
			if err != nil {
				return fmt.Errorf("Error parsing latest form number: %v", err)
			}
			// Increment the latest form number
			formNumber = latestFormNumberInt + 1
		} else {
			return fmt.Errorf("Error: No valid form number found in asset_code")
		}
	}
	year := time.Now().Year()
	month := int(time.Now().Month())                  // Mengubah bulan menjadi angka
	formNumberString := fmt.Sprintf("%d", formNumber) // Menghilangkan leading zero

	// Buat assetCode dengan benar
	assetCode := fmt.Sprintf("%s/%d/%02d/%s/LP%s", "HD", year, month, "YOG", formNumberString)

	// var err error

	_, err = db.NamedExec("INSERT INTO assets_ms (asset_id, asset_uuid, asset_code, asset_name, serial_number, asset_specification, procurement_date, price, asset_description, system_classification, asset_location, asset_status, created_by) VALUES (:asset_id, :asset_uuid, :asset_code, :asset_name, :serial_number, :asset_specification, :procurement_date, :price, :asset_description, :system_classification, :asset_location, :asset_status, :created_by)", map[string]interface{}{
		"asset_id":              appID,
		"asset_uuid":            uuidString,
		"asset_code":            assetCode,
		"asset_name":            addAsset.NamaAsset,
		"serial_number":         addAsset.SerialNumber,
		"asset_specification":   addAsset.Spesifikasi,
		"procurement_date":      addAsset.TglPengadaan,
		"price":                 addAsset.Harga,
		"asset_description":     addAsset.Deskripsi,
		"system_classification": addAsset.Klasifikasi,
		"asset_location":        addAsset.Lokasi,
		"asset_status":          addAsset.Status,
		"created_by":            username,
		"created_at":            currentTimestamp,
	})

	if err != nil {
		return err
	}

	for _, pic := range pic {
		uuidString := uuid.New().String()

		log.Printf("Inserting PIC: %+v\n", map[string]interface{}{
			"pic_uuid":        uuidString,
			"asset_id":        appID,
			"pic_name":        pic.NamaPic,
			"pic_description": pic.Keterangan,
			"created_by":      username,
		})

		_, err := db.NamedExec("INSERT INTO pic_ms (pic_uuid, asset_id, pic_name, pic_description, created_by) VALUES (:pic_uuid, :asset_id, :pic_name, :pic_description, :created_by)", map[string]interface{}{
			"pic_uuid":        uuidString,
			"asset_id":        appID,
			"pic_name":        pic.NamaPic,
			"pic_description": pic.Keterangan,
			"created_by":      username,
		})
		if err != nil {
			log.Println("Error inserting PIC:", err)
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

// type formBA struct {
// 	Asset        models.Asset    `json:"asset"`
// 	Pic []models.Pic `json:"pic"`
// }

func GetAllFormBAAssets() ([]models.FormsBeritaAcara, error) {
	rows, err := db.Query(`
				SELECT 
			f.form_uuid,  f.form_number, f.form_status,
			d.document_name,
			f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
			(f.form_data->>'jenis')::text AS jenis,
			(f.form_data->>'nama_pic')::text AS nama_pic,
			(f.form_data->>'asset_uuid')::text AS asset_uuid,
			(f.form_data->>'kode_asset')::text AS kode_asset,
			(f.form_data->>'jabatan_pic')::text AS jabatan_pic,
			(f.form_data->>'pihak_pertama')::text AS pihak_pertama
			FROM 
			form_ms f
		LEFT JOIN 
			document_ms d ON f.document_id = d.document_id
			WHERE
			d.document_code = 'BA' AND f.deleted_at IS NULL 
		AND  ((f.form_data->>'jenis')::text = 'Peminjaman' OR (f.form_data->>'jenis')::text = 'Pengembalian')
		ORDER BY form_number DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold all form data
	var forms []models.FormsBeritaAcara

	// Iterate through the rows
	for rows.Next() {
		// Scan the row into the Forms struct
		var form models.FormsBeritaAcara
		err := rows.Scan(
			&form.FormUUID,
			&form.FormNumber,
			&form.FormStatus,
			&form.DocumentName,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedBy,
			&form.UpdatedAt,
			&form.DeletedBy,
			&form.DeletedAt,
			&form.Jenis,
			&form.NamaPIC,
			&form.AssetUUID,
			&form.KodeAsset,
			&form.JabatanPIC,
			&form.PihakPertama,
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

func GetAllAssets() ([]models.Asset, error) {
	// 	var formBAWithSignatories formBA

	// 	err := db.Get(&formBAWithSignatories.Asset, `
	//         SELECT
	// 				asset_uuid,
	// 			asset_code
	// 	FROM
	// 	 	assets_ms
	//     `)

	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	err = db.Select(&formBAWithSignatories.Pic, `
	// 	SELECT
	// 	 	pic_name, pic_description
	// 	 FROM
	// 	 	pic_ms ORDER BY pic_description ASC

	// `)
	// if err != nil {
	// return nil, err
	// }

	// return &formBAWithSignatories, nil

	// Query untuk mendapatkan semua aset
	rows, err := db.Query(`SELECT 
		asset_id,
		asset_uuid, 
		asset_code, 
		asset_name,
		serial_number,
		asset_specification,
		procurement_date,
		price,
		asset_description,
		system_classification,
		asset_location,
		asset_status
	FROM 
		assets_ms
WHERE deleted_by IS NULL
ORDER BY CAST(SUBSTRING(asset_code FROM 'LP([0-9]+)$') AS INTEGER) ASC
;
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []models.Asset

	// Mengambil data aset
	for rows.Next() {
		var asset models.Asset
		var assetID string
		err := rows.Scan(
			&assetID,
			&asset.AssetUUID,
			&asset.Kode,
			&asset.NamaAsset,
			&asset.SerialNumber,
			&asset.Spesifikasi,
			&asset.TglPengadaan,
			&asset.Harga,
			&asset.Deskripsi,
			&asset.Klasifikasi,
			&asset.Lokasi,
			&asset.Status,
		)
		if err != nil {
			return nil, err
		}
		asset.AssetID = assetID //kalo gaada code ini nnti eror

		assets = append(assets, asset)
	}

	// Query untuk mendapatkan data PIC
	picRows, err := db.Query(`SELECT 
		asset_id, pic_uuid, pic_name, pic_description 
	FROM 
		pic_ms ORDER BY
    CAST(NULLIF(SUBSTRING(pic_description FROM 5), '') AS INTEGER) ASC;
		`)
	if err != nil {
		return nil, err
	}
	defer picRows.Close()

	// Mengambil data PIC
	picMap := make(map[string][]models.Pic)
	for picRows.Next() {
		var pic models.Pic
		var assetID string
		err := picRows.Scan(
			&assetID,
			&pic.PicUUID,
			&pic.NamaPic,
			&pic.Keterangan,
		)
		if err != nil {
			return nil, err
		}
		// Menambahkan ke picMap berdasarkan asset_id
		picMap[assetID] = append(picMap[assetID], pic)
	}

	// Menggabungkan PIC ke dalam aset
	for i := range assets {
		if pics, exists := picMap[assets[i].AssetID]; exists { // Gunakan asset_id atau yang sesuai
			assets[i].Pic = pics
		}
	}

	return assets, nil
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
			f.user_id = $1 AND d.document_code = 'BA' AND f.project_id IS NOT NULL AND f.deleted_at IS NULL
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Slice to hold all form data
	var forms []models.FormsBA

	fmt.Println("scan woi")
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
			d.document_code = 'BA' AND f.deleted_at IS NULL AND f.project_id IS NOT NULL ORDER BY f.form_number DESC
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

func GetSpecAsset(id string) (models.Asset, error) {
	var specBA models.Asset
	err := db.Get(&specBA, `SELECT 
					asset_uuid, 
					asset_code, 
					asset_name,
					serial_number,
					asset_specification,
					TO_CHAR(procurement_date, 'YYYY-MM-DD') AS procurement_date,
					price,
					asset_description,
					system_classification,
					asset_location,
					asset_status
				FROM 
					assets_ms
        WHERE
            asset_uuid = $1
            AND deleted_at IS NULL
	`, id)

	if err != nil {
		return models.Asset{}, err
	}

	return specBA, nil
}

type FormBAWithSignatories struct {
	Form        models.FormsBAAll    `json:"form"`
	Signatories []models.SignatoryHA `json:"signatories"`
}

func GetSpecAllBA(id string) (*FormBAWithSignatories, error) {
	var formBAWithSignatories FormBAWithSignatories

	err := db.Get(&formBAWithSignatories.Form, `
        SELECT 
            f.form_uuid, 
            f.form_number, 
            f.form_ticket, 
            f.form_status,
            d.document_name,
            p.project_name,
            f.created_by, 
            f.created_at, 
            f.updated_by, 
            f.updated_at, 
            f.deleted_by, 
            f.deleted_at,
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
            f.form_uuid = $1 
            AND d.document_code = 'BA'  
            AND f.deleted_at IS NULL
    `, id)

	if err != nil {
		return nil, err
	}

	err = db.Select(&formBAWithSignatories.Signatories, `
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

	return &formBAWithSignatories, nil
}

type FormBAAssetWithSignatories struct {
	Form        models.FormsBeritaAcara `json:"form"`
	Signatories []models.SignatoryHA    `json:"signatories"`
}

func GetSpecAllBAAssets(id string) (*FormBAAssetWithSignatories, error) {
	var formBAAssetWithSignatories FormBAAssetWithSignatories

	err := db.Get(&formBAAssetWithSignatories.Form, `
        				SELECT 
			f.form_uuid,  f.form_number, f.form_status,
			d.document_name,
			f.created_by, f.created_at, f.updated_by, f.updated_at, f.deleted_by, f.deleted_at,
			(f.form_data->>'jenis')::text AS jenis,
			(f.form_data->>'nama_pic')::text AS nama_pic,
			(f.form_data->>'asset_uuid')::text AS asset_uuid,
			(f.form_data->>'kode_asset')::text AS kode_asset,
			(f.form_data->>'jabatan_pic')::text AS jabatan_pic,
			(f.form_data->>'pihak_pertama')::text AS pihak_pertama,
			(f.form_data->>'jabatan_pihak_pertama')::text AS jabatan_pihak_pertama
			FROM 
			form_ms f
		LEFT JOIN 
			document_ms d ON f.document_id = d.document_id
			WHERE form_uuid = $1 AND 
			d.document_code = 'BA' AND f.deleted_at IS NULL
    `, id)

	if err != nil {
		return nil, err
	}

	err = db.Select(&formBAAssetWithSignatories.Signatories, `
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

	return &formBAAssetWithSignatories, nil
}

type AssetWithPIC struct {
	Asset models.Asset `json:"asset"`
	PIC   []models.Pic `json:"pic"`
}

func GetSpecAllAsset(id string) (*AssetWithPIC, error) {
	var assetWithPIC AssetWithPIC

	err := db.Get(&assetWithPIC.Asset, `
        SELECT 
					asset_uuid, 
					asset_code, 
					asset_name,
					serial_number,
					asset_specification,
					TO_CHAR(procurement_date, 'YYYY-MM-DD') AS procurement_date,
					price,
					asset_description,
					system_classification,
					asset_location,
					asset_status
				FROM 
					assets_ms
        WHERE
            asset_uuid = $1
            AND deleted_at IS NULL
    `, id)

	if err != nil {
		return nil, err
	}

	err = db.Select(&assetWithPIC.PIC, `
        SELECT 
					pic_uuid, pic_name, pic_description 
				FROM 
					pic_ms
        WHERE
            asset_id IN (
                SELECT asset_id 
                FROM assets_ms 
                WHERE asset_uuid = $1
								)`, id)
	if err != nil {
		return nil, err
	}

	return &assetWithPIC, nil
}

func UpdateBA(updateBA models.Form, data models.BA, username string, userID int, isPublished bool, id string, signatories []models.Signatory) (models.Form, error) {
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

	return updateBA, nil
}

func UpdateAsset(assetData models.Asset, dataPIC []models.Pic, username string, userID int) (models.Asset, error) {
	currentTime := time.Now()
	// formStatus := "Draft"
	// if isPublished {
	// 	formStatus = "Published"
	// }
	var err error
	fmt.Println("woii", assetData.AssetUUID)
	// daJSON, err := json.Marshal(assetData)
	// if err != nil {
	// 	log.Println("Error marshaling asset struct:", err)
	// 	return models.Asset{}, err
	// }
	// log.Println("asset JSON:", string(daJSON)) // Periksa hasil marshaling

	_, err = db.NamedExec(`UPDATE assets_ms SET asset_name = :nama_asset, serial_number = :serial_number, asset_specification = :spesifikasi, procurement_date = :tgl_pengadaan, price = :harga, asset_description = :deskripsi, system_classification = :klasifikasi, asset_location = :lokasi, asset_status = :status, updated_by = :updated_by, updated_at = :updated_at WHERE asset_uuid = :id`, map[string]interface{}{
		"nama_asset":    assetData.NamaAsset,
		"serial_number": assetData.SerialNumber,
		"spesifikasi":   assetData.Spesifikasi,
		"tgl_pengadaan": assetData.TglPengadaan,
		"harga":         assetData.Harga,
		"deskripsi":     assetData.Deskripsi,
		"klasifikasi":   assetData.Klasifikasi,
		"lokasi":        assetData.Lokasi,
		"status":        assetData.Status,
		"updated_by":    username,
		"updated_at":    currentTime,
		"id":            assetData.AssetUUID,
	})
	if err != nil {
		fmt.Println("Error during asset update:", err) // Log error untuk debugging
		return models.Asset{}, err
	}

	var assetID string
	err = db.Get(&assetID, "SELECT asset_id FROM assets_ms WHERE asset_uuid = $1", assetData.AssetUUID)
	if err != nil {
		log.Println("Error getting asset_id:", err)
		return models.Asset{}, err
	}

	fmt.Println("iki sek obong obong sopo", assetID)

	_, err = db.Exec("DELETE FROM pic_ms WHERE asset_id = $1", assetID)
	if err != nil {
		log.Println("Error deleting pic_ms records:", err)
		return models.Asset{}, err
	}

	// personalNames, err := GetAllPersonalName()
	// if err != nil {
	// 	log.Println("Error getting personal names:", err)
	// 	return models.Form{}, err
	// }

	for _, pic := range dataPIC {
		uuidString := uuid.New().String()

		// log.Printf("Processing signatory: %+v\n", signatory)
		// var userID string
		// for _, personal := range personalNames {
		// 	if personal.PersonalName == signatory.Name {
		// 		userID = personal.UserID
		// 		break
		// 	}
		// }

		// if userID == "" {
		// 	log.Printf("User ID not found for personal name: %s\n", signatory.Name)
		// 	continue
		// }

		fmt.Println(pic)

		_, err := db.NamedExec("INSERT INTO pic_ms (pic_uuid, asset_id, pic_name, pic_description, created_by) VALUES (:pic_uuid, :asset_id, :pic_name, :pic_description, :created_by)", map[string]interface{}{
			"pic_uuid":        uuidString,
			"asset_id":        assetID,
			"pic_name":        pic.NamaPic,
			"pic_description": pic.Keterangan,
			"created_by":      username,
		})
		if err != nil {
			return models.Asset{}, err
		}
	}

	return assetData, nil
}

func FormBAByDivision(divisionCode string) ([]models.FormsBA, error) {
	var form []models.FormsBA

	// Now use the retrieved documentID in the query
	errSelect := db.Select(&form, `
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
			d.document_code = 'BA' AND f.deleted_at IS NULL AND f.project_id IS NOT NULL AND SPLIT_PART(f.form_number, '/', 2) = $1
		ORDER BY f.form_number DESC;
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

func DeleteBeritaAcara(id string, username string, jenis string) error {
	currentTime := time.Now()

	// Mendapatkan assetUUID dari form
	var assetUUID string
	var picUUID string
	err := db.QueryRow("SELECT form_data->>'asset_uuid' AS asset_uuid, form_data->>'pic_uuid' AS pic_uuid FROM form_ms WHERE form_uuid = $1", id).Scan(&assetUUID, &picUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no asset_uuid found for form_uuid %s", id)
		}
		return err
	}

	// Menentukan status baru berdasarkan jenis
	var newStatus string
	if jenis == "Peminjaman" {
		// Jika jenis penghapusan adalah Peminjaman, set status menjadi Dipinjam
		newStatus = "Tersedia"
	} else if jenis == "Pengembalian" {
		// Jika jenis penghapusan adalah Pengembalian, set status menjadi Tersedia
		newStatus = "Dipinjam"
	} else {
		return fmt.Errorf("unexpected jenis: %s", jenis)
	}

	// Menghapus (soft delete) asset di form_ms
	assetResult, err := db.Exec("UPDATE form_ms SET deleted_by = $1, deleted_at = $2 WHERE form_uuid = $3", username, currentTime, id)
	if err != nil {
		return err
	}

	// Cek apakah ada baris yang terpengaruh
	assetRowsAffected, err := assetResult.RowsAffected()
	if err != nil {
		return err
	}
	if assetRowsAffected == 0 {
		return ErrNotFound
	}

	// Mendapatkan assetID dari asset_uuid
	var assetID int
	err = db.QueryRow("SELECT asset_id FROM assets_ms WHERE asset_uuid = $1", assetUUID).Scan(&assetID)
	if err != nil {
		return fmt.Errorf("error retrieving asset_id for asset_uuid %s: %v", assetUUID, err)
	}

	fmt.Println("new", newStatus)

	// Memperbarui status aset
	fmt.Println("asset id", assetID)
	_, err = db.Exec("UPDATE assets_ms SET asset_status = $1 WHERE asset_id = $2", newStatus, assetID)
	if err != nil {
		return err
	}

	fmt.Println("pic uuid", picUUID)
	// Menghapus entri terkait di pic_ms secara permanen
	if jenis == "Peminjaman" {
		picResult, err := db.Exec("DELETE FROM pic_ms WHERE pic_uuid = $1", picUUID)
		if err != nil {
			return err
		}

		// Cek apakah ada baris yang terpengaruh untuk PIC
		formRowsAffected, err := picResult.RowsAffected()
		if err != nil {
			return err
		}
		if formRowsAffected == 0 {
			return ErrNotFound
		}
	}

	return nil
}

func DeleteAsset(id string, username string) error {
	currentTime := time.Now()
	fmt.Println("aidi", id)

	// Soft delete asset in assets_ms
	assetResult, err := db.Exec("UPDATE assets_ms SET deleted_by = $1, deleted_at = $2 WHERE asset_uuid = $3", username, currentTime, id)
	if err != nil {
		return err
	}

	// Check if any rows were affected for the asset
	assetRowsAffected, err := assetResult.RowsAffected()
	if err != nil {
		return err
	}
	if assetRowsAffected == 0 {
		return ErrNotFound // Return an error if no asset was updated
	}

	// Soft delete related form entries in form_ms
	formResult, err := db.Exec("UPDATE form_ms SET deleted_by = $1, deleted_at = $2 WHERE form_data->>'asset_uuid' = $3", username, currentTime, id)
	if err != nil {
		return err
	}

	// Check if any rows were affected for the form
	formRowsAffected, err := formResult.RowsAffected()
	if err != nil {
		return err
	}
	if formRowsAffected == 0 {
		return ErrNotFound // Return an error if no form was updated
	}

	// Optionally, handle soft deleting related PIC entries
	// deletePICQuery := `
	// 	UPDATE pic_ms
	// 	SET deleted_by = $1, deleted_at = NOW()
	// 	WHERE form_id = (
	// 		SELECT form_id
	// 		FROM form_ms
	// 		WHERE form_uuid = $2
	// 	)
	// `
	// _, err = db.Exec(deletePICQuery, username, id)
	// if err != nil {
	// 	return err
	// }

	return nil
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
		sf.user_id = $1 AND d.document_code = 'BA'  AND f.deleted_at IS NULL
		ORDER BY f.form_number DESC;
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

func GetBACode() (models.DocCodeName, error) {
	var documentCode models.DocCodeName

	err := db.Get(&documentCode, "SELECT document_uuid FROM document_ms WHERE document_code = 'BA'")

	if err != nil {
		return models.DocCodeName{}, err
	}
	return documentCode, nil
}
