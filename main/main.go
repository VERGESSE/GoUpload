package main

import (
	"github.com/go-vgo/robotgo"
	"math/rand"
	"time"
)

func main() {

	for {
		time.Sleep(time.Second * 6)
		robotgo.ScrollMouse(rand.Intn(10), "down")
		robotgo.MoveMouseSmooth(rand.Intn(1000)+400, rand.Intn(800)+200)
		robotgo.MouseClick("left", true)
		time.Sleep(time.Second * 2)
		robotgo.ScrollMouse(rand.Intn(10), "up")
		robotgo.MoveMouseSmooth(rand.Intn(1000)+400, rand.Intn(800)+200)
		robotgo.MouseClick("left", true)
	}
}
