package logger

import "fmt"

//Logger struct
type Logger struct {
	quiet bool
}

// Print print its arguments
func (l *Logger) Print(args ...interface{}) {
	if l.quiet == false {
		fmt.Println(args...)
	}
}

//New start a new
func New(quiet bool) Logger {
	return Logger{
		quiet: quiet,
	}
}
