package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"reflect"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

type SM map[string]interface{}
type SS []interface{}

var url = "http://localhost:8080"

type Case struct {
	Request string
	Result  interface{}
}

func Test(t *testing.T) {
	db, err := sql.Open("mysql", DSN)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	PrepareTable(db)
	//defer CleanupTests(db)

	newDB, err := NewDbSession(db)
	if err != nil {
		panic(err)
	}
	server := rpc.NewServer()
	err = server.Register(newDB)
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", ":8080")
	defer listener.Close()
	if err != nil {
		panic(err)
	}
	go http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCodec := jsonrpc.NewServerCodec(&HttpConn{
			in:  r.Body,
			out: w,
		})
		w.Header().Set("Content-type", "application/json")
		err := server.ServeRequest(serverCodec)
		if err != nil {
			t.Fatalf("Eror while serving JSOn request")
			http.Error(w, `{"error":"cant serve request"}`, 500)
		} else {
			w.WriteHeader(200)
		}
	}))

	cases := []Case{
		//Get-reuquests
		Case{
			Request: `{"jsonrpc": "2.0", "id": 1, "method": "DbExplorer.Get", "params": [{"All": "","ID":"","Fields": ["login"]}]}`,
			Result: SM{
				"id": 1,
				"result": SS{
					SM{"login": "Egor"},
					SM{"login": "Jeka"},
					SM{"login": "George"},
				},
				"error": nil,
			},
		},
		Case{
			Request: `{"jsonrpc": "2.0", "id": 1, "method": "DbExplorer.Get", "params": [{"All": "","ID":"2","Fields": ["login", "time"]}]}`,
			Result: SM{
				"id": 1,
				"result": SS{
					SM{"login": "Jeka",
						"time": "2019-02-01 13:00:00",
					},
				},
				"error": nil,
			},
		},
		Case{
			Request: `{"jsonrpc": "2.0", "id": 1, "method": "DbExplorer.Get", "params": [{"All": "1","ID":"1","Fields": ["login", "time"]}]}`,
			Result: SM{
				"id": 1,
				"result": SS{
					SM{"uuid": "1",
						"login": "Egor",
						"time":  "2019-01-31 12:00:00",
					},
				},
				"error": nil,
			},
		},
		Case{
			Request: `{"jsonrpc": "2.0", "id": 1, "method": "DbExplorer.Get", "params": [{"All": "0","ID":"1","Fields": ["login", "time"]}]}`,
			Result: SM{
				"id":     1,
				"result": nil,
				"error":  "Error. Use 1 to show all fields.",
			},
		},
		//Post-requests
		Case{
			Request: `{"jsonrpc":"2.0", "id": 1, "method": "DbExplorer.Post", "params": [{"Uuid": "1", "Login":"Ilya", "Time":""}]}`,
			Result: SM{
				"id":     1,
				"result": "updated",
				"error":  nil,
			},
		},
		Case{
			Request: `{"jsonrpc":"2.0", "id": 1, "method": "DbExplorer.Post", "params": [{"Uuid": "1", "Login":"Ilya", "Time":"2019-02-01 11:44:00"}]}`,
			Result: SM{
				"id":     1,
				"result": nil,
				"error":  "You cant update a date of registration",
			},
		},
		//Check update
		Case{
			Request: `{"jsonrpc": "2.0", "id": 1, "method": "DbExplorer.Get", "params": [{"All": "","ID":"1","Fields": ["login", "uuid"]}]}`,
			Result: SM{
				"id": 1,
				"result": SS{
					SM{"uuid": "1",
						"login": "Ilya",
					},
				},
				"error": nil,
			},
		},
		//Put-requests
		Case{
			Request: `{"jsonrpc":"2.0", "id": 2, "method": "DbExplorer.Put", "params": [{"Login":"Egor", "Time":"2019-01-31 15:09:00"}]}`,
			Result: SM{
				"id":     2,
				"result": 4,
				"error":  nil,
			},
		},
		//Check update
		Case{
			Request: `{"jsonrpc": "2.0", "id": 1, "method": "DbExplorer.Get", "params": [{"All": "","ID":"","Fields": ["login"]}]}`,
			Result: SM{
				"id": 1,
				"result": SS{
					SM{"login": "Ilya"},
					SM{"login": "Jeka"},
					SM{"login": "George"},
					SM{"login": "Egor"},
				},
				"error": nil,
			},
		},
		Case{
			Request: `{"jsonrpc":"2.0", "id": 2, "method": "DbExplorer.Put", "params": [{"Login":"Ilya", "Time":"2019-01-31 15:09:00"}]}`,
			Result: SM{
				"id":     2,
				"result": nil,
				"error":  "Error, login: Ilya exists!",
			},
		},
	}

	for ind, cs := range cases {
		var result interface{}
		var expected interface{}

		resp, err := http.Post(url, "application/json", bytes.NewBufferString(cs.Request))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			panic(err)
		}
		data, _ := json.Marshal(cs.Result)
		json.Unmarshal(data, &expected)

		if !reflect.DeepEqual(expected, result) {
			t.Fatalf("[%v] Result not match\nGot: %v\nExpected: %v", ind, result, expected)
		}
	}
}

func PrepareTable(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS userdata;`,

		`CREATE TABLE userdata (
  uuid int NOT NULL AUTO_INCREMENT,
  login varchar(300) NOT NULL,
  time datetime NOT NULL,
  PRIMARY KEY (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,

		`INSERT INTO userdata (login, time) VALUES
		('Egor', '2019-01-31 12:00:00'),
		('Jeka', '2019-02-01 13:00:00'),
		('George', '2019-03-11 14:00:00');`,
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
