package api

import (
	"VacancyMonitor/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type VacancyFetcher interface {
	FetchVacancies(ctx context.Context, filter models.Filter) ([]models.Vacancy, error)
}

type VacancyFetcherImpl struct {
	HttpClient *http.Client
	UserAgent  string
}

func NewVacancyFetcher(httpClient *http.Client, userAgent string) *VacancyFetcherImpl {
	return &VacancyFetcherImpl{HttpClient: httpClient, UserAgent: userAgent}
}

type VacanciesResponse struct {
	Items []models.Vacancy `json:"items"`
	Found int              `json:"found"`
	Pages int              `json:"pages"`
	Page  int              `json:"page"`
}

// Сборка URL для обращения к
// api.hh.ru/vacansies
func BuildURL(filter models.Filter) *url.URL {

	u := &url.URL{
		Scheme: "https",
		Host:   "api.hh.ru",
		Path:   "/vacancies",
	}
	q := u.Query()

	if filter.Text != "" {
		q.Add("text", filter.Text)
	}

	if filter.Experience != "" {
		q.Add("experience", filter.Experience)
	}

	if filter.Employment != "" {
		q.Add("employment", filter.Employment)
	}

	if filter.Area != "" {
		q.Add("area", filter.Area)
	}

	if filter.Schedule != "" {
		q.Add("schedule", filter.Schedule)
	}

	if filter.Period != 0 {
		q.Add("period", strconv.Itoa(filter.Period))
	}

	u.RawQuery = q.Encode()
	return u
}

func (v *VacancyFetcherImpl) FetchVacancies(ctx context.Context, filter models.Filter) ([]models.Vacancy, error) {
	u := BuildURL(filter)
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", v.UserAgent)

	resp, err := v.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("hh.ru returned status: %s", resp.Status)
	}

	read, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var vac VacanciesResponse
	err = json.Unmarshal(read, &vac)
	if err != nil {
		return nil, err
	}
	return vac.Items, nil
}
