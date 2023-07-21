package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	// "database/sql"
	// "log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
)

func start(instance_details []string) {
	sess, _ := session.NewSessionWithOptions(session.Options{

		Profile: "default",

		Config: aws.Config{

			Region: aws.String("us-east-2"),

			Credentials: credentials.NewStaticCredentials(os.Getenv("access_key_id"), os.Getenv("aws_secret_access_key"), os.Getenv("aws_session_token")),
		},
	})

	svc := ec2.New(sess)
	for _, id := range instance_details {
		input := &ec2.StopInstancesInput{
			InstanceIds: []*string{
				aws.String(id),
			},
		}
		result, _ := svc.StopInstances(input)

		fmt.Println(result)
	}
}

//DATABASE

func database_check(s string) {
	db, err := sql.Open("pgx", os.Getenv("PG_DSN"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	//check the connection
	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

	defer db.Close()

	/* --------------------------------SELECT--------------------------- */
	a := s
	query := "SELECT machine_id FROM vm_details WHERE schedule_id = $1"
	rows, err := db.Query(query, a)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var values []string

	for rows.Next() {
		var value string

		err = rows.Scan(&value)
		if err != nil {
			log.Fatal(err)
		}

		//fmt.Println(value)
		values = append(values, value)
	}

	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	start(values)

}

func handler(ctx context.Context, event events.CloudWatchEvent) error {
	// Extract the JSON payload from the CloudWatch event

	Name := event.Detail

	// Parse the JSON payload into your struct
	//var eventData = []string{}
	var eventData string
	err := json.Unmarshal([]byte(Name), &eventData)
	if err != nil {
		return fmt.Errorf("failed to parse JSON payload: %v", err)
	}

	// Use the parsed data as needed
	fmt.Println(reflect.TypeOf(Name))
	fmt.Println("Field 1:", eventData)

	// Bharat:2023-05-12:2023-07-20

	Refined_data := strings.Split(eventData, ":")
	schedule_id := Refined_data[0]
	Start_date := Refined_data[1]
	End_date := Refined_data[2]

	fmt.Println(schedule_id)

	today := time.Now()
	parsed_start_date, _ := time.Parse("2006-01-02", Start_date)
	parsed_end_date, _ := time.Parse("2006-01-02", End_date)

	if (today.After(parsed_start_date) || today.Equal(parsed_start_date)) && (today.Before(parsed_end_date) || today.Equal(parsed_end_date)) {
		database_check(schedule_id)
	}
	return nil
}

func main() {
	lambda.Start(handler)

}
