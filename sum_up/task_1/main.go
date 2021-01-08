package main

func main(){

}
func Solution(numbers int) int {
	var result int
	for i := 0; i < numbers; i++ {
		if i%3 == 0 {
			result += i
		} else if i%5 == 0 {
			result += i
		}
	}

	return result
}