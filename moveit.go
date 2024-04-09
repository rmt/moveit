package main

import (
	"fmt"
	"os"

	"github.com/rmt/moveit/desktop"
)

func main() {
	desk := desktop.NewDesktop()
	activeWindow := desk.GetActiveWindow()
	fmt.Printf("Active Window: 0x%x\n", activeWindow)
	fmt.Printf("Head Geometry:\n")
	desk.PrintHeadGeometry()
	printHelp := false
	if len(os.Args) > 2 {
		cmd := os.Args[1]
		placement := os.Args[2]
		if cmd == "move" {
			desk.SmartPlacement(activeWindow, placement, 50)
		} else if cmd == "focus" {
			desk.SmartFocus(activeWindow, placement, 50)
		} else {
			printHelp = true
		}
	} else {
		printHelp = true
	}
	if printHelp {
		fmt.Printf("Syntax: %s move|focus B?[NS][EW]|C\n", os.Args[0])
		fmt.Println("eg. move NE, move BSE (big south-east)")
	}
	fmt.Println("")
}
