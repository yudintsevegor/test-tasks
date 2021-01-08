package main

import (
	"archive/zip"
	"encoding/json"
	"net/http"
	"time"
	//	"fmt"
)

func SendData(w http.ResponseWriter, status int, out []LT) {
	w.WriteHeader(status)
	w.Header().Set("Content-type", "application/json")
	s := FromServer{Out: out}
	jsonOut, _ := json.Marshal(s)
	w.Write(jsonOut)
}

func SendError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	w.Header().Set("Content-type", "application/json")
	jsonError, _ := json.Marshal(FromServer{Error: err.Error()})
	w.Write(jsonError)
}

func (h *Handler) getTz(in []string) []LT {
	out := make([]LT, 0, 1)
	for _, v := range in {
		//		fmt.Println(v)
		if loc, ok := h.TimeZones[v]; ok {
			s := LT{Location: v, Time: time.Now().In(loc).Format("2006-01-02T15:04 MST")}
			out = append(out, s)
			continue
		}
	}
	return out
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
		st.TimeZones[file.Name] = location
	}

	return nil
}
