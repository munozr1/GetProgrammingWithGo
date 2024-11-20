package tickets

import (
	"fmt"
	"math/rand"
)

func main(){




	fmt.Printf("%-20v\t%5v\t%10v\t%-5v\n", "Spaceline", "Days", "Trip   ", "Price")
	fmt.Println("======================================================")



	for i := 0; i < 10; i++ {
		const distance int64 = 62100000 // distance to mars
		var speed = int64((rand.Intn(16) + 16)) // km/s
		var time_s = distance / speed //seconds
		var days = time_s / 86400
		var trip_type = rand.Intn(2)

		switch rand.Intn(3){
		case 0: // round trip
			price := rand.Intn(14) + 36
			if trip_type == 0{
				fmt.Printf("%-20v\t%5v\t%-10v\t$%5v\n", "Virgin Galactic", days, "Round-Trip", 2 * price)
			}else{
				fmt.Printf("%-20v\t%5v\t%-10v\t$%5v\n", "Virgin Galactic", days, "One-way", price)
			}
		case 1:
			price := rand.Intn(14) + 36
			if trip_type == 0{
				fmt.Printf("%-20v\t%5v\t%-10v\t$%5v\n", "SpaceX", days, "Round-Trip", 2 * price)
			}else{
				fmt.Printf("%-20v\t%5v\t%-10v\t$%5v\n", "SpaceX", days, "One-way", price)
			}
		default:
				price := rand.Intn(14) + 36
			if trip_type == 0{
				fmt.Printf("%-20v\t%5v\t%-10v\t$%5v\n", "SpaceX", days, "Round-Trip", 2 * price)
			}else{
				fmt.Printf("%-20v\t%5v\t%-10v\t$%5v\n", "Space Adventures", days, "One-way", price)
			}
		}
	}

}