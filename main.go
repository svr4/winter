package main

import (
	"easyterm"
	"bufio"
	"os"
	"fmt"
	"math"
	//"strings"
)
/* Alias ReadWriter */
type ReadWriter = bufio.ReadWriter


type Cursor struct {
	x int
	y int
}

/* Contains the state of the editor session */
type WinterState struct {
	cursorPos Cursor
	currentLine *BufferNode
}

/* Global State */
var myState WinterState

/* stdin/stdout RW */
var termRW *ReadWriter

func clearBuffer(buff []byte) {
	buff[0] = 0
	buff[1] = 1
	buff[2] = 2
	buff[3] = 3
}

func updateCursorPosX(num int) {
	if (myState.cursorPos.x + num) > 0 {
		myState.cursorPos.x += num
	}
}

func updateCursorPosY(num int) {
	if (myState.cursorPos.y + num) > 0 {
		myState.cursorPos.y += num
	}
}

func setCursorPos(y, x int) {
	myState.cursorPos.x = x
	myState.cursorPos.y = y
}

func moveCursorX(num int, fn func(int)) {
	line_length := myState.currentLine.length
	// Current pos + the next column < the length of the line
	if ((myState.cursorPos.x - 1) + num) <= line_length {
		fn(int(math.Abs(float64(num)))) // Calls the easyterm func with a positive number
		updateCursorPosX(num)
	}
	showEditorData()
}

func moveCursorY(num int, fn func(int)) {
	buffer_length := sbGetBufferLength()
	var (
		currentLine *BufferNode
		nextLine *BufferNode
	)
	if ((myState.cursorPos.y - 1) + num) < buffer_length {
		currentLine = sbGetLine(myState.cursorPos.y)
		//updateCursorPosY(num)
		nextLine = sbGetLine(myState.cursorPos.y + num)

		if nextLine == nil {
			return
		}

		fn(int(math.Abs(float64(num)))) // Calls the easyterm func with a positive number
		updateCursorPosY(num)

		if nextLine.length < currentLine.length {
			//easyterm.CursorPos(0,1)
			easyterm.CursorPos(myState.cursorPos.y, nextLine.length)
			setCursorPos(myState.cursorPos.y, nextLine.length + 1)
		}

		myState.currentLine = nextLine
		showEditorData()
	}
}

func showEditorData() {
	easyterm.CursorPos(40, 1)
	fmt.Printf("Current Line Index: %v", myState.currentLine.index)
	fmt.Print("\n")
	easyterm.CursorPos(41, 1)
	fmt.Print("Position:")
	fmt.Print("\n")
	easyterm.CursorPos(42, 1)
	fmt.Printf("X: %v Y: %v", myState.cursorPos.x, myState.cursorPos.y)
	fmt.Print("\n")
	easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
}

func writeTextToBuffer(letter byte) {

	switch {
	// begining
	case myState.cursorPos.x == 1:
		myState.currentLine.line = string(letter) + myState.currentLine.line
	// end
	case myState.cursorPos.x == (myState.currentLine.length + 1):
		myState.currentLine.line += string(letter)
	// middle
	case myState.cursorPos.x < (myState.currentLine.length + 1):
		total := make([]rune, myState.currentLine.length + 1)
		firstHalf := make([]rune, len(myState.currentLine.line[0:myState.cursorPos.x - 1]))
		copy(firstHalf, []rune(myState.currentLine.line[0:myState.cursorPos.x - 1]))
		secondHalf := make([]rune, len(myState.currentLine.line[myState.cursorPos.x-1:myState.currentLine.length]))
		copy(secondHalf, []rune(myState.currentLine.line[myState.cursorPos.x-1:myState.currentLine.length]))

		copy(total, []rune(string(firstHalf) + string(letter) + string(secondHalf)))
		myState.currentLine.line = string(total)
	}
	easyterm.ClearLine()
	easyterm.CursorPos(myState.cursorPos.y, 1)
	fmt.Print(myState.currentLine.line)
	easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x + 1)
	myState.currentLine.length = len(myState.currentLine.line)
}

func backspaceLine() {

	if myState.currentLine.length == 0 {
		// delete node, rehook list ends and move cursor to end of previous line
		var (
			prev *BufferNode
			next *BufferNode
		)
		prev = myState.currentLine.prev
		next = myState.currentLine.next

		if prev != nil {
			prev.next = next
		} else {
			return
		}

		if next != nil {
			next.prev = prev
		}

		myState.currentLine.prev = nil
		myState.currentLine.next = nil

		myState.currentLine = prev

		updateCursorPosY(-1) // move up one line in state
		sbUpdateBufferIndexes()
		sbReprintBuffer() // reprints complete buffer
		easyterm.CursorPos(myState.cursorPos.y, myState.currentLine.length + 1)
		setCursorPos(myState.cursorPos.y, myState.currentLine.length + 1)
		showEditorData()
		return
	}


	// TODO: single line wrapping handling

	switch {
		// Between the first and last characters of a line
	case myState.cursorPos.x > 1 && myState.cursorPos.x < (myState.currentLine.length + 1):

		total := make([]rune, myState.currentLine.length - 1)
		firstHalf := make([]rune, len(myState.currentLine.line[0:myState.cursorPos.x-2]))
		copy(firstHalf, []rune(myState.currentLine.line[0:myState.cursorPos.x-2]))
		secondHalf := make([]rune, len(myState.currentLine.line[myState.cursorPos.x-1:myState.currentLine.length]))
		copy(secondHalf, []rune(myState.currentLine.line[myState.cursorPos.x-1:myState.currentLine.length]))

		copy(total, []rune(string(firstHalf) + string(secondHalf)))
		myState.currentLine.line = string(total)
		myState.currentLine.length = len(total)

		// reprint line
		easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
		easyterm.ClearLine()
		easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
		fmt.Print(myState.currentLine.line) // write updated line again
		easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
		// move & update cursor
		easyterm.CursorLeft(1)
		updateCursorPosX(-1)

	case myState.cursorPos.x == 1:

		// check if we have line on top, if not do nothing, else move it up
		var prev *BufferNode
		prev = myState.currentLine.prev
		if prev != nil {
			// move current line up
			origPrevLineLength := prev.length // to move the cursor later
			currentText := make([]rune, myState.currentLine.length)
			copy(currentText, []rune(myState.currentLine.line))
			prev.line += string(currentText)
			prev.length = len(prev.line)

			// rehook list, sans the soon to be destroyed node and update currentLine state
			var next *BufferNode
			next = myState.currentLine.next
			if next != nil {
				prev.next = next
				next.prev = prev
				myState.currentLine.prev = nil
				myState.currentLine.next = nil
			} else {
				prev.next = nil
				myState.currentLine.prev = nil
			}

			myState.currentLine = prev
			updateCursorPosY(-1) // move up one line in state
			/*easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start of new line
			easyterm.ClearLine()
			easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
			fmt.Print(myState.currentLine.line) // write updated line again*/
			sbUpdateBufferIndexes()
			sbReprintBuffer() // reprints complete buffer
			easyterm.CursorPos(myState.cursorPos.y, origPrevLineLength+1)
			setCursorPos(myState.cursorPos.y, origPrevLineLength + 1)
			//easyterm.CursorPos(20, 1)
			//fmt.Print(sbGetBufferLength())

		}

	case myState.cursorPos.x == (myState.currentLine.length + 1):

		newLine := make([]rune, myState.currentLine.length-1)
		copy(newLine, []rune(myState.currentLine.line[0:myState.currentLine.length-1]))
		myState.currentLine.line = string(newLine)
		myState.currentLine.length = len(myState.currentLine.line)

		// reprint line
		easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
		easyterm.ClearLine()
		easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
		fmt.Print(myState.currentLine.line) // write updated line again
		easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
		// move & update cursor
		easyterm.CursorLeft(1)
		updateCursorPosX(-1)
	}
	showEditorData()
}

func main() {
	/* Terminal in raw mode */
	easyterm.Init()
	easyterm.Clear()
	easyterm.CursorPos(1,1)

	/* Init my position */
	myState.cursorPos = Cursor{1,1}

	/* Reader and Writer to standard in & out */
	termRW = bufio.NewReadWriter(bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout))

	/* Vars for holding data */
	var (
		buffer []byte
	)

	/* Init those vars */
	buffer = make([]byte, 4) // Length of an int var

	/* TODO:  Parameter handling for either loading a file or creating a new one */
	/* DOING: Working on new file scenario */

	sbInitNewFile()
	sbPrintBuffer()
	/* Get first line node to have something to write to */
	myState.currentLine = sbGetLine(1)
	showEditorData()

	for {
		if bytesRead, err := termRW.Reader.Read(buffer); err == nil {
			//fmt.Printf("%s", string(letter))
			/* Means that the arrow keys where pressed */
			/* This will send 3 bytes: Esc, [ and (A or B or C or D)  */
			if bytesRead > 1 {
				if buffer[0] == 27 && buffer[1] == 91 {

					switch buffer[2] {

					case 68:
						// Left arrow
						moveCursorX(-1, easyterm.CursorLeft)
						//easyterm.CursorLeft(1)
						//updateCursorPosX(-1)

					case 65:
						// Up arrow
						moveCursorY(-1, easyterm.CursorUp)
						//easyterm.CursorUp(1)
						//updateCursorPosY(1)

					case 67:
						// Right arrow
						moveCursorX(1, easyterm.CursorRight)
						//easyterm.CursorRight(1)
						//updateCursorPosX(1)

					case 66:
						// Down arrow
						moveCursorY(1, easyterm.CursorDown)
						//easyterm.CursorDown(1)
						//updateCursorPosY(-1)

					}
					clearBuffer(buffer)
				} // If not [ ignore the Esc

			} else {

				letter := buffer[0]

				switch {
					case letter == 13:
						// Enter
						//easyterm.CursorNextLine(1)
						//updateCursorPosY(1)
						sbAddLineToBuffer(myState.cursorPos.y, myState.cursorPos.x)
						updateCursorPosY(1)
						easyterm.CursorPos(myState.cursorPos.y, 1)
						setCursorPos(myState.cursorPos.y, 1)
						myState.currentLine = sbGetLine(myState.cursorPos.y)
						showEditorData()
						//fmt.Println(sbGetBufferLength())
						//myState.currentLine = sbGetLine(myState.cursorPos.y)
						//easyterm.CursorPos(myState.cursorPos.y, 1)
						//myState.currentLine = sbGetLine(myState.cursorPos.y)
						//fmt.Print(myState.cursorPos.x)
						//fmt.Print(myState.cursorPos.y)
					case letter == 127:
						// Backspace
						//easyterm.CursorLeft(1)
						//updateCursorPosX(-1)
						//easyterm.ClearFromCursor()
						backspaceLine()
				  case letter == 27:
						// Do Nothing for Esc for now

					case letter == 19:
						// Save

					case letter == 17:
						// Ctrl-Q
						easyterm.Clear()
						easyterm.CursorPos(1,1)
						easyterm.End()
						return
					case letter > 0 && letter <= 31:
						// Do nothing

					default:
						//fmt.Print(letter)
						///fmt.Printf("BytesRead: %d\n", bytesRead)
						//fmt.Printf("%s", string(letter))
						//updateCursorPosX(1)
						writeTextToBuffer(letter)
						updateCursorPosX(1)
				}

			}

		} else {
			easyterm.Clear()
			easyterm.CursorPos(1,1)
			easyterm.End()
			return
		}
	}

}
