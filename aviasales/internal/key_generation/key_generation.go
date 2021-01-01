package key_generation

import (
	"fmt"
	"sort"
	"strings"
)

type LetterInfo struct {
	letter string
	counts int
}

func New(val string) string {
	if val == "" {
		return ""
	}

	elements := strings.Split(strings.ToLower(val), "")

	vocabulary := make(map[string]int, len(elements))
	for _, e := range elements {
		vocabulary[e]++
	}

	arr := make([]LetterInfo, 0, len(vocabulary))
	for letter, counts := range vocabulary {
		arr = append(arr, LetterInfo{
			letter: letter,
			counts: counts,
		})
	}

	sort.Slice(arr, func(i, j int) bool {
		return arr[i].letter < arr[j].letter
	})

	var out string
	for _, li := range arr {
		out += fmt.Sprintf("%s%d", li.letter, li.counts)
	}

	return out
}
