package pack

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var All_names []string

func Event_func(time []string, endtime []string, start_date string, end_date string, schedule_id string, day_arr []string) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: "default",
		Config: aws.Config{
			Region:      aws.String("us-east-2"),
			Credentials: credentials.NewStaticCredentials(os.Getenv("access_key_id"), os.Getenv("aws_secret_access_key"), os.Getenv("aws_session_token")),
		},
	})

	faceerr(err)

	svc := eventbridge.New(sess)

	//cronExpressions := []string{"18 * * * ? *", "0 12 * * ? *", "0 6 * * ? *"}

	//scheduleExpression := "cron(" + strings.Join(cronExpressions, "|") + ")"

	//fmt.Println(scheduleExpression)

	data_for_lambda2 := schedule_id + ":" + start_date + ":" + end_date
	fmt.Println(data_for_lambda2)
	for i, e := range time {

		rule_name := "Start" + "_" + schedule_id + day_arr[i] + "1"
		All_names = append(All_names, rule_name)

		ruleOutput, err := svc.PutRule(&eventbridge.PutRuleInput{

			Name: aws.String(rule_name),

			ScheduleExpression: aws.String("cron(" + e + ")"),
		})

		faceerr(err)

		// Define the JSON input
		input := map[string]interface{}{
			"Detail": data_for_lambda2,
		}

		jsonInput, err := json.Marshal(input)

		if err != nil {

			panic(err)

		}

		lambdaClient := lambda.New(sess)

		addPermissionInput := &lambda.AddPermissionInput{

			Action: aws.String("lambda:InvokeFunction"),

			FunctionName: aws.String("Renaissance_stop_vm"),

			Principal: aws.String("events.amazonaws.com"),

			SourceArn: ruleOutput.RuleArn,

			StatementId: aws.String(rule_name),
		}

		_, err = lambdaClient.AddPermission(addPermissionInput)

		faceerr(err)

		_, e := svc.PutTargets(&eventbridge.PutTargetsInput{

			Rule: aws.String(rule_name),

			Targets: []*eventbridge.Target{

				{

					Arn: aws.String("arn:aws:lambda:us-east-2:676634049556:function:Renaissance_stop_vm"),

					Id: aws.String("Id49d2492c-7284-4c5b-a209-0140f1e96f5a"),

					Input: aws.String(string(jsonInput)),
				},
			},
		})

		faceerr(e)

	}

	//RULES FOR STOP TIME

	for i, e := range endtime {
		rule_name := "Stop" + "_" + schedule_id + day_arr[i] + "2"

		All_names = append(All_names, rule_name)

		ruleOutput, err := svc.PutRule(&eventbridge.PutRuleInput{

			Name: aws.String(rule_name),

			ScheduleExpression: aws.String("cron(" + e + ")"),
		})

		faceerr(err)

		// Define the JSON input

		input := map[string]interface{}{

			"Detail": data_for_lambda2,
		}

		jsonInput, err := json.Marshal(input)

		if err != nil {

			panic(err)

		}

		lambdaClient := lambda.New(sess)

		addPermissionInput := &lambda.AddPermissionInput{

			Action: aws.String("lambda:InvokeFunction"),

			FunctionName: aws.String("Renaissance_start_vm"),

			Principal: aws.String("events.amazonaws.com"),

			SourceArn: ruleOutput.RuleArn,

			StatementId: aws.String(rule_name),
		}

		_, err = lambdaClient.AddPermission(addPermissionInput)

		faceerr(err)

		_, e := svc.PutTargets(&eventbridge.PutTargetsInput{

			Rule: aws.String(rule_name),

			Targets: []*eventbridge.Target{

				{

					Arn: aws.String("arn:aws:lambda:us-east-2:676634049556:function:Renaissance_start_vm"),

					Id: aws.String("Id49d2492c-7284-4c5b-a209-0140f1e96f5a"),

					Input: aws.String(string(jsonInput)),
				},
			},
		})

		faceerr(e)

	}

	fmt.Println(All_names)

}

func faceerr(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

// -------------------------------------------------------------------------//

// ALL WORKING FUNCTIONS ARE HERE

func Approach2(Schedule string, StartDate string, EndDate string, time_stamp string, schedule_id string) {

	// save timezone
	timeZone := time_stamp

	day_time_divison2 := strings.Split(Schedule, ",")

	//DIVIDE ONE SCHEDULE AND GET DAY,TIME

	for _, i := range day_time_divison2 {

		with_out_braces := strings.Trim(i, "()")
		//schedule_date := strings.Split(with_out_braces, "_")
		//schedule_date_start := schedule_date[0]
		get_info := strings.Split(with_out_braces, "-")
		dd := get_info[0]
		st_time := get_info[1]
		ed_time := get_info[2]

		// GET WEEKDAY
		var exact_day time.Weekday

		switch dd {
		case "Sun":
			exact_day = time.Sunday
		case "Mon":
			exact_day = time.Monday
		case "Tue":
			exact_day = time.Tuesday
		case "Wed":
			exact_day = time.Wednesday
		case "Thu":
			exact_day = time.Thursday
		case "Fri":
			exact_day = time.Friday
		case "Sat":
			exact_day = time.Saturday

		}

		//GET EXACT DATE

		parseddate, _ := time.Parse("2006-01-02", StartDate)

		// Get the weekday of the input date
		currentWeekday := parseddate.Weekday()

		// Calculate the difference between the target and current weekday
		daysToAdd := int(exact_day-currentWeekday+7) % 7

		// Add the necessary number of days to the input date
		targetDate := parseddate.AddDate(0, 0, daysToAdd)

		exact_date := targetDate.Format("2006-01-02")

		//  DIVIDE DATA

		st_time2 := strings.Split(st_time, ":")

		st_hour := st_time2[0]
		st_min := st_time2[1]

		chunk_date := strings.Split(exact_date, "-")

		ed_time2 := strings.Split(ed_time, ":")

		ed_hh := ed_time2[0]
		ed_min := ed_time2[1]

		//fmt.Println(ed_hh, ed_min)

		st_yy := chunk_date[0]

		st_mon := chunk_date[1]

		st_date := chunk_date[2]

		chunk_date_end := strings.Split(EndDate, "-")
		end_month := chunk_date_end[1]
		end_year := chunk_date_end[0]

		// to int
		year, _ := strconv.Atoi(st_yy)
		month, _ := strconv.Atoi(st_mon)
		dt, _ := strconv.Atoi(st_date)
		hh, _ := strconv.Atoi(st_hour)
		min, _ := strconv.Atoi(st_min)

		end_hour, _ := strconv.Atoi(ed_hh)
		end_min, _ := strconv.Atoi(ed_min)

		timezone(year, month, dt, hh, min, timeZone, end_month, end_hour, end_min, StartDate, EndDate, end_year, schedule_id, dd)

	}

}

// ----------------------------------- CONVERT TIMEZONE --------------------------------------//

func timezone(year int, mon int, date int, hour int, minutes int, timezone string, end_mon string, end_hour int, end_min int, startdate string, enddate string, end_year string, schedule_id string, day string) {
	//fmt.Println(day)

	const (
		BST = 01 * 60 * 60

		CET = 01 * 60 * 60
		EET = 02 * 60 * 60

		CAT = 02 * 60 * 60

		EAT = 03 * 60 * 60

		IST = (05*60*60 + 30*60)

		SGT = 8 * 60 * 60

		JST = 9 * 60 * 60

		GMT = 0 * 60 * 60

		UTC = 0 * 60 * 60

		NST = -(02*60*60 + 30*60)
		NDT = -(03*60*60 + 30*60)

		ECT = -(05 * 60 * 60)

		EST = -(05 * 60 * 60)

		CST = -(06 * 60 * 60)

		MST = -(7 * 60 * 60)

		PST = -(8 * 60 * 60)

		AST = -(04 * 60 * 60)

		HST  = -(10 * 60 * 60)
		SST  = -(11 * 60 * 60)
		CDT  = -(5 * 60 * 60)
		MDT  = -(6 * 60 * 60)
		ADT  = -(3 * 60 * 60)
		EDT  = -(4 * 60 * 60)
		PDT  = -(7 * 60 * 60)
		CEST = (2 * 60 * 60)
		AEDT = (11 * 60 * 60)
		AEST = (10 * 60 * 60)
		NZDT = (13 * 60 * 60)
		NZST = (12 * 60 * 60)
	)
	//fmt.Println(mon)

	timestamp := timezone

	// timestamp
	var time_stamp int

	switch timestamp {
	case "BST":
		time_stamp = BST
	case "CET":
		time_stamp = CET
	case "EET":
		time_stamp = EET
	case "CAT":
		time_stamp = CAT
	case "EAT":
		time_stamp = EAT
	case "IST":
		time_stamp = IST
	case "SGT":
		time_stamp = SGT
	case "GMT":
		time_stamp = GMT
	case "JST":
		time_stamp = JST
	case "UTC":
		time_stamp = UTC
	case "NST":
		time_stamp = NST
	case "ECT":
		time_stamp = ECT
	case "EST":
		time_stamp = EST
	case "CST":
		time_stamp = CST
	case "MST":
		time_stamp = MST
	case "PST":
		time_stamp = PST
	case "AST":
		time_stamp = AST
	case "HST":
		time_stamp = HST
	case "SST":
		time_stamp = SST
	case "CDT":
		time_stamp = CDT
	case "MDT":
		time_stamp = MDT
	case "ADT":
		time_stamp = ADT
	case "EDT":
		time_stamp = EDT
	case "PDT":
		time_stamp = PDT
	case "CEST":
		time_stamp = CEST
	case "AEDT":
		time_stamp = AEDT
	case "AEST":
		time_stamp = AEST
	case "NZDT":
		time_stamp = NZDT
	case "NZST":
		time_stamp = NZST
	case "NDT":
		time_stamp = NDT
	}

	// Define a custom time zone offset
	timeZoneOffset := (time_stamp)

	// Create a fixed time zone using the offset
	timeZone := time.FixedZone("Custom Time Zone", timeZoneOffset)

	// Define the date and time to work with
	t := time.Date(year, time.Month(mon), date, hour, minutes, 0, 0, timeZone)
	t2 := time.Date(year, time.Month(mon), date, end_hour, end_min, 0, 0, timeZone)

	//            yyyy  mm  dd  hh  min sec  tz

	// Create a new fixed time zone with the adjusted offset
	timeZoneWithDST := time.FixedZone("Custom Time Zone with DST", timeZoneOffset)

	// Convert the time to UTC with DST-adjusted time zone

	utc := (t.In(timeZoneWithDST).UTC()).String()
	utc_end := (t2.In(timeZoneWithDST).UTC()).String()

	s_days := (t.In(timeZoneWithDST).UTC()).Weekday()

	day_start := get_days(s_days)

	e_days := (t2.In(timeZoneWithDST).UTC()).Weekday()

	day_end := get_days(e_days)

	// days

	//timeStr := utc.String()

	tocron(utc, day_start, day_end, end_mon, utc_end, startdate, enddate, end_year, schedule_id, day)

}

//--------------------------------------- COVERT TO CRON JOBS ------------------------------//

func tocron(utc string, start_day string, end_day string, end_mon string, utc_end string, startdate string, enddate string, end_year string, schedule_id string, day string) {
	//fmt.Println(utc)

	tocron_division := strings.Split(utc, " ")

	date_divison := strings.Split(tocron_division[0], "-")

	cron_yy := date_divison[0]
	cron_mon_start := date_divison[1]

	time_divison := strings.Split(tocron_division[1], ":")

	cron_hh := time_divison[0]
	cron_min := time_divison[1]

	//FOR UTC END
	tocron_division_END := strings.Split(utc_end, " ")

	time_divison_end := strings.Split(tocron_division_END[1], ":")

	cron_hh_end := time_divison_end[0]
	cron_min_end := time_divison_end[1]

	//cron_sec := time_divison[2]

	//fmt.Println(cron_yy, cron_mon, cron_date, cron_hh, cron_min, cron_sec)

	full_cron := cron_min + " " + cron_hh + " " + "?" + " " + cron_mon_start + "-" + end_mon + " " + start_day + " " + cron_yy + "-" + end_year
	full_cron_end := cron_min_end + " " + cron_hh_end + " " + "?" + " " + cron_mon_start + "-" + end_mon + " " + end_day + " " + cron_yy + "-" + end_year

	var arr []string
	var arr_end []string
	var day_arr []string
	day_arr = append(day_arr, day)
	arr = append(arr, full_cron)
	arr_end = append(arr_end, full_cron_end)

	Event_func(arr, arr_end, startdate, enddate, schedule_id, day_arr)

}

//-------------------------------- GET DAYS -----------------------------//

func get_days(days time.Weekday) string {
	var interger_day string

	switch days {
	case time.Sunday:
		interger_day = "1"
	case time.Monday:
		interger_day = "2"
	case time.Tuesday:
		interger_day = "3"
	case time.Wednesday:
		interger_day = "4"
	case time.Thursday:
		interger_day = "5"
	case time.Friday:
		interger_day = "6"
	case time.Saturday:
		interger_day = "7"

	}

	return interger_day
}

// ----------------------------- DELETE RULES --------------------------------------//

func Delete_rules(event_name []string) {
	sess, err := session.NewSessionWithOptions(session.Options{

		Profile: "default",

		Config: aws.Config{

			Region: aws.String("us-east-2"),

			Credentials: credentials.NewStaticCredentials(os.Getenv("access_key_id"), os.Getenv("aws_secret_access_key"), os.Getenv("aws_session_token")),
		},
	})

	faceerr(err)

	svc := eventbridge.New(sess)
	//TARGET REMOVE
	for _, ruleName := range event_name {
		listTargetsInput := &eventbridge.ListTargetsByRuleInput{
			Rule: aws.String(ruleName),
		}

		listTargetsOutput, err := svc.ListTargetsByRule(listTargetsInput)
		if err != nil {
			fmt.Println("Error listing rule targets:", err)
			return
		}
		for _, target := range listTargetsOutput.Targets {
			removeTargetsInput := &eventbridge.RemoveTargetsInput{
				Rule: aws.String(ruleName),
				Ids:  []*string{target.Id},
			}

			_, err = svc.RemoveTargets(removeTargetsInput)
			if err != nil {
				fmt.Println("Error removing target:", err)
				return
			}

			fmt.Println("Target removed:", *target.Id)
		}
		// DELETE RULE
		deleteInput := &eventbridge.DeleteRuleInput{
			Name: aws.String(ruleName),
		}

		_, er := svc.DeleteRule(deleteInput)
		if er != nil {
			fmt.Println("Error deleting rule:", er)
			return
		}

	}
}

// ------------------------------- DELETE RESOURCE BASED POLICY --------------------------------//
func Delete_policies(policies_name []string) {

	sess, err := session.NewSessionWithOptions(session.Options{

		Profile: "default",

		Config: aws.Config{

			Region: aws.String("us-east-2"),

			Credentials: credentials.NewStaticCredentials(os.Getenv("access_key_id"), os.Getenv("aws_secret_access_key"), os.Getenv("aws_session_token")),
		},
	})

	faceerr(err)
	// Create a new Lambda service client
	svc := lambda.New(sess)
	var functionName string
	for _, statementID := range policies_name {
		Which_lambda := strings.Split(statementID, "_")
		if Which_lambda[0] == "Start" {
			functionName = "Renaissance_stop_vm"
		} else if Which_lambda[0] == "Stop" {
			functionName = "Renaissance_start_vm"

		}

		removePermissionInput := &lambda.RemovePermissionInput{
			FunctionName: aws.String(functionName),
			StatementId:  aws.String(statementID),
		}

		_, err := svc.RemovePermission(removePermissionInput)
		if err != nil {
			fmt.Printf("Error deleting policy with StatementId '%s': %v\n", statementID, err)
			continue
		}

		fmt.Printf("Policy with StatementId '%s' deleted successfully\n", statementID)
	}

}
