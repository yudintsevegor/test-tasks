package main

import (
"sort"
"strconv"
"strings"
)


func main(){

}

func OrderByWeight(strn string) string {
	if strn == "" {
		return strn
	}

	var (
		input      = strings.Split(strn, " ")
		weightInfo = make(map[int][]string)
	)

	for _, s := range input {
		splittedNumbers := strings.Split(s, "")
		var result int
		for _, number := range splittedNumbers {
			n, _ := strconv.Atoi(number)
			result += n
		}

		if _, ok := weightInfo[result]; !ok {
			weightInfo[result] = []string{s}
			continue
		}

		weightInfo[result] = append(weightInfo[result], s)
	}

	type weightInfos struct {
		numbers []string
		weight  int
	}

	weights := make([]weightInfos, 0, len(weightInfo))
	for weight, numbers := range weightInfo {
		sort.Strings(numbers)
		weights = append(weights, weightInfos{
			numbers: numbers,
			weight:  weight,
		})
	}

	sort.Slice(weights, func(i, j int) bool {
		return weights[i].weight < weights[j].weight
	})

	output := make([]string, 0, len(weights))
	for _, v := range weights {
		output = append(output, v.numbers...)
	}

	return strings.Join(output, " ")
}