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

func AddDA(c echo.Context) error {
	const maxRecursionCount = 1000
	recursionCount := 0 // Set nilai awal untuk recursionCount
	var addFormRequest struct {
		IsPublished bool                 `json:"isPublished"`
		FormData    models.Form          `json:"formData"`
		DA          models.DampakAnalisa `json:"data_da"`
		Signatory   []models.Signatory   `json:"signatories"`
	}

	if err := c.Bind(&addFormRequest); err != nil {
		log.Print(err)
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data tidak valid!",
			Status:  false,
		})
	}

	if len(addFormRequest.Signatory) == 0 || addFormRequest.DA == (models.DampakAnalisa{}) {
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data boleh kosong!",
			Status:  false,
		})
	}

	fmt.Println("Nilai isPublished yang diterima di backend:", addFormRequest.IsPublished)

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
	divisionCode := c.Get("division_code").(string)
	userID := c.Get("user_id").(int) // Mengambil userUUID dari konteks
	userName := c.Get("user_name").(string)
	addFormRequest.FormData.UserID = userID
	addFormRequest.FormData.Created_by = userName
	// addFormRequest.FormData.isProject = false
	// addFormRequest.FormData.projectCode =
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
		addroleErr := service.AddDA(addFormRequest.FormData, addFormRequest.IsPublished, userName, userID, divisionCode, recursionCount, addFormRequest.DA, addFormRequest.Signatory)

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

func GetDACode(c echo.Context) error {
	documentCode, err := service.GetDACode()
	if err != nil {
		log.Print(err)
		response := models.Response{
			Code:    500,
			Message: "Terjadi kesalahan internal server. Mohon coba beberapa saat lagi",
			Status:  false,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	return c.JSON(http.StatusOK, documentCode)
}


func GetAllFormDA(c echo.Context) error {
	form, err := service.GetAllFormDA()
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

func GetAllFormDAbyUser(c echo.Context) error {
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
	roleCode := c.Get("role_code").(string)

	fmt.Println("User ID :", userID)
	fmt.Println("Role code", roleCode)
	form, err := service.GetAllDAbyUserID(userID)
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

func GetAllDAbyAdmin(c echo.Context) error {
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
	roleCode := c.Get("role_code").(string)

	fmt.Println("User ID :", userID)
	fmt.Println("Role code", roleCode)
	form, err := service.GetAllDAbyAdmin()
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

func GetSpecDA(c echo.Context) error {
	id := c.Param("id")

	var getDoc models.Formss

	getDoc, err := service.GetSpecDA(id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Print(err)
			response := models.Response{
				Code:    404,
				Message: "Formulir DA tidak ditemukan!",
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

func GetSpecAllDA(c echo.Context) error {
	id := c.Param("id")

	var getDA []models.FormsDAAll
	getDA, err := service.GetSpecAllDA(id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Print(err)
			response := models.Response{
				Code:    404,
				Message: "Formulir DA tidak ditemukan!",
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
	return c.JSON(http.StatusOK, getDA)
}
func GetSpecAllDAa(c echo.Context) error {
	id := c.Param("id")

	// var getDA []models.FormsDAAll
	formWithSignatories, err := service.GetSpecAllDAa(id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Print(err)
			response := models.Response{
				Code:    404,
				Message: "Formulir DA tidak ditemukan!",
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

	// Siapkan data respons
	responseData := map[string]interface{}{
		"form":        formWithSignatories.Form,
		"signatories": formWithSignatories.Signatories,
	}
	return c.JSON(http.StatusOK, responseData)
}
func UpdateFormDA(c echo.Context) error {
	id := c.Param("id")

	var updateFormRequest struct {
		IsPublished bool                 `json:"isPublished"`
		FormData    models.Form          `json:"formData"`
		DA          models.DampakAnalisa `json:"data_da"`
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

	// Lakukan validasi token
	if userID == 0 && userName == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Invalid token atau token tidak ditemukan!",
			"status":  false,
		})
	}

	whitespace := regexp.MustCompile(`^\s`)
	if whitespace.MatchString(updateFormRequest.FormData.FormTicket) {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Ticket tidak boleh dimulai dengan spasi!",
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

	// Tidak perlu melakukan pemeriksaan err lagi di sini

	previousContent, errGet := service.GetSpecDA(id)
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

	_, errService := service.UpdateFormDA(updateFormRequest.FormData, updateFormRequest.DA, userName, userID, updateFormRequest.IsPublished, id)
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
		Message: "Formulir DA berhasil diperbarui!",
		Status:  true,
	})
}

func SignatureUser(c echo.Context) error {
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
	roleCode := c.Get("role_code").(string)

	fmt.Println("User ID :", userID)
	fmt.Println("Role code", roleCode)
	form, err := service.SignatureUser(userID)
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
