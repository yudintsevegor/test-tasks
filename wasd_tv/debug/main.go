package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Object struct {
	Index   int
	Request []string
	Percent float64
}

var (
	queue   []Object
	mu      = &sync.Mutex{}
	busy    bool
	index   int
	cmdName = "ffmpeg"
	exit    = map[string]string{
		"q":    "",
		"exit": "",
	}
	redur = regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2}.\d{2})`)
	//re = regexp.MustCompile(`\d{2}:\d{2}:\d{2}.\d{2}`)
	refps     = regexp.MustCompile(`(\d*) fps`)
	constants = []float64{60 * 60, 60, 1}
)

func main() {
	var args []string

	//	args := []string{
	//		"-i",
	//		"inputfile.avi",
	//		"output.mp4",
	//	}
	//	stat, err := os.Stat("inputfile.avi")
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	fileSize := stat.Size()
	//	fmt.Println("Size of input File: ", fileSize)

	st := Object{}

	for {
		//example: -i inputfile.avi output.mp4
		fmt.Print("Put parameters for ffmpeg: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		line := scanner.Text()
		if _, ok := exit[line]; ok {
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
		args = strings.Split(line, " ")
		index++
		st = Object{
			Index:   index,
			Request: args,
		}
		queue = append(queue, st)
		fmt.Printf("Starting convert video: \nIndex: %v,\nRequest: %v\n", st.Index, line)
		go ffmpeg()
	}
}

func ffmpeg() {
	if busy {
		return
	}
	mu.Lock()
	busy = !busy
	st := queue[0]
	mu.Unlock()

	cmd := exec.Command(cmdName, st.Request...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println()
		mu.Lock()
		busy = !busy
		mu.Unlock()
		log.Printf("\nError with request with index: %v. Error: %v\n", st.Index, err)
		return
	}
	err = cmd.Start()
	if err != nil {
		fmt.Println()
		mu.Lock()
		busy = !busy
		mu.Unlock()
		log.Printf("\nError with request with index: %v. Error: %v\n", st.Index, err)
		return
	}

	scannerFrame := bufio.NewScanner(stderr)
	scannerFrame.Split(bufio.ScanLines)

	//hardcode?
	allFrames, err := getInputFrames(scannerFrame)
	if err != nil {
		fmt.Println()
		mu.Lock()
		busy = !busy
		mu.Unlock()
		log.Print("\nInternal Error\n")
		return
	}
	fmt.Println("\nFRAMES: ", allFrames)

	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanWords)

	isFrame := false
	for scanner.Scan() {
		//		fmt.Println("NEW CHUNK")
		data := scanner.Text()
		//		fmt.Println(data)
		if strings.Contains(data, "frame=") {
			isFrame = true
			continue
		}
		if isFrame {
			currentFrames, err := strconv.ParseFloat(data, 64)
			if err != nil {
				fmt.Println()
				mu.Lock()
				busy = !busy
				mu.Unlock()
				log.Print("\nInternal Error\n")
				return
			}
			percent := currentFrames / allFrames * 100
			//			fmt.Printf("\nReady %v.", percent)
			queue[0].Percent = percent
			isFrame = false
			continue
		}
		//		getCurrentFrame()
	}
	err = cmd.Wait()
	if err != nil {
		fmt.Println()
		mu.Lock()
		busy = !busy
		mu.Unlock()
		log.Printf("\nError with request with index: %v. Error: %v\n", st.Index, err)
		return
	}

	mu.Lock()
	queue = queue[1:len(queue)]
	if len(queue) != 0 {
		busy = !busy
		go ffmpeg()
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
			//time := re.FindAllString(data, -1))[0]
			//splited := strings.Split(time, ":")])
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
