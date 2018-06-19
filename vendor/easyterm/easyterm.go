package easyterm


import (
	"golang.org/x/crypto/ssh/terminal"
	"fmt"
	"strconv"
	"os"
)
/* Alias for type  */
type State = terminal.State

/* Global State */
var (
	terminalState *State
	err error
)


func Init() {
	/* Put terminal in raw mode */
	terminalState, err = terminal.MakeRaw(0)
	if err != nil {
		panic(err)
	}

}

func End() {
	terminal.Restore(0, terminalState)
}

func GetSize() (width, height int, err error) {
	return terminal.GetSize(int(os.Stdout.Fd()))
}

func CursorUp(rows int) {
	fmt.Print("\033[" + strconv.Itoa(rows) + "A")
}

func CursorDown(rows int) {
	fmt.Print("\033[" + strconv.Itoa(rows) + "B")
}

func CursorRight(cols int) {
	fmt.Print("\033[" + strconv.Itoa(cols) + "C")
}

func CursorLeft(cols int) {
	fmt.Print("\033[" + strconv.Itoa(cols) + "D")
}

func CursorNextLine(lines int) {
	fmt.Print("\033[" + strconv.Itoa(lines) + "E")
}

func CursorPreviousLine(lines int) {
	fmt.Print("\033[" + strconv.Itoa(lines) + "F")
}

func CursorPos(y, x int) {
	fmt.Print("\033["+ strconv.Itoa(y) + ";" + strconv.Itoa(x) + "H")
}

func Clear() {	
	fmt.Print("\033[2J")
}

func ScrollUp(lines int) {
	fmt.Print("\033[" + strconv.Itoa(lines) + "S")
}

func ScrollDown(lines int) {
	fmt.Print("\033[" + strconv.Itoa(lines) + "T")
}

func ClearFromCursor() {
	fmt.Print("\033[0K")
}

func ClearLine() {
	fmt.Print("\033[2K")
}
