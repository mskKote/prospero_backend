package dto

//------------------------------ REQUEST

type SearchString struct {
	Search  string `json:"search"`
	IsExact bool   `json:"isExact"`
}

type SearchPeople struct {
	// имя
	FullName string `json:"fullName"`
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

type SearchLanguage struct {
	Name string `json:"name"`
}

type SearchCategory struct {
	Name string `json:"name"`
}

type GrandFilterRequest struct {
	// Массив поисковых строк, оператор объединения &&
	FilterStrings    []SearchString     `json:"filterStrings"`
	FilterPeople     []SearchPeople     `json:"filterPeople"`
	FilterPublishers []SearchPublishers `json:"filterPublishers"`
	FilterCountry    []SearchCountry    `json:"filterCountry"`
	FilterCategories []SearchCategory   `json:"filterCategories"`
	FilterLanguages  []SearchLanguage   `json:"filterLanguages"`
	FilterTime       SearchTime         `json:"filterTime"`
}
