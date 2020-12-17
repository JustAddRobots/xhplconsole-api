package app

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	//	"github.com/jmoiron/modl"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v3"
	"io/ioutil"
	"log"
	"net/http"
	//"reflect"
	// "time"
)

// Need to mitigate issues with Unmarshalling into custom types to/from SQL.
// For now this ugly hack of separate structs for GET/POST works...

// GET
type Mach struct {
	ID              int         `db:"id" json:"id"`
	UUID            null.String `db:"uuid" json:"uuid,omitempty"`
	LogID           null.String `db:"log_id" json:"log_id,omitempty"`
	CPUCoreCount    null.Int    `db:"cpu_core_count" json:"cpu_core_count,omitempty"`
	CPUFamModelStep null.String `db:"cpu_family_model_stepping" json:"cpu_family_model_stepping,omitempty"`
	CPUFlags        null.String `db:"cpu_flags" json:"cpu_flags,omitempty"`
	CPUVendor       null.String `db:"cpu_vendor" json:"cpu_vendor,omitempty"`
	CPUInfo         null.String `db:"cpuinfo" json:"cpuinfo,omitempty"`
	DMIDecode       null.String `db:"dmidecode" json:"dmidecode,omitempty"`
	LsCPU           null.String `db:"lscpu" json:"lscpu,omitempty"`
	MemInfo         null.String `db:"meminfo" json:"meminfo,omitempty"`
	SerialNum       null.String `db:"serial_num" json:"serial_num,omitempty"`
	TestCmd         null.String `db:"test_cmd" json:"test_cmd,omitempty"`
	TestLog         null.String `db:"test_log" json:"test_log,omitempty"`
	TestMetric      null.String `db:"test_metric" json:"test_metric,omitempty"`
	TestName        null.String `db:"test_name" json:"test_name,omitempty"`
	TestParams      null.String `db:"test_params" json:"test_params,omitempty"`
	TestStatus      null.String `db:"test_status" json:"test_status,omitempty"`
	TimeEnd         null.String `db:"time_end" json:"time_end,omitempty"`
	TimeStart       null.String `db:"time_start" json:"time_start,omitempty"`
}

// POST
type Machine struct {
	ID              int                 `db:"id" json:"id"`
	UUID            null.String         `db:"uuid" json:"uuid,omitempty"`
	LogID           null.String         `db:"log_id" json:"log_id,omitempty"`
	CPUCoreCount    null.Int            `db:"cpu_core_count" json:"cpu_core_count,omitempty"`
	CPUFamModelStep null.String         `db:"cpu_family_model_stepping" json:"cpu_family_model_stepping,omitempty"`
	CPUFlags        null.String         `db:"cpu_flags" json:"cpu_flags,omitempty"`
	CPUVendor       null.String         `db:"cpu_vendor" json:"cpu_vendor,omitempty"`
	CPUInfo         []map[string]string `db:"cpuinfo" json:"cpuinfo,omitempty"`
	DMIDecode       map[string][]string `db:"dmidecode" json:"dmidecode,omitempty"`
	LsCPU           map[string]string   `db:"lscpu" json:"lscpu,omitempty"`
	MemInfo         map[string]int      `db:"meminfo" json:"meminfo,omitempty"`
	SerialNum       null.String         `db:"serial_num" json:"serial_num,omitempty"`
	TestCmd         null.String         `db:"test_cmd" json:"test_cmd,omitempty"`
	TestLog         null.String         `db:"test_log" json:"test_log,omitempty"`
	TestMetric      null.String         `db:"test_metric" json:"test_metric,omitempty"`
	TestName        null.String         `db:"test_name" json:"test_name,omitempty"`
	TestParams      map[string]int      `db:"test_params" json:"test_params,omitempty"`
	TestStatus      null.String         `db:"test_status" json:"test_status,omitempty"`
	TimeEnd         null.String         `db:"time_end" json:"time_end"`
	TimeStart       null.String         `db:"time_start" json:"time_start"`
}

type App struct {
	Router   *mux.Router
	Database *sqlx.DB
}

func (app *App) SetupRouter() {
	app.Router.
		Methods("POST").
		Path("/machines").
		HandlerFunc(app.createMachine)
	app.Router.
		Methods("GET").
		Path("/machines").
		HandlerFunc(app.getMachines)
	app.Router.
		Methods("GET").
		Path("/machines/{id}").
		HandlerFunc(app.getMachine)
		/*
			// Need to figure out a good reason to UPDATE a record

					Methods("PUT").
					Path("/machines/{id}").
					HandlerFunc(app.updateMachine)
		*/
	app.Router.
		Methods("DELETE").
		Path("/machines/{id}").
		HandlerFunc(app.deleteMachine)
}

func (app *App) createMachine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("ReadAll failed %s", err)
	}
	var machine Machine
	err = json.Unmarshal(reqBody, &machine)
	if err != nil {
		log.Fatalf("Unmarshal failed %s", err)
	}

	// Marshal maps/lists separately to JSON byte stream
	cpuinfo_json, _ := json.Marshal(machine.CPUInfo)
	dmidecode_json, _ := json.Marshal(machine.DMIDecode)
	lscpu_json, _ := json.Marshal(machine.LsCPU)
	meminfo_json, _ := json.Marshal(machine.MemInfo)
	testparams_json, _ := json.Marshal(machine.TestParams)

	// Yeah its ugly AF, but will do for now. Need more time to sort it.

	// sqlx.NamedExec() would've been great here, but since maps/lists must be
	// Marshalled, creating another struct just for JSON columns seems just about
	// as ugly. Converting Map/lists -> JSON SQL is surprisingly a PITA in golang.

	tx := app.Database.MustBegin()
	tx.MustExec("INSERT INTO xhpltest ( uuid, log_id, cpu_vendor, cpu_family_model_stepping,  cpu_core_count, cpu_flags, lscpu, cpuinfo, dmidecode,  meminfo, serial_num, test_name, test_cmd, time_start,  time_end, test_params, test_metric, test_status, test_log ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )", machine.UUID, machine.LogID, machine.CPUVendor, machine.CPUFamModelStep, machine.CPUCoreCount, machine.CPUFlags, lscpu_json, cpuinfo_json, dmidecode_json, meminfo_json, machine.SerialNum, machine.TestName, machine.TestCmd, machine.TimeStart, machine.TimeEnd, testparams_json, machine.TestMetric, machine.TestStatus, machine.TestLog)
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Fatalf("Commit failed on INSERT, %s", err)
	}
	log.Print("Database INSERT complete.")
}

func (app *App) getMachines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error

	SQL := "SELECT id, serial_num, uuid, log_id, cpu_vendor, " +
		"cpu_family_model_stepping FROM xhpltest ORDER BY id DESC LIMIT 3"
	var machines []Mach
	rows, err := app.Database.Queryx(SQL)
	if err != nil {
		log.Fatalf("Queryx failed, %s", err)
	}
	err = sqlx.StructScan(rows, &machines)
	if err != nil {
		log.Fatalf("StructScan failed, %s", err)
	}
	json.NewEncoder(w).Encode(machines)
}

func (app *App) getMachine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	var m Mach
	params := mux.Vars(r)

	SQL := "SELECT * FROM xhpltest " +
		fmt.Sprintf("WHERE id = %s", params["id"])
	row := app.Database.QueryRowx(SQL)
	err = row.StructScan(&m)
	if err != nil {
		log.Fatalf("StructScan failed, %s", err)
	}
	/*
		for _, colName := range cols {
			if colName == "lscpu" {
				var b []byte
				err = rw.Scan(&b)
				if err != nil {
					log.Fatalf("Database rw.Scan failed, %s", err)
				}
				err := json.Unmarshal(b, &m.LsCPU)
				if err != nil {
					log.Fatalf("Unmarshal failed, %s", err)
				}
			}
		}
	*/
	json.NewEncoder(w).Encode(m)
}

func (app *App) deleteMachine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	params := mux.Vars(r)

	SQL := "DELETE FROM xhpltest " +
		fmt.Sprintf("WHERE id = %s", params["id"])
	tx := app.Database.MustBegin()
	tx.MustExec(SQL)
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Fatalf("Commit failed on DELETE, %s", err)
	}
}
