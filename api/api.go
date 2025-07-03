package api

import (
	"encoding/json"
	"net/http"
)

type (
	jsonMapResponse  = map[string]any
	jsonListResponse = []any
)

// FetchJsonObject fetches a JSON object from the given URL and decodes it into the given struct
func FetchJsonObject(url string) (jsonMapResponse, error) {
	var jsonResponse jsonMapResponse
	if err := fetchJson(url, &jsonResponse); err != nil {
		return nil, err
	}
	return jsonResponse, nil
}

// FetchJsonArray fetches a JSON array from the given URL and decodes it into the given struct
func FetchJsonArray(url string) (jsonListResponse, error) {
	var jsonResponse jsonListResponse
	if err := fetchJson(url, &jsonResponse); err != nil {
		return nil, err
	}
	return jsonResponse, nil
}

// fetchJson fetches a JSON response from the given URL and decodes it into the given struct
func fetchJson(url string, decodeInto any) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(decodeInto)
	if err != nil {
		return err
	}

	return nil
}
