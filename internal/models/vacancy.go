package models

type Vacancy struct {
	Id           string     `json:"id"`
	Name         string     `json:"name"`
	AlternateUrl string     `json:"alternate_url"`
	Salary       *Salary    `json:"salary"`
	Area         Area       `json:"area"`
	Experience   Experience `json:"experience"`
}
