package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

/*
ERROR const => red
WARN const => yellow
INPUT const => blue
INFO const => green
*/
const (
	ERROR = "[ERROR] "
	WARN  = "[WARNING] "
	INPUT = "[INPUT] "
	INFO  = "[INFO] "
)

/*
ColorPrint formats and prints the text according
to the colorText provided. Other options are
passed over to fmt.Printf to process and
format correctly.
*/
func ColorPrint(colorText string, text string, option ...interface{}) {
	if colorText == ERROR {
		fmt.Printf(color.RedString(colorText)+text, option...)
		fmt.Println()
		os.Exit(0)
	} else if colorText == INFO {
		fmt.Printf(color.GreenString(colorText))
	} else if colorText == WARN {
		fmt.Printf(color.YellowString(colorText))
	} else if colorText == INPUT {
		fmt.Printf(color.BlueString(colorText))
	} else {
		fmt.Printf(colorText+text, option...)
		fmt.Println()
		return
	}

	fmt.Printf(text, option...)
	if colorText != INPUT && colorText != ERROR {
		fmt.Println()
	}
}
