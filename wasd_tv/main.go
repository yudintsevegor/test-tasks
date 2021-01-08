package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Object struct {
	Index     int64
	TheLastId int
	Request   []string
	Percent   float64
}

type Info struct {
	Id      int
	Pid     int
	Request string
	Status  interface{}
}

var (
	queue   []Object
	mu      = &sync.Mutex{}
	busy    bool
	cmdName = "ffmpeg"
	exit    = map[string]string{
		"q":    "",
		"\\q":  "",
		"exit": "",
	}
	redur     = regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2}.\d{2})`)
	refps     = regexp.MustCompile(`(\d*) fps`)
	reframe   = regexp.MustCompile(`frame=(\d{0,})`)
	constants = []float64{60 * 60, 60, 1}

	table    = "queue"
	done     = "DONE"
	rejected = "REJECTED"
	//converting = "CONVERTING"

	theLastId   int
	theFirstRow bool
	pid         int
)

func main() {
	var args []string
	pid = os.Getpid()
	fmt.Println("PID: ", pid)

	db, err := sql.Open("mysql", DSN)
	if err != nil {
		log.Printf("\nInternal Error: %v\n", err)
	}

	err = db.Ping()
	if err != nil {
		log.Printf("\nInternal Error: %v\n", err)
	}

	var count int
	rowCount := db.QueryRow("SELECT COUNT(*) FROM " + table)
	err = rowCount.Scan(&count)
	if err != nil {
		log.Printf("\nInternal Error: %v\n", err)
	}
	if count == 0 {
		theFirstRow = true
	}
	for {
		//example: -i inputfile.avi output.mp4
		fmt.Print("Put parameters for ffmpeg: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		line := scanner.Text()
		if _, ok := exit[line]; ok {
			err = dbFinishProc(db)
			if err != nil {
				log.Printf("\nInternal Error: %v\n", err)
			}
			return
		}
		if line == "show" {
			for _, val := range queue {
				fmt.Printf("\nQUEUE:\nIndex: %v,\nQuery: %v,\nPercent Ready: %v.\n", val.Index, strings.Join(val.Request, " "), val.Percent)
			}
			continue
		}
		if line == "" {
			fmt.Println("Error. You dont put any parameters")
			continue
		}
		isNew := true
		args = strings.Split(line, " ")
		go ffmpeg(args, db, isNew)
	}
}

func ErrorHandler(db *sql.DB, index int64) {
	fmt.Println()
	busy = !busy
	dbUpdate(rejected, index, db)
	queue = queue[1:len(queue)]
}

func ffmpeg(args []string, db *sql.DB, isNew bool) {
	if isNew {
		info, err := dbExplore(db)
		if err != nil {
			log.Printf("\nInternal Error: %v\n", err)
			return
		}
		currentId, err := dbInsert(pid, strings.Join(args, " "), db)
		if err != nil {
			log.Printf("\nInternal Error: %v\n", err)
			return
		}

		st := Object{
			Request:   args,
			TheLastId: info.Id,
			Index:     currentId,
		}
		queue = append(queue, st)
	}

	if busy {
		return
	}

	mu.Lock()
	busy = !busy
	obj := queue[0]
	mu.Unlock()

	cmd := exec.Command(cmdName, obj.Request...)
	stderr, err := cmd.StderrPipe()

	if err != nil {
		mu.Lock()
		ErrorHandler(db, obj.Index)
		mu.Unlock()
		log.Printf("\nError with request with index: %v. Error: %v\n", obj.Index, err)
		return
	}

	for {
		// necessary to rewrite
		isOk, err := dbCheckStatus(obj.TheLastId, db)
		if err != nil {
			mu.Lock()
			ErrorHandler(db, obj.Index)
			mu.Unlock()
			log.Printf("\nInternal Error: %v\n", err)
			return
		}
		if isOk {
			break
		}
		time.Sleep(time.Second * 120)
	}

	err = cmd.Start()
	//dbUpdate(converting, obj.Index, db)
	if err != nil {
		mu.Lock()
		ErrorHandler(db, obj.Index)
		mu.Unlock()
		log.Printf("\nError with request with index: %v. Error: %v\n", obj.Index, err)
		return
	}

	scannerFrame := bufio.NewScanner(stderr)
	scannerFrame.Split(bufio.ScanLines)

	//hardcode?
	allFrames, err := getInputFrames(scannerFrame)
	if err != nil {
		mu.Lock()
		ErrorHandler(db, obj.Index)
		mu.Unlock()
		log.Printf("\nInternal Error: %v\n", err)
		return
	}
	fmt.Println("\nFRAMES: ", allFrames)

	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanWords)

	isFrame := false
	for scanner.Scan() {
		data := scanner.Text()
		if strings.Contains(data, "frame=") {
			isFrame = true
			resFrame := reframe.FindStringSubmatch(data)
			if resFrame[1] != "" {
				currentFrames, err := strconv.ParseFloat(resFrame[1], 64)
				if err != nil {
					mu.Lock()
					ErrorHandler(db, obj.Index)
					mu.Unlock()
					log.Printf("\nInternal Error: %v\n", err)
					return
				}
				percent := currentFrames / allFrames * 100
				queue[0].Percent = percent
				isFrame = false
			}
			continue
		}

		if isFrame {
			currentFrames, err := strconv.ParseFloat(data, 64)
			if err != nil {
				mu.Lock()
				ErrorHandler(db, obj.Index)
				mu.Unlock()
				log.Printf("\nInternal Error: %v\n", err)
				return
			}
			percent := currentFrames / allFrames * 100
			queue[0].Percent = percent
			isFrame = false
			continue
		}
		//		getCurrentFrame()
	}
	err = cmd.Wait()
	if err != nil {
		mu.Lock()
		ErrorHandler(db, obj.Index)
		mu.Unlock()
		log.Printf("\nError with request with index: %v. Error: %v\n", obj.Index, err)
		return
	}

	mu.Lock()
	queue = queue[1:len(queue)]
	busy = !busy
	if len(queue) != 0 {
		isNew := false
		dbUpdate(done, obj.Index, db)
		go ffmpeg(queue[0].Request, db, isNew)
	}
	mu.Unlock()
}

func getInputFrames(scanner *bufio.Scanner) (float64, error) {
	var seconds float64
	var fps float64
	var err error
	next := false
	for scanner.Scan() {
		data := scanner.Text()
		if strings.Contains(data, "Duration") {
			result := redur.FindStringSubmatch(data)
			result = result[1:]
			for ind, val := range result {
				num, err := strconv.ParseFloat(val, 64)
				if err != nil {
					return 0, err
				}
				seconds = seconds + num*constants[ind]
			}
			next = !next
			continue
		}
		if next {
			fpsStr := refps.FindStringSubmatch(data)[1]
			fps, err = strconv.ParseFloat(fpsStr, 64)
			if err != nil {
				return 0, err
			}
			break
		}
	}
	return fps * seconds, nil
}
