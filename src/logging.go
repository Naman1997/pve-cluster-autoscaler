package main

import (
	"log"
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
passed over to log.Printf to process and
format correctly.
*/
func ColorPrint(colorText string, text string, option ...interface{}) {
	if colorText == ERROR {
		log.Printf(color.RedString(colorText)+text, option...)
		log.Println()
		os.Exit(0)
	} else if colorText == INFO {
		log.Printf(color.GreenString(colorText))
	} else if colorText == WARN {
		log.Printf(color.YellowString(colorText))
	} else if colorText == INPUT {
		log.Printf(color.BlueString(colorText))
	} else {
		log.Printf(colorText+text, option...)
		log.Println()
		return
	}

	log.Printf(text, option...)
	if colorText != INPUT && colorText != ERROR {
		log.Println()
	}
}
