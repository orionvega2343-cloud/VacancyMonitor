package models

type Salary struct {
	From     *int   `json:"from"`
	To       *int   `json:"to"`
	Currency string `json:"currency"`
	Gross    bool   `json:"gross"`
}
