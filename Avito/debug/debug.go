package main

import (
	"archive/zip"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var mu = &sync.Mutex{}

type User struct {
	Login    string
	Password string
}

type Handler struct {
	TimeZones map[string]time.Time
	Sessions  map[string]User
	Slice     []string
}

type FromRequest struct {
	Tz []string `json:"tz"`
}

type LT struct {
	Location string
	Time     time.Time
}

type FromServer struct {
	Error string `json:"error"`
	Out   []LT   `json:"out"`
}

func (h *Handler) handleMain(w http.ResponseWriter, r *http.Request) {
	Sessions := h.Sessions
	cookie, err := r.Cookie("session")
	if err == nil {
		if _, ok := Sessions[cookie.Value]; ok {
			http.Redirect(w, r, "/tz", http.StatusFound)
			return
		}
	}
	tmpl := template.Must(template.ParseFiles("checkbox.html"))
	tmpl.Execute(w, struct {
		Locations []string
	}{
		h.Slice,
	})
}

func getRandomString() string {
	size := 16
	rb := make([]byte, size)
	_, err := rand.Read(rb)
	if err != nil {
		log.Fatal(err)
	}
	oauthString := base64.URLEncoding.EncodeToString(rb)

	return oauthString
}

func (h *Handler) handleResult(w http.ResponseWriter, r *http.Request) {
	length := len(h.Slice)
	tz := make([]string, 0, 1)
	for i := 0; i < length; i++ {
		j := strconv.Itoa(i)
		val := r.FormValue(j)
		if val != "" {
			tz = append(tz, val)
		}
	}
	st := &FromRequest{
		Tz: tz,
	}
	data, err := json.Marshal(st)
	if err != nil {
		log.Fatal(err)
	}
	reqBody := bytes.NewReader(data)
}

func (h *Handler) handleGetTZ(w http.ResponseWriter, r *http.Request) {
	timeZones := h.TimeZones

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-type", "application/json")
		jsonError, _ := json.Marshal(FromServer{Error: "InternalServerError"})
		w.Write(jsonError)
		return
	}

	st := &FromRequest{}
	err = json.Unmarshal(body, &st)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-type", "application/json")
		jsonError, _ := json.Marshal(FromServer{Error: "BadRequest"})
		w.Write(jsonError)
		return
	}

	out := make([]LT, 0, 1)
	//	notMatched := make([]string, 0, 1)
	for _, v := range st.Tz {
		if t, ok := timeZones[v]; ok {
			fmt.Println(v, t)
			s := LT{Location: v, Time: t}
			out = append(out, s)
			continue
		}
		//		notMatched = append(notMatched, v)
	}

	if len(out) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Header().Set("Content-type", "application/json")
		jsonError, _ := json.Marshal(FromServer{Error: "NoContent"})
		w.Write(jsonError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")
	s := FromServer{Out: out}
	jsonOut, _ := json.Marshal(s)
	w.Write(jsonOut)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.handleMain(w, r)
	case "/result":
		h.handleResult(w, r)
	case "/tz":
		h.handleGetTZ(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	//Copied from GOROOT/go/...
	zipFile := "zoneinfo.zip"
	var timeZones = make(map[string]time.Time)
	var sessions = make(map[string]User)
	var slice = make([]string, 0, 1)

	handler := &Handler{
		TimeZones: timeZones,
		Sessions:  sessions,
		Slice:     slice,
	}
	err := handler.OpenZip(zipFile)
	if err != nil {
		log.Fatal(err)
	}

	//	for key, val := range timeZones{
	//		fmt.Printf("TZ: %v\nTime: %v\n", key, val)
	//	}
	//

	port := "8080"
	fmt.Println("starting server at :" + port)
	http.ListenAndServe(":"+port, handler)
}

func (st *Handler) OpenZip(zipFile string) error {

	zf, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}

	defer zf.Close()
	for _, file := range zf.File {
		info := file.FileInfo()
		if info.IsDir() {
			continue
		}
		location, err := time.LoadLocation(file.Name)
		if err != nil {
			return err
		}

		time := time.Now().In(location)
		st.TimeZones[file.Name] = time
		st.Slice = append(st.Slice, file.Name)
	}

	return nil
}
