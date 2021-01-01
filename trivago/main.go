package main

import "log"

func main() {
	if err := convert("hotels.csv", ',', []FormatTag{xmlTag, jsonTag}, sortByNames); err != nil {
		log.Println(err)
		return
	}
}
