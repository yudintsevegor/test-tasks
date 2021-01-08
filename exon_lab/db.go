package main

import (
	"database/sql"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type DbExplorer struct {
	DB      *sql.DB
	Columns []string
	PrKey   string
}

var (
	tableName = "userdata"
)

type HttpConn struct {
	in  io.Reader
	out io.Writer
}

type Request struct {
	All    string
	ID     string
	Fields []string
}

type UserData struct {
	Uuid  string
	Login string
	Time  string
}

func (c *HttpConn) Read(p []byte) (n int, err error)  { return c.in.Read(p) }
func (c *HttpConn) Write(d []byte) (n int, err error) { return c.out.Write(d) }
func (c *HttpConn) Close() error                      { return nil }

func (h *DbExplorer) Post(in *UserData, out *string) error {
	fmt.Println("Call Post")

	db := h.DB
	var err error
	var requestToSQL = make([]string, 0, 1)
	var values = make([]interface{}, 0, 1)
	var Id interface{}

	val := reflect.ValueOf(*in)
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i).Name
		if h.PrKey == strings.ToLower(field) {
			Id = val.Field(i).Interface()
			continue
		}
		if field == "Time" && val.Field(i).Interface() != "" {
			return fmt.Errorf("You cant update a date of registration")
		}
		if field == "Time" {
			continue
		}

		req := field + "=?"
		requestToSQL = append(requestToSQL, req)
		values = append(values, val.Field(i).Interface())
	}
	requestString := strings.Join(requestToSQL, ", ")
	values = append(values, Id)

	_, err = db.Exec("UPDATE "+tableName+" SET "+requestString+" WHERE "+h.PrKey+"=?", values...)
	if err != nil {
		return err
	}
	*out = "updated"

	return nil
}

func (h *DbExplorer) Put(in *UserData, out *int64) error {
	fmt.Println("Call Put")

	db := h.DB
	var err error
	var columns = make([]string, 0, 1)
	var values = make([]interface{}, 0, 1)

	val := reflect.ValueOf(*in)
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i).Name
		if h.PrKey == strings.ToLower(field) {
			continue
		}

		if field == "Login" {
			isOk, err := checkLogin(db, val.Field(i).Interface())
			if !isOk {
				return err
			}
		}

		columns = append(columns, field)
		values = append(values, val.Field(i).Interface())
	}

	prepareColumns := strings.Join(columns, ", ")
	quesString := strings.Repeat("?, ", len(values)-1) + "?"
	statement, err := db.Prepare("INSERT INTO " + tableName + " ( " + prepareColumns + " ) " + "VALUES (" + quesString + ")")
	if err != nil {
		return err
	}
	resultFromEXEC, err := statement.Exec(values...)
	if err != nil {
		return err
	}
	insertedId, err := resultFromEXEC.LastInsertId()
	if err != nil {
		return err
	}
	*out = insertedId

	return nil
}

func checkLogin(db *sql.DB, loginFromRequest interface{}) (bool, error) {
	var count int64

	row := db.QueryRow("SELECT COUNT(*) FROM "+tableName+" WHERE login=?", loginFromRequest)
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	if count != 0 {
		return false, fmt.Errorf("Error, login: " + loginFromRequest.(string) + " exists!")
	}
	return true, nil
}

func (h *DbExplorer) Get(in *Request, out *[]interface{}) error {
	fmt.Println("Call GET")

	db := h.DB
	var numberOfColumns = 3
	var reqFields string
	var row *sql.Rows
	var err error

	switch in.All {
	case "":
		if len(in.Fields) == 0 {
			return fmt.Errorf("Error in request, no fields found")
		}
		reqFields = strings.Join(in.Fields, ", ")
		numberOfColumns = len(in.Fields)
	case "1":
		reqFields = "*"
	default:
		return fmt.Errorf("Error. Use 1 to show all fields.")
	}

	switch in.ID {
	case "":
		row, err = db.Query("SELECT " + reqFields + " FROM " + tableName)
		defer row.Close()
		if err != nil {
			return err
		}
	default:
		id, err := strconv.Atoi(in.ID)
		if err != nil {
			return err
		}
		row, err = db.Query("SELECT "+reqFields+" FROM "+tableName+" where "+h.PrKey+"=?", id)
		defer row.Close()
		if err != nil {
			return err
		}
	}

	rawResult := make([]sql.RawBytes, numberOfColumns)
	vals := make([]interface{}, numberOfColumns)

	for i, _ := range rawResult {
		vals[i] = &rawResult[i]
	}

	for row.Next() {
		err := row.Scan(vals...)
		if err != nil {
			return err
		}
		result := make(map[string]interface{})
		if reqFields == "*" {
			for i, raw := range rawResult {
				result[h.Columns[i]] = string(raw)
			}
		} else {
			for i, raw := range rawResult {
				result[in.Fields[i]] = string(raw)
			}
		}
		*out = append(*out, result)
	}
	return nil
}

func idHandler(columns []string) string {
	for _, val := range columns {
		ok := strings.Contains(val, "id")
		if ok {
			return val
		}
	}
	return ""
}

func NewDbSession(db *sql.DB) (*DbExplorer, error) {
	var idColumn string

	rows, err := db.Query("SELECT * FROM " + tableName)
	if err != nil {
		return &DbExplorer{}, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return &DbExplorer{}, err
	}
	idColumn = idHandler(columns)

	return &DbExplorer{
		DB:      db,
		Columns: columns,
		PrKey:   idColumn,
	}, nil
}
