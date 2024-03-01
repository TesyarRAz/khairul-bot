package services

import (
	"encoding/json"
	"net/http"

	"github.com/poseisharp/khairul-bot/internal/domain/aggregates"
	"github.com/poseisharp/khairul-bot/internal/domain/entities"
)

const (
	hadithURL = "https://api.hadith.gading.dev"
)

type HadithService struct {
	client *http.Client
}

func NewHadithService() *HadithService {
	return &HadithService{
		client: &http.Client{},
	}
}

func (p *HadithService) GetBooks() ([]entities.HadithBook, error) {
	var data aggregates.ApiResponse[[]entities.HadithBook]

	req, err := http.NewRequest(http.MethodGet, hadithURL+"/books", nil)
	if err != nil {
		return nil, err
	}

	response, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&data)

	if err != nil {
		return nil, err
	}

	return data.Data, nil
}

func (p *HadithService) GetHadith(bookID string, number int) (*entities.Hadith, error) {
	var data aggregates.ApiResponse[entities.Hadith]

	req, err := http.NewRequest(http.MethodGet, hadithURL+"/books/"+bookID+"/"+string(rune(number)), nil)
	if err != nil {
		return nil, err
	}

	response, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&data)

	if err != nil {
		return nil, err
	}

	return &data.Data, nil
}
