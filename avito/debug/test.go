package main
  
import (
    "fmt"
    "time"
)
  
func main() {
    t := time.Now()
    fmt.Println("Location : ", t.Location(), " Time : ", t) // local time
      
    location, err := time.LoadLocation("America/New_York")
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println("Location : ", location, " Time : ", t.In(location)) // America/New_York
	fmt.Println(t.In(location).Format("2006-01-02T15:04 MST"))
	fmt.Println(t.Format("2006-01-02T15:04 MST"))
}
