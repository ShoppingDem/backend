package models

type User struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Email       string `json:"email,omitempty"`
	OktaID      string `json:"oktaId"`
}
