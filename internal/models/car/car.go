package car

import "github.com/guregu/null/v5"

type Car struct {
	RegNum string     `json:"regNum" required:"true" example:"X123XX150"`
	Mark   string     `json:"mark" required:"true" example:"Lada"`
	Model  string     `json:"model" required:"true" example:"Vesta"`
	Year   null.Int16 `json:"year" swaggertype:"integer" example:"2001"`
	Owner  People
}

type People struct {
	Name       string      `json:"name" required:"true" example:"Ivan"`
	Surname    string      `json:"surname" required:"true" example:"Ivanov"`
	Patronymic null.String `json:"patronymic" swaggertype:"string" example:"Ivanovich"`
}

type CarWithOwner struct {
	Id     int        `json:"id" required:"true"`
	RegNum string     `json:"regNum" required:"true"`
	Mark   string     `json:"mark" required:"true"`
	Model  string     `json:"model" required:"true"`
	Year   null.Int16 `json:"year" swaggertype:"integer"`
	People `json:"owner"`
}

type PatchPeople struct {
	Name       null.String `json:"name,omitempty"`
	Surname    null.String `json:"surname,omitempty"`
	Patronymic null.String `json:"patronymic,omitempty"`
}

type PatchCar struct {
	RegNum      null.String `json:"reg_num,omitempty"`
	Mark        null.String `json:"mark,omitempty"`
	Model       null.String `json:"model,omitempty"`
	Year        null.Int16  `json:"year,omitempty"`
	PatchPeople `json:"owner,omitempty"`
}

type CarFilter struct {
	YearFilter       string
	RegNumFilter     string
	ModelFilter      string
	MarkFilter       string
	NameFilter       string
	SurnameFilter    string
	PatronymicFilter string
}

func New(regNum, mark, model string, year null.Int16, name, surname string, patronymic null.String) *Car {
	return &Car{RegNum: regNum, Mark: mark, Model: model, Year: year,
		Owner: People{Name: name, Surname: surname, Patronymic: patronymic}}
}
