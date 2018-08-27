package easyterm


import (
	"golang.org/x/crypto/ssh/terminal"
	"github.com/pkg/term/termios"
	"syscall"
	"fmt"
	"strconv"
	"os"
)
/* Alias for type  */
type Termios = syscall.Termios
//type State = terminal.State

/* Global State */
var (
	terminalState = &Termios{}
	//terminalState *State
	err error
)


func Init() {
	/* Put terminal in raw mode */
	err := termios.Tcgetattr(uintptr(syscall.Stdin), terminalState)
	if err != nil {
		panic(err)
	}

	// IT'S A STRUCT!
	var tempState = &Termios{}
	*tempState = *terminalState

	tempState.Iflag &^= (syscall.IGNBRK | syscall.PARMRK | syscall.INLCR | syscall.IGNCR | syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON)
	tempState.Oflag &^= syscall.OPOST
	tempState.Cflag &^= (syscall.CSIZE | syscall.PARENB)
	tempState.Cflag |= syscall.CS8
	tempState.Lflag &^= (syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.IEXTEN | syscall.ISIG)

	err2 := termios.Tcsetattr(uintptr(syscall.Stdin), termios.TCSAFLUSH, tempState)
	//terminalState, err = terminal.MakeRaw(0)
	if err2 != nil {
		panic(err)
	}

	//terminalState, err = terminal.MakeRaw(0)
	//if err != nil {
		//panic(err)
	//}

}

func End() {
	//terminal.Restore(0, terminalState)
	termios.Tcsetattr(uintptr(syscall.Stdin), termios.TCSAFLUSH, terminalState)
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
