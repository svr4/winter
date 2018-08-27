package easyterm


import (
	"golang.org/x/sys/unix"
	"github.com/pkg/term/termios"
	"syscall"
	"fmt"
	"strconv"
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
	err := termios.Tcgetattr(uintptr(unix.Stdin), terminalState)
	if err != nil {
		panic(err)
	}

	// IT'S A STRUCT!
	var tempState = &Termios{}
	*tempState = *terminalState

	tempState.Iflag &^= (unix.IGNBRK | unix.PARMRK | unix.INLCR | unix.IGNCR | unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON)
	tempState.Oflag &^= unix.OPOST
	tempState.Cflag &^= (unix.CSIZE | unix.PARENB)
	tempState.Cflag |= unix.CS8
	tempState.Lflag &^= (unix.ECHO | unix.ECHONL | unix.ICANON | unix.IEXTEN | unix.ISIG)

	err2 := termios.Tcsetattr(uintptr(unix.Stdin), termios.TCSAFLUSH, tempState)
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
	termios.Tcsetattr(uintptr(unix.Stdin), termios.TCSAFLUSH, terminalState)
}

func GetSize() (width, height int, err error) {
	winSize, err := unix.IoctlGetWinsize(unix.Stdout, unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}
	return int(winSize.Col), int(winSize.Row), nil
	//return terminal.GetSize(int(os.Stdout.Fd()))
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
