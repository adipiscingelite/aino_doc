package controller

import (
	"database/sql"
	"document/models"
	"document/service"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

func AddForm(c echo.Context) error {
	const maxRecursionCount = 1000
	recursionCount := 0 // Set nilai awal untuk recursionCount
	var addFormRequest struct {
		IsPublished bool        `json:"isPublished"`
		FormData    models.Form `json:"formData"`
	}

	if err := c.Bind(&addFormRequest); err != nil {
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data tidak valid!",
			Status:  false,
		})
	}

	fmt.Println("Nilai isPublished yang diterima di backend:", addFormRequest.IsPublished)

	tokenString := c.Request().Header.Get("Authorization")
	secretKey := "secretJwToken"
	// Mendapatkan division code dari token
	// divisionCode, err := service.GetDivisionCode(tokenString)
	// if err != nil {
	// 	log.Print(err)
	// 	// Handle error jika tidak dapat mengambil division code dari token
	// 	return c.JSON(http.StatusInternalServerError, map[string]interface{}{
	// 		"message": "Failed to get division code from token",
	// 	})
	// }

	// // Buat trigger di database dengan menggunakan division code
	// _, err = db.Exec("CREATE TRIGGER generate_surat_trigger AFTER INSERT ON form_ms FOR EACH ROW EXECUTE FUNCTION generate_surat_number($1)", divisionCode)
	// if err != nil {
	// 	// Handle error jika gagal membuat trigger
	// 	return c.JSON(http.StatusInternalServerError, map[string]interface{}{
	// 		"message": "Failed to create trigger",
	// 	})
	// }
	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak ditemukan!",
			"status":  false,
		})
	}

	// Periksa apakah tokenString mengandung "Bearer "
	if !strings.HasPrefix(tokenString, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	// Hapus "Bearer " dari tokenString
	tokenOnly := strings.TrimPrefix(tokenString, "Bearer ")

	//dekripsi token JWE
	decrypted, err := DecryptJWE(tokenOnly, secretKey)
	if err != nil {
		fmt.Println("Gagal mendekripsi token:", err)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	var claims JwtCustomClaims
	errJ := json.Unmarshal([]byte(decrypted), &claims)
	if errJ != nil {
		fmt.Println("Gagal mengurai klaim:", errJ)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}
	divisionCode := c.Get("division_code").(string)
	userID := c.Get("user_id").(int) // Mengambil userUUID dari konteks
	userName := c.Get("user_name").(string)
	addFormRequest.FormData.UserID = userID
	addFormRequest.FormData.Created_by = userName
	// Token yang sudah dideskripsi
	fmt.Println("Token yang sudah dideskripsi:", decrypted)
	fmt.Println("User ID:", userID)
	fmt.Println("User Name:", userName)
	fmt.Println("Division Code:", divisionCode)
	// Lakukan validasi token
	if userID == 0 && userName == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Invalid token atau token tidak ditemukan!",
			"status":  false,
		})
	}

	// Validasi spasi untuk Code, Name, dan NumberFormat
	whitespace := regexp.MustCompile(`^\s`)
	if whitespace.MatchString(addFormRequest.FormData.FormTicket) || whitespace.MatchString(addFormRequest.FormData.FormNumber) {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Ticket atau Nomor tidak boleh dimulai dengan spasi!",
			Status:  false,
		})
	}

	errVal := c.Validate(&addFormRequest.FormData)
	//	addFormRequest.FormData.UserID = userID
	if errVal == nil {
		// Gunakan addFormRequest.IsPublished untuk menentukan apakah menyimpan sebagai draft atau mempublish
		addroleErr := service.AddForm(addFormRequest.FormData, addFormRequest.IsPublished, userName, userID, divisionCode, recursionCount)
		if addroleErr != nil {
			log.Print(addroleErr)
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server. Coba beberapa saat lagi",
				Status:  false,
			})
		}

		return c.JSON(http.StatusCreated, &models.Response{
			Code:    201,
			Message: "Berhasil menambahkan formulir!",
			Status:  true,
		})

	} else {
		fmt.Println(errVal)
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Data tidak boleh kosong!",
			Status:  false,
		})
	}
}

func GetAllForm(c echo.Context) error {
	form, err := service.GetAllForm()
	if err != nil {
		log.Print(err)
		response := models.Response{
			Code:    500,
			Message: "Terjadi kesalahan internal server. Mohon coba beberapa saat lagi",
			Status:  false,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
	return c.JSON(http.StatusOK, form)

}

func MyForm(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")
	secretKey := "secretJwToken"

	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak ditemukan!",
			"status":  false,
		})
	}

	// Periksa apakah tokenString mengandung "Bearer "
	if !strings.HasPrefix(tokenString, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	// Hapus "Bearer " dari tokenString
	tokenOnly := strings.TrimPrefix(tokenString, "Bearer ")

	//dekripsi token JWE
	decrypted, err := DecryptJWE(tokenOnly, secretKey)
	if err != nil {
		fmt.Println("Gagal mendekripsi token:", err)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	var claims JwtCustomClaims
	errJ := json.Unmarshal([]byte(decrypted), &claims)
	if errJ != nil {
		fmt.Println("Gagal mengurai klaim:", errJ)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}
	userID := c.Get("user_id").(int)

	fmt.Println("User ID :", userID)

	myform, err := service.MyForm(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Print(err)
			response := models.Response{
				Code:    404,
				Message: "Form tidak ditemukan!",
				Status:  false,
			}
			return c.JSON(http.StatusNotFound, response)
		} else {
			log.Print(err)
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
				Status:  false,
			})
		}
	}
	return c.JSON(http.StatusOK, myform)
}

func FormByDivision(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")
	secretKey := "secretJwToken"

	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak ditemukan!",
			"status":  false,
		})
	}

	// Periksa apakah tokenString mengandung "Bearer "
	if !strings.HasPrefix(tokenString, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	// Hapus "Bearer " dari tokenString
	tokenOnly := strings.TrimPrefix(tokenString, "Bearer ")

	//dekripsi token JWE
	decrypted, err := DecryptJWE(tokenOnly, secretKey)
	if err != nil {
		fmt.Println("Gagal mendekripsi token:", err)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	var claims JwtCustomClaims
	errJ := json.Unmarshal([]byte(decrypted), &claims)
	if errJ != nil {
		fmt.Println("Gagal mengurai klaim:", errJ)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}
	userID := c.Get("user_id").(int)
	divisionCode := c.Get("division_code").(string)

	fmt.Println("User ID :", userID)
	fmt.Println("Division Code :", divisionCode)

	myform, err := service.FormByDivision(divisionCode)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Print(err)
			response := models.Response{
				Code:    404,
				Message: "Form tidak ditemukan!",
				Status:  false,
			}
			return c.JSON(http.StatusNotFound, response)
		} else {
			log.Print(err)
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
				Status:  false,
			})
		}
	}
	return c.JSON(http.StatusOK, myform)
}

func ShowFormById(c echo.Context) error {
	id := c.Param("id")

	var getDoc models.Forms

	getDoc, err := service.ShowFormById(id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Print(err)
			response := models.Response{
				Code:    404,
				Message: "Formulir tidak ditemukan!",
				Status:  false,
			}
			return c.JSON(http.StatusNotFound, response)
		} else {
			log.Print(err)
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
				Status:  false,
			})
		}
	}

	return c.JSON(http.StatusOK, getDoc)
}

func UpdateForm(c echo.Context) error {

	const maxRecursionCount = 1000
	recursionCount := 0 // Set nilai awal untuk recursionCount
	id := c.Param("id")
	var updateFormRequest struct {
		IsPublished bool        `json:"isPublished"`
		FormData    models.Form `json:"formData"`
	}

	if err := c.Bind(&updateFormRequest); err != nil {
		log.Print("error saat binding:", err)
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data tidak valid!",
			Status:  false,
		})
	}

	tokenString := c.Request().Header.Get("Authorization")
	secretKey := "secretJwToken"

	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak ditemukan!",
			"status":  false,
		})
	}

	// Periksa apakah tokenString mengandung "Bearer "
	if !strings.HasPrefix(tokenString, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	// Hapus "Bearer " dari tokenString
	tokenOnly := strings.TrimPrefix(tokenString, "Bearer ")

	//dekripsi token JWE
	decrypted, err := DecryptJWE(tokenOnly, secretKey)
	if err != nil {
		fmt.Println("Gagal mendekripsi token:", err)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	var claims JwtCustomClaims
	errJ := json.Unmarshal([]byte(decrypted), &claims)
	if errJ != nil {
		fmt.Println("Gagal mengurai klaim:", errJ)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}
	var userID int
	var userName string
	if claims, ok := c.Get("user_id").(int); ok {
		userID = claims
	} else {
		// Jika gagal mengonversi ke int, tangani kesalahan di sini
		log.Println("Tidak dapat mengonversi user_id ke int")
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data tidak valid!",
			Status:  false,
		})
	}

	if name, ok := c.Get("user_name").(string); ok {
		userName = name
	} else {
		// Jika gagal mendapatkan nama pengguna, tangani kesalahan di sini
		log.Println("Tidak dapat mengonversi user_name ke string")
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data tidak valid!",
			Status:  false,
		})
	}

	//updateFormRequest.FormData.UserID = userID

	divisionCode := c.Get("division_code").(string)
	updateFormRequest.FormData.UserID = userID

	var updatedBy sql.NullString
	if userName != "" {
		updatedBy.String = userName
		updatedBy.Valid = true
	} else {
		updatedBy.Valid = false
	}

	updateFormRequest.FormData.Updated_by = updatedBy

	// Token yang sudah dideskripsi
	fmt.Println("Token yang sudah dideskripsi:", decrypted)
	fmt.Println("User ID:", userID)
	fmt.Println("user name: ", userName)
	fmt.Println("division code:", divisionCode)

	// Lakukan validasi token
	if userID == 0 && userName == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Invalid token atau token tidak ditemukan!",
			"status":  false,
		})
	}

	// if userID != updateFormRequest.FormData.UserID {
	// 	return c.JSON(http.StatusUnauthorized, map[string]interface{}{
	// 		"code":    401,
	// 		"message": "Anda tidak diizinkan untuk memperbarui formulir ini",
	// 		"status":  false,
	// 	})
	// }
	whitespace := regexp.MustCompile(`^\s`)
	if whitespace.MatchString(updateFormRequest.FormData.FormTicket) {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Ticket tidak boleh dimulai dengan spasi!",
			Status:  false,
		})
	}

	if whitespace.MatchString(updateFormRequest.FormData.FormNumber) {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Name tidak boleh dimulai dengan spasi!",
			Status:  false,
		})
	}

	if err := c.Validate(&updateFormRequest.FormData); err != nil {
		return c.JSON(http.StatusInternalServerError, &models.Response{
			Code:    422,
			Message: "Data tidak boleh kosong!",
			Status:  false,
		})
	}

	previousContent, errGet := service.ShowFormById(id)
	if errGet != nil {
		log.Print(errGet)
		return c.JSON(http.StatusNotFound, &models.Response{
			Code:    404,
			Message: "Gagal mengupdate formulir. Formulir tidak ditemukan!",
			Status:  false,
		})
	}
	if previousContent.FormStatus == "Published" {
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Tidak dapat memperbarui dokumen yang sudah dipublish",
			Status:  false,
		})
	}

	_, errService := service.UpdateForm(updateFormRequest.FormData, id, updateFormRequest.IsPublished, userName, userID, divisionCode, recursionCount)
	if errService != nil {
		log.Println("Kesalahan selama pembaruan:", errService)
		if errService.Error() == "You are not authorized to update this form" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "Anda tidak diizinkan untuk memperbarui formulir ini",
				"status":  false,
			})
		} else {
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
				Status:  false,
			})
		}
	}

	log.Println(previousContent)
	return c.JSON(http.StatusOK, &models.Response{
		Code:    200,
		Message: "Formulir berhasil diperbarui!",
		Status:  true,
	})
}

func DeleteForm(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")
	secretKey := "secretJwToken"

	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak ditemukan!",
			"status":  false,
		})
	}

	if !strings.HasPrefix(tokenString, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	tokenOnly := strings.TrimPrefix(tokenString, "Bearer ")

	//dekripsi token JWE
	decrypted, err := DecryptJWE(tokenOnly, secretKey)
	if err != nil {
		fmt.Println("Gagal mendekripsi token:", err)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	var claims JwtCustomClaims
	errJ := json.Unmarshal([]byte(decrypted), &claims)
	if errJ != nil {
		fmt.Println("Gagal mengurai klaim:", errJ)
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Token tidak valid!",
			"status":  false,
		})
	}

	var userName string

	if name, ok := c.Get("user_name").(string); ok {
		userName = name
	} else {
		// Jika gagal mendapatkan nama pengguna, tangani kesalahan di sini
		log.Println("Tidak dapat mengonversi user_name ke string")
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data tidak valid!",
			Status:  false,
		})
	}
	id := c.Param("id")

	perviousContent, errGet := service.ShowFormById(id)
	if errGet != nil {
		log.Print(errGet)
		return c.JSON(http.StatusNotFound, &models.Response{
			Code:    404,
			Message: "Gagal menghapus formulir. Formulir tidak ditemukan!",
			Status:  false,
		})
	}

	errService := service.DeleteForm(id, userName)
	if errService != nil {
		log.Println("Kesalahan selama pembaruan:", errService)
		if errService.Error() == "You are not authorized to update this form" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "Anda tidak diizinkan untuk menghapus formulir ini",
				"status":  false,
			})
		} else {
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
				Status:  false,
			})
		}
	}

	log.Println(perviousContent)
	return c.JSON(http.StatusOK, &models.Response{
		Code:    200,
		Message: "Formulir berhasil dihapus!",
		Status:  true,
	})
}
