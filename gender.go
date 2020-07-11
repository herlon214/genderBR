package genderBR

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
)

type NamesAPIResult struct {
	Frequency int `json:"freq"`
}

type Result struct {
	Name            string `json:"name"`
	Gender          string `json:"gender"`
	BothFrequency   int    `json:"bothGendersFrequency"`
	MaleFrequency   int    `json:"maleFrequency"`
	FemaleFrequency int    `json:"femaleFrequency"`
	BothError       error
	MaleErr         error
	FemaleErr       error
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
			result.BothFrequency = 0
			result.BothError = bothErr
		} else {
			result.BothFrequency = bothResult.Frequency
		}

		// Search only for male
		maleResult, maleErr := basicSearch(name, "m")
		if maleErr != nil {
			result.MaleFrequency = 0
			result.MaleErr = maleErr
		} else {
			result.MaleFrequency = maleResult.Frequency
		}

		// Search only for female
		femaleResult, femaleErr := basicSearch(name, "f")
		if femaleErr != nil {
			result.FemaleFrequency = 0
			result.FemaleErr = femaleErr
		} else {
			result.FemaleFrequency = femaleResult.Frequency
		}

		// Calculate frequencies
		if bothErr == nil && maleErr == nil && femaleErr == nil {
			total := float64(bothResult.Frequency)
			male := float64(maleResult.Frequency)
			female := float64(femaleResult.Frequency)

			malePercentage := male / total * 100
			femalePercentage := female / total * 100

			if malePercentage > femalePercentage {
				result.Gender = "Male"
			} else {
				result.Gender = "Female"
			}
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
