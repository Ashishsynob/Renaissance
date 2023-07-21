package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	api "main.go/API"
	"main.go/pack"
)

type VmDetails struct {
	VMName    string `json:"VMName"`
	AccountID string `json:"AccountID"`
	ObjectID  string `json:"ObjectID"`
}

type RequestData struct {
	RITM         string      `json:"RITM"`
	CTASK        string      `json:"CTASK"`
	Action       string      `json:"Action"`
	Name         string      `json:"ScheduleName"`
	Old_schedule string      `json:"OldScheduleName"`
	New_schedule string      `json:"NewScheduleName"`
	Schedule     string      `json:"Schedule"`
	StartDate    string      `json:"StartDate"`
	EndDate      string      `json:"EndDate"`
	TimeZone     string      `json:"TimeZone"`
	VmDetails    []VmDetails `json:"VmDetails"`
}

func handler(request RequestData) (events.APIGatewayProxyResponse, error) {
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
	var status string
	fmt.Print(status)
	logs := fmt.Sprintf("%+v", request)
	Time := time.Now()
	s_id := uuid.New().String()
	schedule_id := s_id[:23]

	//create
	if request.Action == "Create" {
		//Cretaing events custom schedule in EventBridge.

		pack.Approach2(request.Schedule, request.StartDate, request.EndDate, request.TimeZone, schedule_id)
		// Pushing data to Log tables.
		logsql := `INSERT INTO "logs"("ritm_id", "timestamp","logs","action")VALUES($1,$2,$3,$4)`
		_, err = db.Exec(logsql, request.RITM, Time, logs, request.Action)
		if err != nil {
			panic(err)
		} else {
			fmt.Println("\n Row inserted  to log table successfully!")
		}

		//Pushing data to the schedule table.
		schedulesql := `INSERT INTO "schedule"("schedule_id","schedule_name", "start_date", "stop_date","timezone","event_name")VALUES($1,$2,$3,$4,$5,$6)`
		_, err := db.Exec(schedulesql, schedule_id, request.Name, request.StartDate, request.EndDate, request.TimeZone, pack.All_names)
		pack.All_names = nil
		if err != nil {
			panic(err)
		} else {
			fmt.Println("\n Row inserted  to schedule table successfully!")
		}
		res1 := strings.Split(request.Schedule, ",")
		for _, value := range res1 {
			schedule_value := value
			Sc_data1 := strings.Trim(schedule_value, "(")
			Sc_data2 := strings.Trim(Sc_data1, ")")
			Sc_data3 := strings.Split(Sc_data2, "-")
			// Getting schedule values
			Start_day := Sc_data3[0]
			Stop_day := Sc_data3[3]
			Start_time := Sc_data3[1]
			Stop_time := Sc_data3[2]
			Slot_id, _ := uuid.New().Value()

			// Pushing time values to database
			schedule_timesql := `INSERT INTO "schedule_time"("slot_id","schedule_id", "start_time", "stop_time", "start_day","stop_day")VALUES($1,$2,$3,$4,$5,$6)`
			_, err1 := db.Exec(schedule_timesql, Slot_id, schedule_id, Start_time, Stop_time, Start_day, Stop_day)
			if err1 != nil {
				panic(err)
			} else {
				fmt.Println("\n Row inserted  to schedule_time table successfully!")
			}
		}

		for _, data := range request.VmDetails {
			vmName := data.VMName
			accID := data.AccountID
			Machine_id := data.ObjectID
			vm_id, _ := uuid.New().Value()
			queryvmtable := `INSERT INTO "vm_details"("vm_id","schedule_id", "machine_id", "account", "vm_name")VALUES($1,$2,$3,$4,$5)`
			_, err2 := db.Exec(queryvmtable, vm_id, schedule_id, Machine_id, accID, vmName)
			if err2 != nil {
				panic(err)
			} else {
				fmt.Println("\n Row inserted  to vm table successfully!")
			}
		}
		status = "New Schedule created Successfully"
		// Calling SNOW API
		api.Servicenow_api(request.CTASK, request.Name, status)
	}
	if request.Action == "Delete" {
		rows, _ := db.Query("SELECT schedule_id FROM schedule WHERE schedule_name = $1", request.Name)

		defer rows.Close()
		// Iterate over the results
		for rows.Next() {
			var sched_id string
			if err := rows.Scan(&sched_id); err != nil {
				panic(err)
			}
			_, err := db.Exec(`DELETE FROM schedule_time WHERE schedule_id = $1`, sched_id)
			check(err)
			_, err4 := db.Exec(`DELETE FROM vm_details WHERE Schedule_id = $1`, sched_id)
			check(err4)
			_, err5 := db.Exec(`DELETE FROM schedule WHERE Schedule_id = $1`, sched_id)
			check(err5)
			logsql := `INSERT INTO "logs"("ritm_id", "action", "timestamp","logs")VALUES($1,$2,$3,$4)`
			_, err6 := db.Exec(logsql, request.RITM, request.Action, Time, logs)
			check(err6)
		}
		if err := rows.Err(); err != nil {
			check(err)
		}
		status = fmt.Sprintf("Schedule name %s deleted Successfully", request.Name)
		// Calling SNOW API
		api.Servicenow_api(request.CTASK, request.Name, status)
	}
	if request.Action == "Attach" {
		rows, _ := db.Query("SELECT schedule_id FROM schedule WHERE schedule_name = $1", request.Name)
		//check(err4)
		defer rows.Close()
		// Iterate over the results
		for rows.Next() {
			var sched_id string
			if err := rows.Scan(&sched_id); err != nil {
				check(err)
			}
			for _, data := range request.VmDetails {
				vmName := data.VMName
				accID := data.AccountID
				Machine_id := data.ObjectID
				vm_id, err1 := uuid.New().Value()
				check(err1)
				queryvmtable := `INSERT INTO "vm_details"("vm_id","schedule_id", "machine_id", "account", "vm_name")VALUES($1,$2,$3,$4,$5)`
				_, err3 := db.Exec(queryvmtable, vm_id, sched_id, Machine_id, accID, vmName)
				check(err3)
				logsql := `INSERT INTO "logs"("ritm_id", "action","timestamp","logs")VALUES($1,$2,$3,$4)`
				_, err4 := db.Exec(logsql, request.RITM, request.Action, Time, logs)
				check(err4)
			}

		}
		status = "Servers added to the schedule Successfully"
		// Calling SNOW API
		api.Servicenow_api(request.CTASK, request.Name, status)
	}
	if request.Action == "Detach" {
		rows, _ := db.Query("SELECT schedule_id FROM schedule WHERE schedule_name = $1", request.Name)
		//check(err3)
		defer rows.Close()
		// Iterate over the results
		for rows.Next() {
			var sched_id string
			if err := rows.Scan(&sched_id); err != nil {
				check(err)
			}
			for _, data := range request.VmDetails {
				vmName := data.VMName
				_, err := db.Exec(`DELETE FROM vm_details WHERE schedule_id = $1 AND vm_name = $2`, sched_id, vmName)
				check(err)
				logsql := `INSERT INTO "logs"("ritm_id","action","timestamp","logs")VALUES($1,$2,$3,$4)`
				_, err4 := db.Exec(logsql, request.RITM, request.Action, Time, logs)
				check(err4)

			}
		}
		status = "Servers removed from the schedule Successfully"
		// Calling SNOW API
		api.Servicenow_api(request.CTASK, request.Name, status)
	}
	if request.Action == "RenameSchedule" {
		_, err5 := db.Exec("UPDATE schedule SET schedule_name = $1 WHERE schedule_name = $2", request.New_schedule, request.Old_schedule)
		check(err5)
		print("confirm")
		logsql := `INSERT INTO "logs"("ritm_id", "timestamp","logs","action")VALUES($1,$2,$3,$4)`
		_, err4 := db.Exec(logsql, request.RITM, Time, logs, request.Action)
		check(err4)
		status = "Schedule name changed to new schedule name Successfully"
		// Calling SNOW API
		api.Servicenow_api(request.CTASK, request.Action, status)
	}
	if request.Action == "MoveSchedule" {
		Old_schedule, _ := db.Query("SELECT schedule_id FROM schedule WHERE schedule_name = $1", request.Old_schedule)

		var old_id string
		for Old_schedule.Next() {
			err5 := Old_schedule.Scan(&old_id)
			check(err5)
		}
		new_schedule, _ := db.Query("SELECT schedule_id FROM schedule WHERE schedule_name = $1", request.New_schedule)

		var new_id string
		for new_schedule.Next() {
			err6 := new_schedule.Scan(&new_id)
			check(err6)
		}
		for _, data := range request.VmDetails {
			vmName := data.VMName
			result, _ := db.Exec("UPDATE vm_details SET schedule_id = $1 WHERE schedule_id = $2 AND vm_name = $3", new_id, old_id, vmName)
			// check(err)
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				log.Fatal("Failed to retrieve the number of rows affected:", err)
			}
			fmt.Printf("Update successful. %d row(s) affected.\n", rowsAffected)
			logsql := `INSERT INTO "logs"("ritm_id", "timestamp","logs","action")VALUES($1,$2,$3,$4)`
			_, err4 := db.Exec(logsql, request.RITM, Time, logs, request.Action)
			check(err4)
		}
		status = fmt.Sprintf("Servers moved to new schedule %s Successfully", request.New_schedule)
		// Calling SNOW API
		api.Servicenow_api(request.CTASK, request.Name, status)
	}
	if request.Action == "ChangeSchedule" {

		// fetching events to change it from here.
		event_name, err := db.Query("SELECT event_name FROM schedule WHERE schedule_name = $1", request.Name)
		check(err)
		defer event_name.Close()
		schedule_data, err1 := db.Query("SELECT schedule_id FROM schedule WHERE schedule_name = $1", request.Name)
		check(err1)
		defer schedule_data.Close()
		var get_array string
		for event_name.Next() {
			if err := event_name.Scan(&get_array); err != nil {
				panic(err)
			}
		}
		var schedule_id string
		for schedule_data.Next() {
			if err := schedule_data.Scan(&schedule_id); err != nil {
				panic(err)
			}
		}
		// Deleting the Old events
		with_out_braces := strings.Trim(get_array, "{}")
		old_events := strings.Split(with_out_braces, ",")
		fmt.Println(old_events)
		pack.Delete_rules(old_events)
		pack.Delete_policies(old_events)
		pack.Approach2(request.Schedule, request.StartDate, request.EndDate, request.TimeZone, schedule_id)
		fmt.Println(pack.All_names)
		// update
		_, err6 := db.Exec("UPDATE schedule SET event_name = $1 WHERE schedule_name = $2", pack.All_names, request.Name)
		check(err6)
		pack.All_names = nil
		logsql := `INSERT INTO "logs"("ritm_id", "timestamp","logs","action")VALUES($1,$2,$3,$4)`
		_, err7 := db.Exec(logsql, request.RITM, Time, logs, request.Action)
		check(err7)
		//Schedule Table
		_, err8 := db.Exec("UPDATE schedule SET start_date = $1 , stop_date = $2 , timezone = $3 WHERE schedule_name = $4", request.StartDate, request.EndDate, request.TimeZone, request.Name)
		check(err8)
		// Filter the time from data stream
		res1 := strings.Split(request.Schedule, ",")
		for _, value := range res1 {
			schedule_value := value
			Sc_data1 := strings.Trim(schedule_value, "(")
			Sc_data2 := strings.Trim(Sc_data1, ")")
			Sc_data3 := strings.Split(Sc_data2, "-")
			// Getting schedule values
			Start_day := Sc_data3[0]
			Stop_day := Sc_data3[3]
			Start_time := Sc_data3[1]
			Stop_time := Sc_data3[2]
			_, err9 := db.Exec("UPDATE schedule_time SET start_time = $1 , stop_time = $2 , Start_day = $3 , stop_day = $4 WHERE schedule_id = $5", Start_time, Stop_time, Start_day, Stop_day, schedule_id)
			check(err9)
		}
		status = fmt.Sprintf("Schedule time changed to new schedule time Successfully %s", request.Schedule)
		// Calling SNOW API
		api.Servicenow_api(request.CTASK, request.Name, status)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       status,
	}, nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	lambda.Start(handler)
}
