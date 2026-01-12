package main

import (
	"fmt"
	"pupload/internal/worker"
)

func main() {
	if err := worker.Run(); err != nil {
		fmt.Printf("error running worker: %s", err)
	}
}
