package db

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
)

type INI struct {
	APIHost        string `json:"apihost"`
	BuildHost      string `json:"buildhost"`
	DockerHost     string `json:"dockerhost"`
	JenkinsHost    string `json:"jenkinshost"`
	XHPLConsoleURL string `json:"xhplconsole_url"`
	KubeConfig     string `json:"kubeconfig"`
	DBPassword     string `json:"dbpassword"`
}

func getJSON(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		log.Fatalf("URL Retrieve failed, %s", err)
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func Conn() (*sqlx.DB, error) {
	var ini INI
	iniURL := "http://hosaka.local/ini/builder.json"
	getJSON(iniURL, &ini)
	dblogin := fmt.Sprintf("root:%s", ini.DBPassword) +
		"@tcp(hosaka.local)" +
		"/xhplconsole?parseTime=true"
	db, err := sqlx.Connect("mysql", dblogin)
	return db, err
}
