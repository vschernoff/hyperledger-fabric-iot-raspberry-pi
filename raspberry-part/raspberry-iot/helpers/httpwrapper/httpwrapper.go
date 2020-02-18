package httpwrapper

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type SendElementStructure struct {
	Data     string                     `json:"data"`
	Url      string                     `json:"url"`
	CheckFcn func(string) (bool, error) `json:"checkfcn"`
}

func PostReq(sendElement *SendElementStructure) (bool, error) {
	var check bool
	resp, err := http.Post(sendElement.Url, "application/json", bytes.NewBuffer([]byte(sendElement.Data)))
	if err != nil {
		return check, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return check, err
	}

	check, err = sendElement.CheckFcn(string(body))

	return check, err
}

func GetReq(url string) (string, error) {
	var responseStr string
	resp, err := http.Get(url)
	if err != nil {
		return responseStr, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return responseStr, err
	}

	return string(body), nil
}
