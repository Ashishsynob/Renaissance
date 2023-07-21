package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type SNOW_Struct struct {
	CTASK            string `json:"CTASK"`
	Schedule_Name    string `json:"Schedule_Name"`
	Response_Message string `json:"Response_Message"`
	Error            string `json:"Error"`
	Status           string `json:"Status"`
}

func Servicenow_api(ID string, request_name string, status string) {
	url := "https://cloudeqincdemo1.service-now.com/oauth_token.do"
	method := "POST"
	message := "Successful"
	err_or := ""

	payload := strings.NewReader("grant_type=password&client_id=0a63458468a76d108b99e8b45659c490&client_secret=%5Bbq1UE6!%7Bx&username=Renaissance_User&password=7si%40%5EH%2CzD54c%5E6b%3D%2C3xk%26%5D%3Cj%3B")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))

	var respData map[string]interface{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	accessToken, ok := respData["access_token"].(string)
	if !ok {
		fmt.Println("Error: access_token not found in response")
		return
	}

	// Print the access_token value
	fmt.Println("Access Token:", accessToken)

	PostReq2SNOW(accessToken, ID, request_name, message, status, err_or)
}

func PostReq2SNOW(access_token string, ctask string, schedule_name string, status string, response_message string, err_or string) {

	url := "https://cloudeqincdemo1.service-now.com/api/x_ceq_hibernate_cl/renaissance/Renaissance_AWS_Status"
	method := "POST"

	payload := strings.NewReader(fmt.Sprint(`{"CTASK":  "`, ctask, `","Schedule_Name":  "`, schedule_name, `","Status":  "`, status, `","Response_Message":  "`, response_message, `","Error":  "`, err_or, `"}`))

	fmt.Println(payload)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprint("Bearer ", access_token))

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
	fmt.Println(res.StatusCode)
}
