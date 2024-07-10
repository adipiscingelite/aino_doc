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
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func AddProject(c echo.Context) error {
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
	userName := c.Get("user_name").(string) // Mengambil userUUID dari konteks

	// Token yang sudah dideskripsi
	fmt.Println("Token yang sudah dideskripsi:", decrypted)

	fmt.Println("User name:", userName)

	// Lakukan validasi token
	if userName == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Invalid token atau token tidak ditemukan!",
			"status":  false,
		})
	}

	var project models.Project
	if err := c.Bind(&project); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": err.Error(),
			"status":  false,
		})
	}

	whitespace := regexp.MustCompile(`^\s`)
	if whitespace.MatchString(project.ProjectCode) {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Code tidak boleh dimulai dengan spasi!",
			Status:  false,
		})
	}

	if whitespace.MatchString(project.ProjectName) {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Name tidak boleh dimulai dengan spasi!",
			Status:  false,
		})
	}

	if whitespace.MatchString(project.ProjectManager) {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Project Manager tidak boleh dimulai dengan spasi!",
			Status:  false,
		})
	}

	errVal := c.Validate(project)
	if errVal == nil {

		var existingProjectID int
		err := db.Get(&existingProjectID, "SELECT project_id FROM project_ms WHERE project_code = $1 AND deleted_at IS NULL", project.ProjectCode)
		if err == nil {
			return c.JSON(http.StatusUnprocessableEntity, &models.Response{
				Code:    400,
				Message: "Gagal menambahkan project. Project sudah ada!",
				Status:  false,
			})
		} else {
			addProjectErr := service.AddProject(project, userName)

			if addProjectErr != nil {
				log.Print(addProjectErr)
				return c.JSON(http.StatusInternalServerError, &models.Response{
					Code:    500,
					Message: "Terjadi kesalahan internal pada server. Coba beberapa saat lagi",
					Status:  false,
				})
			}
			return c.JSON(http.StatusCreated, &models.Response{
				Code:    201,
				Message: "Berhasil menambahkan project!",
				Status:  true,
			})
		}
	} else {
		log.Print(errVal)
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Data tidak boleh kosong!",
			Status:  false,
		})
	}

}

func GetAllProject(c echo.Context) error {
	projects, err := service.GetAllProject()
	if err != nil {
		log.Print(err)
		response := models.Response{
			Code:    500,
			Message: "Terjadi kesalahan internal server. Mohon coba beberapa saat lagi",
			Status:  false,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
	return c.JSON(http.StatusOK, projects)

}

func ShowProjectById(c echo.Context) error {
	id := c.Param("id")

	var projects models.Projects

	projects, err := service.ShowProjectById(id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Print(err)
			response := models.Response{
				Code:    404,
				Message: "Project tidak ditemukan!",
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

	return c.JSON(http.StatusOK, projects)
}

func UpdateProject(c echo.Context) error {
	id := c.Param("id")
	previousProject, errGet := service.ShowProjectById(id)
	if errGet != nil {
		return c.JSON(http.StatusNotFound, &models.Response{
			Code:    404,
			Message: "Gagal mengupdate project. Project tidak ditemukan!",
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
	userName := c.Get("user_name").(string) // Mengambil userUUID dari konteks

	// Token yang sudah dideskripsi
	fmt.Println("Token yang sudah dideskripsi:", decrypted)

	// User UUID
	fmt.Println("User name:", userName)

	// Lakukan validasi token
	if userName == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "Invalid token atau token tidak ditemukan!",
			"status":  false,
		})
	}

	var editPro models.Project
	if err := c.Bind(&editPro); err != nil {
		return c.JSON(http.StatusBadRequest, &models.Response{
			Code:    400,
			Message: "Data invalid!",
			Status:  false,
		})
	}
	whitespace := regexp.MustCompile(`^\s`)
	if whitespace.MatchString(editPro.ProjectCode) {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Code tidak boleh dimulai dengan spasi!",
			Status:  false,
		})
	}

	if whitespace.MatchString(editPro.ProjectName) {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Name tidak boleh dimulai dengan spasi!",
			Status:  false,
		})
	}

	if whitespace.MatchString(editPro.ProjectManager) {
		return c.JSON(http.StatusUnprocessableEntity, &models.Response{
			Code:    422,
			Message: "Project Manager tidak boleh dimulai dengan spasi!",
			Status:  false,
		})
	}

	err = c.Validate(&editPro)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &models.Response{
			Code:    422,
			Message: "Data tidak boleh kosong!",
			Status:  false,
		})
	}
	if err == nil {
		existingPro, err := service.GetProjectCodeName(id)
		if err != nil {
			log.Printf("Error getting existing project data: %v", err)
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server.",
				Status:  false,
			})
		}

		if editPro.ProjectCode != "" && editPro.ProjectCode != existingPro.ProjectCode {
			existingProID, err := service.GetProjectIDByCode(editPro.ProjectCode)
			if err == nil && strconv.Itoa(existingProID) != id {
				return c.JSON(http.StatusBadRequest, &models.Response{
					Code:    400,
					Message: "Project dengan code tersebut sudah ada! Project tidak boleh sama!",
					Status:  false,
				})
			}
		}

		// Check for uniqueness only if document_name is being updated
		if editPro.ProjectName != "" && editPro.ProjectName != existingPro.ProjectName {
			existingProID, err := service.GetProjectIDByName(editPro.ProjectName)
			if err == nil && strconv.Itoa(existingProID) != id {
				return c.JSON(http.StatusBadRequest, &models.Response{
					Code:    400,
					Message: "Project dengan name tersebut sudah ada! Project tidak boleh sama!",
					Status:  false,
				})
			}
		}

		_, errService := service.UpdateProject(editPro, id, userName)
		if errService != nil {
			log.Println("Kesalahan selama pembaruan:", errService)
			return c.JSON(http.StatusInternalServerError, &models.Response{
				Code:    500,
				Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
				Status:  false,
			})
		}

		log.Println(previousProject)
		return c.JSON(http.StatusOK, &models.Response{
			Code:    200,
			Message: "Project berhasil diperbarui!",
			Status:  true,
		})
	} else {
		log.Println("Kesalahan sebelum pembaruan:", err)
		return c.JSON(http.StatusInternalServerError, &models.Response{
			Code:    500,
			Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
			Status:  false,
		})
	}
}

func DeleteProject(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")
	secretKey := "secretJwToken" // Ganti dengan kunci yang benar

	// Periksa apakah tokenString tidak kosong
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

	// Langkah 1: Mendekripsi token JWE
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

	userName := c.Get("user_name").(string)
	id := c.Param("id")

	_, errGet := service.ShowProjectById(id)
	if errGet != nil {
		log.Println("Kesalahan saat penghapusan:", errGet)
		return c.JSON(http.StatusNotFound, &models.Response{
			Code:    404,
			Message: "Gagal menghapus project. Project tidak ditemukan!",
			Status:  false,
		})
	}

	errService := service.DeleteProject(id, userName)
	if errService != nil {
		log.Println("Kesalahan saat penghapusan:", errService)
		return c.JSON(http.StatusInternalServerError, &models.Response{
			Code:    500,
			Message: "Terjadi kesalahan internal pada server. Mohon coba beberapa saat lagi!",
			Status:  false,
		})

	}

	return c.JSON(http.StatusOK, &models.Response{
		Code:    200,
		Message: "Project berhasil dihapus!",
		Status:  true,
	})

}
