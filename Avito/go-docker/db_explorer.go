package main

import (
	"strings"
)

func (h *Handler) getDataFromDb(name string) (bool, error) {
	db := h.DB
	rows, err := db.Query("SELECT * FROM userdata")
	if err != nil {
		return false, err
	}
	for rows.Next() {
		user := User{}
		text := ""
		err = rows.Scan(&user.Id, &user.Login, &user.Password, &text)
		if err != nil {
			return false, err
		}
		if user.Login == name {
			user.Tz = strings.Split(text, " ")
			mu.Lock()
			h.Sessions[user.Login] = ""
			h.UserInfo[user.Login] = user
			mu.Unlock()
			return true, nil
		}
	}
	return false, nil
}

func (h *Handler) putDataToDb(user User) error {
	db := h.DB
	statement, err := db.Prepare("INSERT INTO userdata" + " ( login, password, TZ ) " + "VALUES ( ?, ?, ? )")
	if err != nil {
		return err
	}
	_, err = statement.Exec(user.Login, user.Password, strings.Join(user.Tz, " "))
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) updateDataInDb(login string, in []string) error {
	db := h.DB
	mu.Lock()
	id := h.UserInfo[login].Id
	mu.Unlock()
	request := strings.Join(in, " ")
	_, err := db.Exec("UPDATE userdata SET "+"TZ=?"+" WHERE id=?", request, id)
	if err != nil {
		return err
	}
	return nil
}
