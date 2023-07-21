package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	_ "github.com/jackc/pgx/v5/stdlib"
	api "main.go/API"
)

type VmDetails struct {
	VMName    string `json:"VMName"`
	AccountID string `json:"AccountID"`
	ObjectID  string `json:"ObjectID"`
}
type RequestData struct {
	RITM      string      `json:"RITM"`
	CTASK     string      `json:"CTASK"`
	Action    string      `json:"Action"`
	Name      string      `json:"ScheduleName"`
	VmDetails []VmDetails `json:"VmDetails"`
}

func faceerr(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

func main() {
	lambda.Start(handler)
}

func handler(request RequestData) (events.APIGatewayProxyResponse, error) {
	sess, err := session.NewSessionWithOptions(session.Options{ //sess
		Profile: "default",
		Config: aws.Config{
			Region:      aws.String("us-east-2"),
			Credentials: credentials.NewStaticCredentials(os.Getenv("access_key_id"), os.Getenv("aws_secret_access_key"), os.Getenv("aws_session_token")),
		},
	})
	//database connection
	Time := time.Now()
	logs := fmt.Sprintf("%+v", request)
	db, err := sql.Open("pgx", os.Getenv("PG_DSN"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	Action := request.Action
	var instance_details []string
	var status string
	// Taking vmdetails
	for _, data := range request.VmDetails {
		// vmName := data.VMName
		// accID := data.AccountID
		instance_id := data.ObjectID
		instance_details = append(instance_details, instance_id)
		faceerr(err)
	}
	if Action == "START" {
		svc := ec2.New(sess)
		for _, id := range instance_details {
			input := &ec2.StartInstancesInput{
				InstanceIds: []*string{
					aws.String(id),
				},
				DryRun: aws.Bool(true),
			}
			result, err := svc.StartInstances(input)
			awsErr, ok := err.(awserr.Error)

			//Pushing status to Log table
			logsql := `INSERT INTO "logs"("ritm_id","action", "timestamp","logs")VALUES($1,$2,$3,$4)`
			_, err = db.Exec(logsql, request.RITM, status, Time, logs)
			if err != nil {
				panic(err)
			} else {
				fmt.Println("\n Row inserted  to log table successfully!")
			}

			if ok && awsErr.Code() == "DryRunOperation" {
				// Let's now set dry run to be false. This will allow us to start the instances
				input.DryRun = aws.Bool(false)
				result, err = svc.StartInstances(input)
				if err != nil {
					fmt.Println("Error", err)
				} else {
					fmt.Println("Success", result.StartingInstances)
				}
			} else { // This could be due to a lack of permissions
				fmt.Println("Error", err)
			}
		}
		status = "Successfully started servers "
		// Calling SNOW API
		api.Servicenow_api(request.CTASK, request.Action, status)
	}
	if Action == "STOP" {
		svc := ec2.New(sess)
		for _, id := range instance_details {
			input := &ec2.StopInstancesInput{
				InstanceIds: []*string{
					aws.String(id),
				},
				DryRun: aws.Bool(true),
			}
			result, err := svc.StopInstances(input)
			awsErr, ok := err.(awserr.Error)

			//Pushing status to Log table
			logsql := `INSERT INTO "logs"("ritm_id","action", "timestamp","logs")VALUES($1,$2,$3,$4)`
			_, err = db.Exec(logsql, request.RITM, status, Time, logs)
			if err != nil {
				panic(err)
			} else {
				fmt.Println("\n Row inserted  to log table successfully!")
			}

			if ok && awsErr.Code() == "DryRunOperation" {
				input.DryRun = aws.Bool(false)
				result, err = svc.StopInstances(input)
				if err != nil {
					fmt.Println("Error", err)
				} else {
					fmt.Println("Success", result.StoppingInstances)
				}
			} else {
				fmt.Println("Error", err)
			}
		}
		status = "Successfully stoped servers "
		// Calling SNOW API
		api.Servicenow_api(request.CTASK, request.Action, status)
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "We Got the Request for the On Demand VM hybernation",
	}, nil
}
