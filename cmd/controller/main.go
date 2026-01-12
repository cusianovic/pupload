package main

import (
	"fmt"
	"pupload/internal/controller"
)

func main() {
	if err := controller.Run(); err != nil {
		fmt.Printf("error running controller: %s", err)
	}
}
