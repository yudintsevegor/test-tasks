package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xeipuuv/gojsonschema"
)

var (
	jsonSchema = `{
	"type": "object",
	"properties": {
		"Login": {
			"type": "string"
		},
		"Tz": {
			"type": "array",
			"items": {"type": "string"}
		}
	}
}`
	schemaLoader = gojsonschema.NewStringLoader(jsonSchema)
	mu           = &sync.Mutex{}
)

func (h *Handler) handleAuth(w http.ResponseWriter, r *http.Request) {
	login, password, _ := r.BasicAuth()
	if login == "" {
		http.Redirect(w, r, "http://localhost:8080/tz", http.StatusTemporaryRedirect)
		return
	}
	if password == "" {
		SendError(w, http.StatusBadRequest, fmt.Errorf("Empty password"))
		return
	}

	mu.Lock()
	notSame := (password != h.UserInfo[login].Password)
	mu.Unlock()

	inDB, err := h.getDataFromDb(login)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err)
		return
	}

	mu.Lock()
	notSame = password != h.UserInfo[login].Password
	mu.Unlock()

	if inDB && notSame {
		SendError(w, http.StatusUnauthorized, fmt.Errorf("Bad password"))
		return
	}
	if inDB {
		http.Redirect(w, r, "http://localhost:8080/tz", http.StatusTemporaryRedirect)
		return
	}

	mu.Lock()
	h.Sessions[login] = ""
	slice := make([]string, 0, 1)
	user := User{
		Login:    login,
		Password: password,
		Tz:       slice,
	}
	h.UserInfo[login] = user
	mu.Unlock()

	err = h.putDataToDb(user)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err)
		return
	}
	http.Redirect(w, r, "http://localhost:8080/tz", http.StatusTemporaryRedirect)
}

func (h *Handler) handleGetTz(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err)
		return
	}

	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err)
		return
	}
	loader := gojsonschema.NewBytesLoader(body)
	result, err := schema.Validate(loader)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err)
		return
	}
	if !result.Valid() {
		SendError(w, http.StatusBadRequest, fmt.Errorf("Bad JSON-Schema"))
		return
	}

	st := &FromRequest{}
	err = json.Unmarshal(body, &st)
	if err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	if st.Login == "" {
		mu.Lock()
		out := h.getTz(st.Tz)
		mu.Unlock()
		if len(out) == 0 {
			SendError(w, http.StatusNotFound, fmt.Errorf("NoContent"))
			return
		}
		SendData(w, http.StatusOK, out)
		return
	}

	if _, ok := h.Sessions[st.Login]; !ok {
		SendError(w, http.StatusUnauthorized, fmt.Errorf("Unauthorized"))
		return
	}

	if r.Method == http.MethodGet {
		out := h.getTz(h.UserInfo[st.Login].Tz)
		mu.Lock()
		delete(h.Sessions, st.Login)
		mu.Unlock()
		if len(out) == 0 {
			SendError(w, http.StatusNotFound, fmt.Errorf("NoContent"))
			return
		}
		SendData(w, http.StatusOK, out)
		return
	}

	if r.Method == http.MethodPost {
		err = h.updateDataInDb(st.Login, st.Tz)
		if err != nil {
			SendError(w, http.StatusInternalServerError, err)
			return
		}
		out := h.getTz(st.Tz)
		if len(out) == 0 {
			SendError(w, http.StatusNotFound, fmt.Errorf("NoContent"))
			return
		}
		mu.Lock()
		delete(h.Sessions, st.Login)
		mu.Unlock()
		SendData(w, http.StatusOK, out)
		return
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.handleAuth(w, r)
	case "/tz":
		h.handleGetTz(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	//Copied from GOROOT/go/...
	zipFile := "zoneinfo.zip"
	var timeZones = make(map[string]*time.Location)
	var sessions = make(map[string]string)
	var users = make(map[string]User)

	db, err := sql.Open("mysql", DSN)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	handler := &Handler{
		TimeZones: timeZones,
		Sessions:  sessions,
		UserInfo:  users,
		DB:        db,
	}
	err = handler.OpenZip(zipFile)
	if err != nil {
		log.Fatal(err)
	}

	port := "8080"
	fmt.Println("starting server at :" + port)
	http.ListenAndServe(":"+port, handler)
}
