package main

import (
	"database/sql"
)

func dbUpdate(status string, id int64, db *sql.DB) (error){
	_, err := db.Exec("UPDATE " + table + " SET status = '" + status + "' WHERE uuid=?;", id)
	if err != nil {
		return err
	}
	return nil
}

func dbInsert(pid int, request string, db *sql.DB) (int64, error){
	valuesToDb := make([]interface{}, 0, 1)
	valuesToDb = append(valuesToDb, pid)
	valuesToDb = append(valuesToDb, request)
	
	statement, err := db.Prepare("INSERT INTO " + table + " (process_id, request) VALUES (?, ?) ;")
	if err != nil {
		return 0, err
	}
	resultFromExec, err := statement.Exec(valuesToDb...)
	if err != nil {
		return 0, err
	}
	insertedId, err := resultFromExec.LastInsertId()
	if err != nil{
		return 0, err
	}
	
	theFirstRow = false
	
	return insertedId, nil
}

func dbFinishProc(db *sql.DB) (error){
	_, err := db.Exec("UPDATE " + table + " SET status = 'REJECTED' WHERE process_id=? AND status IS NULL;", pid)
	if err != nil {
		return err
	}
	return nil
}

func dbExplore(db *sql.DB) (Info, error){
	if theFirstRow {
		return Info{
			Status: "EMPTY",
		}, nil
	}
	info := Info{}
	row := db.QueryRow("SELECT * FROM " + table + " ORDER BY uuid DESC LIMIT 1;")
	err := row.Scan(&info.Id, &info.Pid, &info.Request, &info.Status)
	if err != nil {
		return Info{}, err
	}
	return info, nil
}

func dbCheckStatus(theLastId int, db *sql.DB) (bool, error){
	var status interface{}
	
	row, err := db.Query("SELECT status FROM " + table + " WHERE uuid<=?;", theLastId)
	defer row.Close()
	
	if err != nil{
		return false, err
	}
	
	for row.Next() {
		err := row.Scan(&status)
		if err != nil {
			return false, nil
		}
		if status == nil {
			return false, nil
		}
	}
	
	return true, nil
}

