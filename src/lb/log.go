package lb

import (
	"log"
)

const (
	GREEN   = "\x1b[32m"
	YELLOW  = "\x1b[33m"
	MAGENTA = "\x1b[35m"
	RED     = "\x1b[31m"
	CYAN    = "\x1b[36m"
	RESET   = "\x1b[0m"
)

// TODO check if outputting to terminal for colour output
func logColour(msg, colour string) {
	log.Printf("%s%s%s", colour, msg, RESET)
}

func logGreen(msg string) {
	logColour(msg, GREEN)
}

func logRed(msg string) {
	logColour(msg, RED)
}

func logYellow(msg string) {
	logColour(msg, YELLOW)
}
