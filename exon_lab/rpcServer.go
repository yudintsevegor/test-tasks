package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	_ "github.com/go-sql-driver/mysql"
)

var (
	DSN = "yudintsev:UsePerl_1234@tcp(localhost:3306)/exonLab?charset=utf8"
)

//Get
/*
curl -v -X POST -H "Content-Type: application/json" -H "X-Auth: 123" -d '{"jsonrpc":"2.0", "id": 1, "method": "DbExplorer.Get", "params": [{"All": "","ID":"2","Fields": ["login","uuid"]}]}' http://localhost:8080
*/
//Put
/*
curl -v -X POST -H "Content-Type: application/json" -H "X-Auth: 123" -d '{"jsonrpc":"2.0", "id": 1, "method": "DbExplorer.Put", "params": [{"Login":"jeka", "Time":"2019-01-31 15:09:00"}]}' http://localhost:8080
*/
//Post
/*
curl -v -X POST -H "Content-Type: application/json" -H "X-Auth: 123" -d '{"jsonrpc":"2.0", "id": 1, "method": "DbExplorer.Post", "params": [{"Uuid": "3", "Login":"ilya", "Time":""}]}' http://localhost:8080
*/

type Handler struct {
	rpcServer *rpc.Server
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serverCodec := jsonrpc.NewServerCodec(&HttpConn{
		in:  r.Body,
		out: w,
	})
	w.Header().Set("Content-type", "application/json")
	err := h.rpcServer.ServeRequest(serverCodec)
	if err != nil {
		log.Printf("Error while serving JSON request: %v", err)
		http.Error(w, `{"error":"cant serve request"}`, 500)
	} else {
		w.WriteHeader(200)
	}
}

func main() {
	db, err := sql.Open("mysql", DSN)
	err = db.Ping()

	if err != nil {
		log.Printf("Error while openning DB: %v", err)
	}

	newDB, err := NewDbSession(db)
	server := rpc.NewServer()
	err = server.Register(newDB)
	if err != nil {
		log.Printf("Error during registering: %v", err)
	}

	rpcHandler := &Handler{
		rpcServer: server,
	}

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", rpcHandler)
}
