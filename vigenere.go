package cipher

import "fmt"

func main() {

	cipher := "CSOITEUIWUIZNSROCKNKFD"
	key := "GOLANG"

	for i := 0; i < len(cipher); i++ {
		if cipher[i] == ' '{
			fmt.Print(" ")
		}else {
			shiftby := key[i%len(key)] - 'A'
			c := (cipher[i] - 'A' + shiftby) % 26
			fmt.Printf("%c",c + 'A')
		}
	}

}