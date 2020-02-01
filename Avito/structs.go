package main

import (
	"database/sql"
	"time"
)

type FromServer struct {
	Error string `json:"error"`
	Out   []LT   `json:"out"`
}

type FromRequest struct {
	Tz    []string `json:"tz"`
	Login string   `json:"login"`
}

type Handler struct {
	TimeZones map[string]*time.Location
	Sessions  map[string]string
	UserInfo  map[string]User
	DB        *sql.DB
}

type User struct {
	Id       string
	Login    string
	Password string
	Tz       []string
}

type LT struct {
	Location string
	Time     string
}
