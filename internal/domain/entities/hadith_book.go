package entities

type HadithBook struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Available int    `json:"available"`

	Contents []Hadith `json:"contents"`
}
