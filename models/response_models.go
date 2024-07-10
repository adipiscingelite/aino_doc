package models

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
}

type ResponseLogin struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
	Token   string `json:"token"`
}
