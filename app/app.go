package app

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v3"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Machine struct {
	/*
	   Field names adhere to Go conventions:
	   https://golang.org/ref/spec#Exported_identifiers
	   https://github.com/golang/go/wiki/CodeReviewComments
	   https://github.com/golang/go/wiki/CodeReviewComments#initialisms
	   Note: All fields must be exported for access by sqlx package.
	   for fields named after commands:
	       if command is executable:
	           capitilise first letter (e.g. LsCPU)
	           if command uses options:
	               capitalise options in suffix (e.g. LspciVVV)
	*/
	ID                     int         `db:"ID"`
	UUID                   null.String `db:"uuid"`
	LogID                  null.String `db:"log_id"`
	LogTime                time.Time   `db:"log_time"`
	SerialNum              null.String `db:"serial_num"`
	CPUVendor              null.String `db:"cpu_vendor"`
	CPUFamilyModelStepping null.String `db:"cpu_family_model_stepping"`
	CPUCoreCount           null.String `db:"cpu_core_count"`
	CPUFlags               null.String `db:"cpu_flags"`
	LsCPU                  null.String `db:"lscpu"`
	CPUInfo                null.String `db:"cpuinfo"`
	DMIdecode              null.String `db:"dmidecode"`
	MemInfo                null.String `db:"meminfo"`
	TestName               null.String `db:"test_name"`
	TestCmd                null.String `db:"test_cmd"`
	TimeStart              time.Time   `db:"time_start"`
	TimeEnd                time.Time   `db:"time_end"`
	TestParams             null.String `db:"test_params"`
	TestMetric             null.String `db:"test_metric"`
	TestStatus             null.String `db:"test_status"`
	TestLog                null.String `db:"test_log"`
	CPUs                   map[string]map[string]string
	DIMMs                  map[string]map[string]string
}

type NullString null.String

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
		app.Router.
			Methods("PUT").
			Path("/machines/{id}").
			HandlerFunc(app.updateMachine)
		app.Router.
			Methods("DELETE").
			Path("/machines/{id}").
			HandlerFunc(app.deleteMachine)
		   app.Router.
		       Methods("GET").
		       Path("/machines/{id}/hardware-only").
		       HandlerFunc(app.getMachineDevicesOnly)
	*/
}

func (app *App) createMachine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//reqBody, _ := ioutil.ReadAll(r.body)
	var machine Machine
	//json.Unmarshal(reqBody, &machine)
	json.NewDecoder(r.Body).Decode(&machine)

	SQL_id := "INSERT IGNORE INTO machine.id (" +
		"uuid, serial_num" +
		") VALUES (" +
		":uuid, :serial_num" +
		")"

	SQL_hardware := "INSERT INTO machine.hardware (" +
		"uuid, log_id, cpu_vendor, cpu_family_model_stepping, " +
		"cpu_core_count, cpu_flags, lscpu, cpuinfo, dmidecode, meminfo" +
		") VALUES (" +
		":uuid, :log_id, :cpu_vendor, :cpu_family_model_stepping, " +
		":cpu_core_count, :cpu_flags, :lscpu, :cpuinfo, :dmidecode, :meminfo" +
		")"

	SQL_test := "INSERT INTO machine.test (" +
		"uuid, log_id, test_name, test_cmd, time_start, time_end, " +
		"test_params, test_metric, test_status, test_log, log_time" +
		") VALUES (" +
		":uuid, :log_id, :test_name, :test_cmd, :time_start, :time_end, " +
		":test_params, :test_metric, :test_status, :test_log, :log_time" +
		")"

	tx, err := db.MustBegin()
	if err != nil {
		tx.Rollback()
		log.Fatal(fmt.Sprintf("Database INSERT failed, %s", err))
	}
	_, err = tx.NamedExec(SQL_id, &machine)
	if err != nil {
		tx.Rollback()
		log.Fatal(fmt.Sprintf("id Database INSERT failed, %s", err))
	}
	_, err = tx.NamedExec(SQL_hardware, &machine)
	if err != nil {
		tx.Rollback()
		log.Fatal(fmt.Sprintf("hardware Database INSERT failed, %s", err))
	}
	_, err = tx.NamedExec(SQL_test, &machine)
	if err != nil {
		tx.Rollback()
		log.Fatal(fmt.Sprintf("test Database INSERT failed, %s", err))
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Fatal(fmt.Sprintf("Database COMMIT failed, %s", err))
	}
}

func (app *App) getMachines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	SQL := "SELECT t.id, id.serial_num, hw.uuid, hw.log_id " +
		"FROM xhplconsole.machine_id AS id, xhplconsole.machine_hardware AS hw " +
		"LEFT JOIN ON id.uuid = hw.uuid LEFT JOIN ORDER BY hw.id DESC LIMIT 3"
	machines := []Machine{}
	err = app.Database.Select(&machines, SQL)
	if err != nil {
		log.Fatal(fmt.Sprintf("Database SELECT failed, %s", err))
	}
	json.NewEncoder(w).Encode(machines)
}

func (app *App) getMachine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	var err error
	SQL := "SELECT t.id, id.serial_num, hw.uuid, hw.logid, hw.cpu_vendor, " +
		"hw.cpu_family_model_stepping, hw.cpu_core_count, t.test_name, " +
		"t.test_cmd, t.test_params, t.test_metric, t.test_status FROM " +
		"xhplconsole.machine_hardware AS hw, xhplconsole.machine_test AS t, " +
		"xhplconsole.machine_id AS id LEFT JOIN ON hw.uuid = t.uuid LEFT JOIN " +
		"ON t.uuid = id.uuid " +
		fmt.Sprintf("WHERE t.id = %s", params["id"])
	machine := Machine{}
	err = app.Database.Get(&machine, SQL)
	if err != nil {
		log.Fatal(fmt.Sprintf("Database SELECT failed, %s", err))
	}
	json.NewEncoder(w).Encode(machine)
}

/*
func (app *App) getMachineHardwareOnly(w httpResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    params := mux.Vars(r)

    var err error
    SQL := "SELECT id,uuid,log_id,cpu_vendor,cpu_family_model_stepping," +
        "cpu_core_count,cpu_flags FROM xhplconsole.machine_hardware " +
        fmt.Sprintf("WHERE id = %s", params["id"])
    machine := Machine{}
    err = app.Database.Get(&machine, SQL)
    if err != nil {
        log.Fatal(fmt.Sprintf("Database SELECT failed, %s", err))
    }
    m := make(map[string]map[string]string)
    stanzas := regexp.MustCompile(`\n\n`).Split(machine.CPUInfo, -1)
    for _, stanza := range stanzas {
        kv := make(map[string]string)
        lines := regexp.MustCompile("\n").Split(stanza, -1)
        for _, line := range lines {
            if line != "" {
                fields := strings.FieldsFunc(line, func(r rune) bool {
                    if r == ':' {
                        return true
                    }
                    return false
                })
                k := strings.TrimSpace(fields[0])
                v := ""
                if len(fields) == 2 {
                    v = strings.TrimSpace(fields[1])
                }
                kv[k] = v
            }
        }
        m[kv["processor"]] = kv
    }
    machine.CPUs = m
    json.NewEncoder(w).Encode(machine)
}
*/
