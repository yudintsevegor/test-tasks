package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type TZ struct {
	Tz       []string
	Login string
}

type Bad struct {
	Tz       []int
	Login string
}

func main() {
	client := &http.Client{}

	tz := []string{"Europe/Moscow", "Europe/London"}
//	tz := []string{}
//	badtz := []int{}

	login := "kek"
	pass := "lol"
	st := &TZ{
		Tz: tz,
		Login: login,
	}
	data, err := json.Marshal(st)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
	reqBody := bytes.NewReader(data)
	method := http.MethodPost
	req, err := http.NewRequest(method, "http://localhost:8080/", reqBody)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(login, pass)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))
}

