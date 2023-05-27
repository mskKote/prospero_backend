package dto

//------------------------------ REQUEST

type SearchString struct {
	Search  string `json:"search"`
	IsExact bool   `json:"isExact"`
}

type SearchPeople struct {
	// имя
	Name string `json:"name"`
}

type SearchPublishers struct {
	// название издания
	Name string `json:"name"`
}

type SearchCountry struct {
	// название страны
	Country string `json:"country"`
}

type SearchTime struct {
	// Начало временного диапазона
	Start string `json:"start"`
	// Окончание временного диапазона
	End string `json:"end"`
}

type GrandFilterRequest struct {
	// Массив поисковых строк, оператор объединения &&
	FilterStrings    []SearchString     `json:"filterStrings"`
	FilterPeople     []SearchPeople     `json:"filterPeople"`
	FilterPublishers []SearchPublishers `json:"filterPublishers"`
	FilterCountry    []SearchCountry    `json:"filterCountry"`
	FilterTime       SearchTime         `json:"filterTime"`
}
