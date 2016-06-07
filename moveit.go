package main

import (
	"fmt"
	"os"

	"./desktop"
)

func main() {
	desk := desktop.NewDesktop()
	activeWindow := desk.ActiveWindow()
	fmt.Printf("Active Window: %d\n", activeWindow)
	fmt.Printf("Head Geometry:\n")
	desk.PrintHeadGeometry()
	if len(os.Args) > 1 {
		command := os.Args[1]
		if command == "list" {
			fmt.Printf("---\n")
			desk.PrintWindowsOnCurrentDesktop()
		} else {
			desk.SmartPlacement(activeWindow, command, 60)
		}
	}
}
