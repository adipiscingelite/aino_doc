package utils

import (
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
	// Optionally, you could return the error to give each route more control over the status code
}

type JwtCustomClaims struct {
	IdUser string `json:"id_user"`
	jwt.RegisteredClaims
}
