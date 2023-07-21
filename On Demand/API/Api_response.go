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

var payload_SNOW SNOW_Struct

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func Servicenow_api(ID string, request_name string, status string) {
	fmt.Println("Calling Access token")
	url := "https://cloudeqincdemo1.service-now.com/oauth_token.do"
	method := "POST"

	// Checking the type of payload
	payload_SNOW.CTASK = ID
	payload_SNOW.Schedule_Name = request_name
	payload_SNOW.Status = "Successfull"
	payload_SNOW.Response_Message = status
	payload_SNOW.Error = ""

	// Generating Token
	token_payload := strings.NewReader("grant_type=password&client_id=0a63458468a76d108b99e8b45659c490&client_secret=%5Bbq1UE6!%7Bx&username=Renaissance_User&password=7si%40%5EH%2CzD54c%5E6b%3D%2C3xk%26%5D%3Cj%3B")
	client1 := &http.Client{}
	req, err := http.NewRequest(method, url, token_payload)
	check(err)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err1 := client1.Do(req)
	check(err1)
	defer res.Body.Close()

	body, err2 := io.ReadAll(res.Body)
	check(err2)
	// Printing Body
	fmt.Println(string(body))

	var respData map[string]interface{}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// ACCESS TOKEN
	accessToken, ok := respData["access_token"].(string)
	if !ok {
		fmt.Println("Error: access_token not found in response")
		return
	}
	// JSON body
	payload := strings.NewReader(fmt.Sprint(`{"CTASK":  "`, payload_SNOW.CTASK, `","Schedule_Name":  "`, payload_SNOW.Schedule_Name, `","Status": "`, payload_SNOW.Status, `","Response_Message":  "`, payload_SNOW.Response_Message, `","Error":  "`, payload_SNOW.Error, `"}`))
	//Declaring http endpoint
	posturl := "https://cloudeqincdemo1.service-now.com/api/x_ceq_hibernate_cl/renaissance/Renaissance_AWS_Status"
	apireq, err := http.NewRequest("POST", posturl, payload)
	check(err)
	//Adding request headers
	api_client := &http.Client{}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprint("Bearer ", accessToken))
	response, err := api_client.Do(apireq)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	Body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(Body))
	fmt.Println(res.StatusCode)
}
