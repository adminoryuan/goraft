package main

import (
	"fmt"
	"math/rand"
	"time"
)

func SleepRandomTime() {
	rand.Seed(time.Now().UnixMilli())
	m := rand.Intn(155) + 150
	fmt.Println(m)
	time.Sleep(time.Microsecond * time.Duration(m))
}
