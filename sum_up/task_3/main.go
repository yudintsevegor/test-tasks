package main


import (
	"strings"

)

func main()  {
	
}

const (
	curly = "curly"
	parentheses = "parentheses"
	brackets = "brackets"
)

func ValidBraces(name string) bool {
	if name == ""{
		return false
	}

	input := strings.Split(name, "")
	if len(input) == 1{
		return false
	}
	counter := map[string]int{
		curly:0,
		parentheses:0,
		brackets:0,
	}


	stack := make([]string,0,1)
	for _, el := range input{
		switch el {
		case "{":
			stack = append(stack,"{")
			counter[curly]++
		case "}":
			if counter[curly] == 0{
				return false
			}
			if stack[len(stack)-1] != "{"{
				return false
			}
			stack = stack[:len(stack)-1]
			counter[curly]--
		case "(":
			stack = append(stack,"(")
			counter[parentheses]++
		case ")":
			if counter[parentheses] == 0{
				return false
			}
			if stack[len(stack)-1] != "("{
				return false
			}
			stack = stack[:len(stack)-1]
			counter[parentheses]--
		case "[":
			stack = append(stack,"[")
			counter[brackets]++
		case "]":
			if counter[brackets] == 0{
				return false
			}
			if stack[len(stack)-1] != "["{
				return false
			}
			stack = stack[:len(stack)-1]
			counter[brackets]--
		}
	}

	for _, v := range counter {
		if v != 0{
			return false
		}
	}

	return true
}