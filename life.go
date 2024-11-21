package life

import (
	"fmt"
	"math/rand"
	"time"
)

type Universe [][]bool

func NewUniverse(rows int, cols int) Universe {
	uni := make(Universe, rows)
	for i := 0; i < rows; i++ { 
		uni[i] = make([]bool, cols)
	}
	return uni
}


func PrintUniverse(uni Universe, w int, h int) {
	for row := 0; row < w; row++{
	for col := 0; col < h; col++{
		if uni[row][col] {
			fmt.Print("*")
		}else{
			fmt.Print(" ")
		}
	}
	fmt.Print("\n")
	}
}

func Seed(uni Universe, rows int, cols int){
	seeds := rows*cols/ 4

	for {
		// gen random row and col
		row_r := rand.Intn(rows)
		col_r := rand.Intn(cols)

		//if not already alive, set live

		if uni[row_r][col_r] == false {
			uni[row_r][col_r] = true
			// break condition
			seeds--
			if seeds <= 0 {
				break
			}
		}
	}

}

func (uni Universe) Alive(x, y int) bool {
	x = (x + len(uni[0])) % len(uni[0])
	y = (y + len(uni)) % len(uni)

	return uni[y][x]
}

func (uni Universe) Neighbors(x, y int) int {
	n := 0
	if uni.Alive(x,y+1) {
		n++
	}
	if uni.Alive(x,y-1) {
		n++
	}
	if uni.Alive(x-1,y) {
		n++
	}
	if uni.Alive(x+1,y) {
		n++
	}
	if uni.Alive(x+1,y+1) {
		n++
	}
	if uni.Alive(x+1,y-1) {
		n++
	}
	if uni.Alive(x-1,y-1){
		n++
	}
	if uni.Alive(x-1,y+1){
		n++
	}

	return n
}

func (uni Universe) Next(x,y int) bool {
	alive := uni.Alive(x,y)
	neighbors := uni.Neighbors(x,y)
	live := false

	if alive {
		if neighbors < 2 {
			live = false
		}else if neighbors > 3{
			live = false
		}else {
			live=true
		}
	}

	if neighbors == 3 {
		live = true
	}

	return live
}

func main() {

	const rows = 30
	const cols = 60

	a := NewUniverse(rows, cols)
	b := NewUniverse(rows, cols)

	Seed(a, rows, cols)

	for {
		for row := 0; row < rows; row++{
		for col := 0; col< cols; col++{
			b[row][col] = a.Next(row, col)
		}
		}

		a,b = b,a
		fmt.Print("\033[H\033[2J")
		PrintUniverse(a, rows, cols)
		time.Sleep(350 * time.Millisecond)
	}

}