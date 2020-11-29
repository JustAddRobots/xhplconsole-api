package app

import (
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	//	"github.com/doug-martin/goqu/v9"
	"github.com/gorilla/mux"
	"github.com/jmoiron/modl"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v3"
	"io/ioutil"
	"log"
	"net/http"
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
	ID              int                 `db:"id" json:"id"`
	UUID            null.String         `db:"uuid" json:"uuid,string"`
	LogID           null.String         `db:"log_id" json:"log_id,string"`
	CPUCoreCount    null.Int            `db:"cpu_core_count" json:"cpu_core_count,int"`
	CPUFamModelStep null.String         `db:"cpu_family_model_stepping" json:"cpu_family_model_stepping,string"`
	CPUFlags        null.String         `db:"cpu_flags" json:"cpu_flags,string"`
	CPUVendor       null.String         `db:"cpu_vendor" json:"cpu_vendor,string"`
	CPUInfo         []map[string]string `db:"cpuinfo" json:"cpuinfo"`
	DMIdecode       map[string][]string `db:"dmidecode" json:"dmidecode"`
	LogTime         time.Time           `db:"log_time"`
	LsCPU           map[string]string   `db:"lscpu" json:"lscpu"`
	MemInfo         map[string]int      `db:"meminfo" json:"meminfo"`
	SerialNum       null.String         `db:"serial_num" json:"serial_num,string"`
	TestCmd         null.String         `db:"test_cmd" json:"test_cmd,string"`
	TimeEnd         time.Time           `db:"time_end" json:"time_start"`
	TestLog         null.String         `db:"test_log" json:"test_log,string"`
	TestMetric      null.String         `db:"test_metric" json:"test_metric,string"`
	TestName        null.String         `db:"test_name" json:"test_name,string"`
	TestParams      map[string]int      `db:"test_params" json:"test_params"`
	TimeStart       time.Time           `db:"time_start" json:"time_start"`
	TestStatus      null.String         `db:"test_status" json:"test_status,string"`
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
		Path("/machines").
		HandlerFunc(app.createMachine)
	/*
		app.Router.
			Methods("GET").
			Path("/machines").
			HandlerFunc(app.getMachines)
		app.Router.
			Methods("GET").
			Path("/machines/{id}").
			HandlerFunc(app.getMachine)
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
	var err error
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	var machine Machine
	err = json.Unmarshal(reqBody, &machine)
	if err != nil {
		log.Fatal(err)
		//log.Fatalf("%s, %s", string(reqBody), err)
	}

	dbmap := modl.NewDbMap(app.Database.DB, modl.MySQLDialect{})
	dbmap.AddTable(Machine{}, "xhpltest").SetKeys(true, "id")
	/*
		s, err := dbmap.CreateTablesIfNotExistsSql()
		if err != nil {
			log.Fatalf("%s, %s", err, s)
		}
	*/
	err = dbmap.CreateTables()
	if err != nil {
		log.Fatal(err)
	}

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

	/*
		SQL := "INSERT INTO xhpltest (" +
			"uuid, log_id, cpu_vendor, cpu_family_model_stepping, " +
			"cpu_core_count, cpu_flags, lscpu, cpuinfo, dmidecode, meminfo " +
			"serial_num, test_name, test_cmd, time_start, time_end, " +
			"test_params, test_metric, test_status, test_log, log_time" +
			") VALUES (" +
			":uuid, :log_id, :cpu_vendor, :cpu_family_model_stepping, " +
			":cpu_core_count, :cpu_flags, :lscpu, :cpuinfo, :dmidecode, :meminfo " +
			":serial_num, :test_name, :test_cmd, :time_start, :time_end, " +
			":test_params, :test_metric, :test_status, :test_log, :log_time" +
			")"

		tx := app.Database.MustBegin()
		_, err = tx.NamedExec(SQL, &machine)
		if err != nil {
			tx.Rollback()
			log.Fatal(fmt.Sprintf("Database INSERT failed, %s", err))
		}
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			log.Fatal(fmt.Sprintf("Database COMMIT failed, %s", err))
		}
		log.Print("Database INSERT complete.")
	*/
}

/*
func (app *App) getMachines(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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

	var err error
	SQL := "SELECT hw.id, hw.uuid, hw.log_id " +
		"FROM xhplconsole.xhpltest hw " +
		"ORDER BY " +
		"hw.id DESC LIMIT 3"
*/ /*
	SQL := "SELECT hw.id, id.serial_num, hw.uuid, hw.log_id " +
		"FROM xhplconsole.machine_id AS id, xhplconsole.machine_hardware AS hw " +
		"LEFT JOIN ON id.uuid = hw.uuid LEFT JOIN ON t.uuid = hw.uuid ORDER BY " +
		"hw.id DESC LIMIT 3"
*/ /*
	machines := []Machine{}
	err = app.Database.Select(&machines, SQL)
	if err != nil {
		log.Fatal(fmt.Sprintf("Database SELECT failed, %s", err))
	}
	json.NewEncoder(w).Encode(machines)
}
*/ /*
func (app *App) getMachine(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	var err error
*/ /*
	SQL := "SELECT t.id, id.serial_num, hw.uuid, hw.logid, hw.cpu_vendor, " +
		"hw.cpu_family_model_stepping, hw.cpu_core_count, t.test_name, " +
		"t.test_cmd, t.test_params, t.test_metric, t.test_status FROM " +
		"xhplconsole.machine_hardware AS hw, xhplconsole.machine_test AS t, " +
		"xhplconsole.machine_id AS id LEFT JOIN ON hw.uuid = t.uuid LEFT JOIN " +
		"ON t.uuid = id.uuid " +
*/ /*
	SQL := "SELECT hw.uuid, hw.logid, hw.cpu_vendor, " +
		"hw.cpu_family_model_stepping, hw.cpu_core_count " +
		"FROM " +
		"xhplconsole.xhpltest AS hw " +
		fmt.Sprintf("WHERE hw.id = %s", params["id"])
	machine := Machine{}
	err = app.Database.Get(&machine, SQL)
	if err != nil {
		log.Fatal(fmt.Sprintf("Database SELECT failed, %s", err))
	}
	json.NewEncoder(w).Encode(machine)
}
*/ /*
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
