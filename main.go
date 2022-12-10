package main

import (
	"flag"
	"sync"
)

func main() {
	fptr := flag.Int("id", 0, "请输入id")
	flag.Parse()

	s := NewServer(*fptr)

	s.StartServer()

	sync := sync.WaitGroup{}
	sync.Add(1)

	sync.Wait()

}
