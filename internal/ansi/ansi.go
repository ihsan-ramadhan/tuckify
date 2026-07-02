package ansi

import "fmt"

// ANSI escape codes
const (
	Reset   = "\x1b[0m"
	Bold    = "\x1b[1m"
	Red     = "\x1b[31m"
	Green   = "\x1b[32m"
	Yellow  = "\x1b[33m"
)

// Printf-style helpers
func PrintYellow(format string, a ...interface{}) {
	fmt.Printf(Yellow+format+Reset, a...)
}

func PrintRed(format string, a ...interface{}) {
	fmt.Printf(Red+format+Reset, a...)
}

func PrintGreen(format string, a ...interface{}) {
	fmt.Printf(Green+format+Reset, a...)
}

func PrintBold(format string, a ...interface{}) {
	fmt.Printf(Bold+format+Reset, a...)
}
