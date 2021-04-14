package gandi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
)

const (
	apiUrl     = "https://api.gandi.net/v5/livedns"
	defaultTtl = 1800
	MinTtl     = 300
	MaxTtl     = 2592000
)

type ritems struct {
	Items []*rrset `json:"items"`
}

type rrset struct {
	RrsetType   string   `json:"rrset_type"`
	RrsetTTL    int      `json:"rrset_ttl"`
	RrsetName   string   `json:"rrset_name,omitempty"`
	RrsetHref   string   `json:"rrset_href,omitempty"`
	RrsetValues []string `json:"rrset_values"`
}

func getDomainRecords(zone, location, apiKey string) ([]*rrset, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("%s/domains/%s/records/%s", apiUrl, zone, location), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Apikey %s", apiKey))
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		fmt.Println("getDomainRecords - unexpected status ", res.Status, res.StatusCode)
		return nil, fmt.Errorf("%d - %s", res.StatusCode, res.Status)
	}
	var mdl []*rrset
	if err := json.NewDecoder(res.Body).Decode(&mdl); err != nil {
		fmt.Println("getDomainRecords - error decoding json: ", err)
		return nil, err
	}
	return mdl, nil
}

func setDomainRecords(method, zone, location, apiKey string, data []*rrset) error {
	d := ritems{Items: data}
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	client := &http.Client{}
	req, err := http.NewRequest(method,
		fmt.Sprintf("%s/domains/%s/records/%s", apiUrl, zone, location), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Apikey %s", apiKey))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {

		body, _ := ioutil.ReadAll(res.Body)
		log.Error().Str("status", res.Status).Int("code", res.StatusCode).
			Str("body", string(body)).
			Msgf("setDomainRecords [%s] - unexpected status ", method)
		return fmt.Errorf("%d - %s", res.StatusCode, res.Status)
	}
	return nil
}

func addDomainRecord(zone, location, apiKey string, data rrset) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/domains/%s/records/%s", apiUrl, zone, location), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Apikey %s", apiKey))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {

		body, _ := ioutil.ReadAll(res.Body)
		log.Error().Str("status", res.Status).Int("code", res.StatusCode).
			Str("body", string(body)).
			Msg("addDomainRecord - unexpected status ")
		return fmt.Errorf("%d - %s", res.StatusCode, res.Status)
	}
	return nil
}

func updateDomainRecords(zone, location, apiKey string, data []*rrset) error {
	d := ritems{Items: data}
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/domains/%s/records/%s", apiUrl, zone, location), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Apikey %s", apiKey))
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {

		body, _ := ioutil.ReadAll(res.Body)
		log.Error().Str("status", res.Status).Int("code", res.StatusCode).
			Str("body", string(body)).
			Msg("updateDomainRecords - unexpected status ")
		return fmt.Errorf("%d - %s", res.StatusCode, res.Status)
	}
	return nil
}
