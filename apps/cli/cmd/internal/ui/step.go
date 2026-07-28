package ui

import "fmt"

func Step(number int, total int, message string) {

	fmt.Printf(
		"\n[%d/%d] %s\n",
		number,
		total,
		message,
	)
}