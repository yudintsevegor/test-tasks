package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	client = &http.Client{Timeout: time.Second}
	url    = "http://localhost:8080"
)

// CaseResponse
type CR map[string]interface{}

type TL struct {
	Tz    []string
	Login string
}

type Bad struct {
	Tz    []int
	Login string
}

type Case struct {
	Path       string
	Method     string
	Status     int
	Login      string
	Password   string
	Request    TL
	BadRequest Bad
	Result     interface{}
}

func TestTz(t *testing.T) {
	zipFile := "zoneinfo.zip"
	var timeZones = make(map[string]*time.Location)
	var sessions = make(map[string]string)
	var users = make(map[string]User)

	db, err := sql.Open("mysql", DSN)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	PrepareTable(db)
	//defer CleanupTestApis(db)

	zones := []string{"Europe/London", "Europe/Moscow", "Africa/Abidjan"}
	loc := make(map[string]*time.Location)
	for _, zone := range zones {
		location, err := time.LoadLocation(zone)
		if err != nil {
			panic(err)
		}
		loc[zone] = location
	}

	handler := &Handler{
		TimeZones: timeZones,
		Sessions:  sessions,
		UserInfo:  users,
		DB:        db,
	}
	err = handler.OpenZip(zipFile)
	if err != nil {
		panic(err)
	}
	go http.ListenAndServe(":8080", handler)

	cases := []Case{
		//0
		Case{
			Path:     "/",
			Method:   http.MethodGet,
			Status:   http.StatusBadRequest,
			Login:    "first",
			Password: "",
			Request: TL{
				Tz:    []string{},
				Login: "first",
			},
			Result: CR{
				"error": "Empty password",
				"out":   nil,
			},
		},
		//1
		Case{
			Path:     "/",
			Method:   http.MethodGet,
			Status:   http.StatusNotFound,
			Login:    "first",
			Password: "p1",
			Request: TL{
				Tz:    []string{},
				Login: "first",
			},
			Result: CR{
				"error": "NoContent",
				"out":   nil,
			},
		},
		//2
		Case{
			Path:     "/",
			Method:   http.MethodGet,
			Status:   http.StatusUnauthorized,
			Login:    "first",
			Password: "p2",
			Request: TL{
				Tz:    []string{},
				Login: "first",
			},
			Result: CR{
				"error": "Bad password",
				"out":   nil,
			},
		},
		//3
		Case{
			Path:     "/tz",
			Method:   http.MethodGet,
			Status:   http.StatusNotFound,
			Login:    "first",
			Password: "p1",
			Request: TL{
				Tz:    []string{},
				Login: "first",
			},
			Result: CR{
				"error": "NoContent",
				"out":   nil,
			},
		},
		//4
		Case{
			Path:     "/tz",
			Method:   http.MethodPost,
			Status:   http.StatusUnauthorized,
			Login:    "first",
			Password: "p1",
			Request: TL{
				Tz:    []string{},
				Login: "first",
			},
			Result: CR{
				"error": "Unauthorized",
				"out":   nil,
			},
		},
		//5
		Case{
			Path:     "/",
			Method:   http.MethodPost,
			Status:   http.StatusNotFound,
			Login:    "first",
			Password: "p1",
			Request: TL{
				Tz:    []string{},
				Login: "first",
			},
			Result: CR{
				"error": "NoContent",
				"out":   nil,
			},
		},
		//6
		Case{
			Path:     "/",
			Method:   http.MethodGet,
			Status:   http.StatusBadRequest,
			Login:    "first",
			Password: "p1",
			BadRequest: Bad{
				Tz:    []int{1, 2},
				Login: "first",
			},
			Result: CR{
				"error": "Bad JSON-Schema",
				"out":   nil,
			},
		},
		//7
		Case{
			Path:     "/",
			Method:   http.MethodGet,
			Status:   http.StatusNotFound,
			Login:    "newuser",
			Password: "p2",
			Request: TL{
				Tz:    []string{},
				Login: "newuser",
			},
			Result: CR{
				"error": "NoContent",
				"out":   nil,
			},
		},
		//8
		Case{
			Path:     "/",
			Method:   http.MethodPost,
			Status:   http.StatusOK,
			Login:    "first",
			Password: "p1",
			Request: TL{
				Tz:    []string{"Africa/Abidjan"},
				Login: "first",
			},
			Result: CR{
				"out": []CR{
					CR{
						"Location": "Africa/Abidjan",
						"Time":     time.Now().In(loc["Africa/Abidjan"]).Format("2006-01-02T15:04 MST"),
					},
				},
				"error": "",
			},
		},
		//9
		Case{
			Path:     "/",
			Method:   http.MethodPost,
			Status:   http.StatusOK,
			Login:    "first",
			Password: "p1",
			Request: TL{
				Tz:    []string{"Europe/London", "Europe/Moscow"},
				Login: "first",
			},
			Result: CR{
				"out": []CR{
					CR{
						"Location": "Europe/London",
						"Time":     time.Now().In(loc["Europe/London"]).Format("2006-01-02T15:04 MST"),
					},
					CR{
						"Location": "Europe/Moscow",
						"Time":     time.Now().In(loc["Europe/Moscow"]).Format("2006-01-02T15:04 MST"),
					},
				},
				"error": "",
			},
		},
		//10
		Case{
			Path:     "/",
			Method:   http.MethodGet,
			Status:   http.StatusOK,
			Login:    "first",
			Password: "p1",
			Request: TL{
				Tz:    []string{"Europe/London"},
				Login: "first",
			},
			Result: CR{
				"out": []CR{
					CR{
						"Location": "Europe/London",
						"Time":     time.Now().In(loc["Europe/London"]).Format("2006-01-02T15:04 MST"),
					},
					CR{
						"Location": "Europe/Moscow",
						"Time":     time.Now().In(loc["Europe/Moscow"]).Format("2006-01-02T15:04 MST"),
					},
				},
				"error": "",
			},
		},
		//11
		Case{
			Path:     "/",
			Method:   http.MethodPost,
			Status:   http.StatusOK,
			Login:    "newuser",
			Password: "p2",
			Request: TL{
				Tz:    []string{"Europe/London"},
				Login: "newuser",
			},
			Result: CR{
				"out": []CR{
					CR{
						"Location": "Europe/London",
						"Time":     time.Now().In(loc["Europe/London"]).Format("2006-01-02T15:04 MST"),
					},
				},
				"error": "",
			},
		},
	}
	runCases(t, db, cases)
}

func runCases(t *testing.T, db *sql.DB, cases []Case) {
	for idx, item := range cases {
		var (
			err      error
			data     []byte
			result   interface{}
			expected interface{}
			req      *http.Request
		)

		caseName := fmt.Sprintf("case %d: [%s] %s;\n Login: %s\nPassword: %s\n", idx, item.Method, item.Path, item.Login, item.Password)
		if idx == 6 {
			data, err = json.Marshal(item.BadRequest)
		} else {
			data, err = json.Marshal(item.Request)
		}
		if err != nil {
			panic(err)
		}
		reqBody := bytes.NewReader(data)
		req, err = http.NewRequest(item.Method, url+item.Path, reqBody)
		if err != nil {
			panic(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.SetBasicAuth(item.Login, item.Password)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("[%s] request error: %v", caseName, err)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if item.Status == 0 {
			item.Status = http.StatusOK
		}

		if resp.StatusCode != item.Status {
			t.Fatalf("[%s] expected http status %v, got %v", caseName, item.Status, resp.StatusCode)
			continue
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatalf("[%s] cant unpack json: %v", caseName, err)
			continue
		}

		data, err = json.Marshal(item.Result)
		json.Unmarshal(data, &expected)

		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("[%s] results not match\nGot : %#v\nWant: %#v", caseName, result, expected)
			continue
		}
	}
}

func PrepareTable(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS userdata;`,

		`CREATE TABLE userdata (
		  id int NOT NULL AUTO_INCREMENT,
		  login varchar(300) NOT NULL,
		  password varchar(3000) NOT NULL,
		  TZ text NOT NULL,
		  PRIMARY KEY (id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,
	}

	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}

func CleanupTests(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS userdata;`,
	}
	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}
