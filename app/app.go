package app

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v3"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Machine struct {
	/*
			Note: All fields must be exported for access by sqlx package.
			for fields named after commands:
		       if command is executable:
		           capitilise first letter (e.g. LsCPU)
		           if command uses options:
		               capitalise options in suffix (e.g. LsPCIVVV)
	*/
	ID              int                 `db:"id" json:"id"`
	UUID            null.String         `db:"uuid" json:"uuid"`
	LogID           null.String         `db:"log_id" json:"log_id"`
	CPUCoreCount    null.Int            `db:"cpu_core_count" json:"cpu_core_count"`
	CPUFamModelStep null.String         `db:"cpu_family_model_stepping" json:"cpu_family_model_stepping"`
	CPUFlags        null.String         `db:"cpu_flags" json:"cpu_flags"`
	CPUVendor       null.String         `db:"cpu_vendor" json:"cpu_vendor"`
	CPUInfo         []map[string]string `db:"cpuinfo" json:"cpuinfo"`
	DMIDecode       map[string][]string `db:"dmidecode" json:"dmidecode"`
	LogTime         time.Time           `db:"log_time"`
	LsCPU           map[string]string   `db:"lscpu" json:"lscpu"`
	MemInfo         map[string]int      `db:"meminfo" json:"meminfo"`
	SerialNum       null.String         `db:"serial_num" json:"serial_num"`
	TestCmd         null.String         `db:"test_cmd" json:"test_cmd"`
	TimeEnd         time.Time           `db:"time_end" json:"time_start"`
	TestLog         null.String         `db:"test_log" json:"test_log"`
	TestMetric      null.String         `db:"test_metric" json:"test_metric"`
	TestName        null.String         `db:"test_name" json:"test_name"`
	TestParams      map[string]int      `db:"test_params" json:"test_params"`
	TimeStart       time.Time           `db:"time_start" json:"time_start"`
	TestStatus      null.String         `db:"test_status" json:"test_status"`
	CPUs            map[string]map[string]string
	DIMMs           map[string]map[string]string
}

type App struct {
	Router   *mux.Router
	Database *sqlx.DB
}

func (app *App) SetupRouter() {
	app.Router.
		Methods("POST").
		Path("v1/machines").
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
			Methods("GET").
			Path("/machines/{id}/hardware-only").
			HandlerFunc(app.getMachineDevicesOnly)
		app.Router.
			Methods("PUT").
			Path("/machines/{id}").
			HandlerFunc(app.updateMachine)
		app.Router.
			Methods("DELETE").
			Path("/machines/{id}").
			HandlerFunc(app.deleteMachine)
	*/
}

func (app *App) createMachine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	var machine Machine
	err = json.Unmarshal(reqBody, &machine)
	if err != nil {
		log.Fatal(err)
	}

	/*
		// modl/DbMap would be cool, but it doesn't encode golang maps or lists
		// to SQL tables. DbMap.Insert() really needs a json.Marshal pre-hook
		// for those types.

		dbmap := modl.NewDbMap(app.Database.DB, modl.MySQLDialect{})
		dbmap.AddTable(Machine{}, "xhpltest").SetKeys(true, "id")

		tx, err := dbmap.Begin()
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
		err = tx.Insert(&machine)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Transaction INSERT failed, %s", err)
		}
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			log.Fatalf("Transaction COMMIT failed, %s", err)
		}
		log.Print("Database INSERT complete.")
	*/

	// Marshal maps/lists to JSON byte stream
	cpuinfo_json, _ := json.Marshal(machine.CPUInfo)
	dmidecode_json, _ := json.Marshal(machine.DMIDecode)
	lscpu_json, _ := json.Marshal(machine.LsCPU)
	meminfo_json, _ := json.Marshal(machine.MemInfo)
	testparams_json, _ := json.Marshal(machine.TestParams)

	// Yeah its an ugly wall of text, but will do for now.
	// sqlx.NamedExec() would've been great here, but since maps/lists must be
	// Marshalled, creating another struct just for JSON columns seems just about
	// as ugly. IMO map/lists -> JSON SQL is suprisingly janky in golang.
	tx := app.Database.MustBegin()
	tx.MustExec("INSERT INTO xhpltest ( uuid, log_id, cpu_vendor, cpu_family_model_stepping,  cpu_core_count, cpu_flags, lscpu, cpuinfo, dmidecode,  meminfo, serial_num, test_name, test_cmd, time_start,  time_end, test_params, test_metric, test_status, test_log,  log_time) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )", machine.UUID, machine.LogID, machine.CPUVendor, machine.CPUFamModelStep, machine.CPUCoreCount, machine.CPUFlags, lscpu_json, cpuinfo_json, dmidecode_json, meminfo_json, machine.SerialNum, machine.TestName, machine.TestCmd, machine.TimeStart, machine.TimeEnd, testparams_json, machine.TestMetric, machine.TestStatus, machine.TestLog, machine.LogTime)

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Fatalf("Database COMMIT failed, %s", err)
	}
	log.Print("Database INSERT complete.")
}

func (app *App) getMachines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error

	SQL := "SELECT id, serial_num, uuid, log_id, cpu_vendor, " +
		"cpu_family_model_stepping FROM xhpltest ORDER BY id DESC LIMIT 3"
	var machines []Machine
	rows, err := app.Database.Queryx(SQL)
	if err != nil {
		log.Fatalf("Database SELECT failed, %s", err)
	}
	err = sqlx.StructScan(rows, &machines)
	if err != nil {
		log.Fatalf("Database StructScan failed, %s", err)
	}
	json.NewEncoder(w).Encode(machines)
}

func (app *App) getMachine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	params := mux.Vars(r)

	SQL := "SELECT uuid, serial_num, log_id, cpu_vendor, cpu_family_model_stepping, " +
		"cpu_core_count, cpu_flags, lscpu, cpuinfo, dmidecode, meminfo, test_name " +
		"test_cmd, time_start, time_end, test_params, test_metric, test_status " +
		"test_log, log_time FROM xhpltest " +
		fmt.Sprintf("WHERE id = %s", params["id"])
	var machine Machine
	err = app.Database.QueryRowx(SQL).StructScan(&machine)
	if err != nil {
		log.Fatalf("Database SELECT/QueryRowx failed, %s", err)
	}
	json.NewEncoder(w).Encode(machine)
}

/*
// Need to figure out how to handle Memory devices on a VM.

func (app *App) getMachineHardwareOnly(w httpResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    var err error
    params := mux.Vars(r)

	SQL := "SELECT uuid, serial_num, logid, cpu_vendor, cpu_family_model_stepping, " +
		"cpu_core_count, cpu_flags, lscpu, cpuinfo, dmidecode, meminfo, test_name " +
		"test_cmd, time_start, time_end, test_params, test_metric, test_status "+
		"test_log, log_time FROM xhplconsole.xhpltest " +
		fmt.Sprintf("WHERE id = %s", params["id"])
    machine := Machine{}
    err = app.Database.Get(&machine, SQL)
    if err != nil {
        log.Fatalf("Database SELECT failed, %s", err)
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
