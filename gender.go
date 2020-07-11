package genderBR

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"sync"
)

type NamesAPIResult struct {
	Frequency int `json:"freq"`
}

type Result struct {
	Name      string  `json:"name"`
	Gender    string  `json:"gender"`
	Frequency float64 `json:"frequency"`
	Error     error   `json:"error"`
}

var cacheMx sync.Mutex
var searchCache = make(map[string]NamesAPIResult, 0)

// For returns the a gender for the given names
func For(names []string) []Result {
	results := make([]Result, 0)

	for _, name := range names {
		result := Result{
			Name:   name,
			Gender: "",
		}

		// Search for both genders
		bothResult, bothErr := basicSearch(name, "")
		if bothErr != nil {
			result.Error = bothErr

			results = append(results, result)
			continue
		}

		// Search only for male
		maleResult, maleErr := basicSearch(name, "m")
		if maleErr != nil {
			result.Error = maleErr

			results = append(results, result)
			continue
		}

		// Calculate frequencies
		total := float64(bothResult.Frequency)
		male := float64(maleResult.Frequency)

		malePercentage := male / total
		femalePercentage := math.Abs(malePercentage - 1)

		if malePercentage > femalePercentage {
			result.Gender = "Male"
			result.Frequency = malePercentage
		} else {
			result.Gender = "Female"
			result.Frequency = femalePercentage
		}

		results = append(results, result)
	}

	return results
}

func basicSearch(name string, gender string) (*NamesAPIResult, error) {
	scapedName := url.QueryEscape(name)
	endpoint := fmt.Sprintf("https://servicodados.ibge.gov.br/api/v1/censos/nomes/basica?nome=%s", scapedName)

	if gender != "" {
		endpoint = fmt.Sprintf("%s&sexo=%s", endpoint, gender)
	}

	// Generate the hash
	hash := getHash(endpoint)

	// Verify if already has result
	cacheMx.Lock()
	if result, ok := searchCache[hash]; ok {
		cacheMx.Unlock()
		return &result, nil
	} else {
		cacheMx.Unlock()
	}

	res, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response []NamesAPIResult
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	if len(response) > 0 {
		// Update cache
		cacheMx.Lock()
		searchCache[hash] = response[0]
		cacheMx.Unlock()

		return &response[0], nil
	}

	return nil, errors.New("not found")
}

func getHash(input string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(input)))
}
