package main

import (
	"log"
	"strconv"
	"strings"
)

func main() {
	/*
		query := []string{"insert",
			"addToValue",
			"get",
			"insert",
			"addToKey",
			"addToValue",
			"get"}

		vals := [][]int{[]int{1, 2}, []int{2}, []int{1}, []int{2, 3}, []int{1}, []int{-1}, []int{3}}
	*/

	query := []string{"insert",
		"insert",
		"addToValue",
		"addToKey",
		"get"}
	vals := [][]int{[]int{1, 2}, []int{2, 3}, []int{2}, []int{1}, []int{3}}

	log.Println(hashMap(query, vals))
}

const (
	insert     = "insert"
	addToValue = "addToValue"
	addToKey   = "addToKey"
	get        = "get"
)

func hashMap(queryType []string, query [][]int) int64 {
	if len(queryType) != len(query) {
		return 0
	}

	var (
		sum     int64
		hashMap = make(map[int]int, len(query))
	)

	for i := 0; i < len(queryType); i++ {
		currentQuery := query[i]
		switch queryType[i] {
		case insert:
			hashMap[currentQuery[0]] = currentQuery[1]
		case addToKey:
			newHash := make(map[int]int, len(hashMap))
			for k, v := range hashMap {
				newHash[k+currentQuery[0]] = v
			}
			hashMap = newHash
		case addToValue:
			for k, v := range hashMap {
				hashMap[k] = v + currentQuery[0]
			}
		case get:
			sum += int64(hashMap[currentQuery[0]])
		default:
			return 0
		}

		log.Print(hashMap, queryType[i])
	}

	return sum
}

func digitsManipulations(n int) int {
	res := strings.Split(strconv.Itoa(n), "")

	intSlice := make([]int, 0, len(res))
	for _, v := range res {
		num, err := strconv.Atoi(v)
		if err != nil {
			return 0
		}
		intSlice = append(intSlice, num)
	}

	var (
		product = 1
		sum     = 0
	)

	for _, v := range intSlice {
		product *= v
		sum += v
	}

	// log.Print(product)
	// log.Print(sum)

	return product - sum
}
