package article

import "time"

type EsArticleDBO struct {
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	URL           string      `json:"URL"`
	Address       AddressES   `json:"address"`
	Publisher     PublisherES `json:"publisher"`
	Categories    []string    `json:"categories"`
	People        []PersonES  `json:"people"`
	Links         []string    `json:"links"`
	DatePublished time.Time   `json:"datePublished"`
	//Image       string  `json:"image"`
	// tags any[],
	// companies? [{
	//   name string,
	//   country string,
	// }],
	//emotionalDescription? string,
}

type PublisherES struct {
	Name    string    `json:"name"`
	Address AddressES `json:"address"`
}

type PersonES struct {
	// AddressES AddressES,
	//Type string
	FullName string `json:"fullName"`
}

type AddressES struct {
	Coords  [2]float64 `json:"coords"`
	Country string     `json:"country"`
	City    string     `json:"city"`
}
