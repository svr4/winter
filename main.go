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
	if (myState.cursorPos.x + num) >= 0 {
		myState.cursorPos.x += num
	}
}

func updateCursorPosY(num int) {
	if (myState.cursorPos.y + num) >= 0 {
		myState.cursorPos.y += num
	}
}

func moveCursorX(num int, fn func(int)) {
	line_length := myState.currentLine.length
	// Current pos + the next column < the length of the line
	if (myState.cursorPos.x + num) <= line_length {
		fn(int(math.Abs(float64(num)))) // Calls the easyterm func with a positive number
		updateCursorPosX(num)
	}
}

func moveCursorY(num int, fn func(int)) {
	buffer_length := sbGetBufferLength()
	var (
		currentLine *BufferNode
		nextLine *BufferNode
	)
	if (myState.cursorPos.y + num) < (buffer_length - 1) {
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
		}

		myState.currentLine = nextLine
	}
}

func writeTextToBuffer(letter byte) {
	myState.currentLine.line += string(letter)
	myState.currentLine.length = len(myState.currentLine.line)
}

func backspaceLine() {

	if myState.currentLine.length == 0 {
		return
	}

	// TODO: move line up if length == 0

	easyterm.CursorLeft(1)
	updateCursorPosX(-1)
	//var tempPos := Cursor{myState.cursorPos.x, myState.cursorPos.y}
	lineRune := []rune(myState.currentLine.line) // Current text in the line
	if myState.cursorPos.x > 0 {
		myState.currentLine.line = string(lineRune[0:myState.cursorPos.x]) + string(lineRune[(myState.cursorPos.x + 1):myState.currentLine.length])
		myState.currentLine.length = len(myState.currentLine.line)
		easyterm.CursorPos(myState.cursorPos.y, 0) // move cursor to start
		easyterm.ClearLine()
		easyterm.CursorPos(myState.cursorPos.y, 0) // move cursor to start
		fmt.Print(myState.currentLine.line) // write updated line again
		easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x + 1)
	} else if myState.cursorPos.x == 0{
		myState.currentLine.line = string(lineRune[1:len(myState.currentLine.line)])
		myState.currentLine.length = len(myState.currentLine.line)
		easyterm.CursorPos(myState.cursorPos.y, 0) // move cursor to start
		easyterm.ClearLine()
		fmt.Print(myState.currentLine.line) // write updated line again	
		easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
	}
}

func main() {
	/* Terminal in raw mode */
	easyterm.Init()
	easyterm.Clear()
	easyterm.CursorPos(0,0)

	/* Init my position */
	myState.cursorPos = Cursor{0,0}
	
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
	myState.currentLine = sbGetLine(0)

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
						// TODO: Insert new line in list
						//easyterm.CursorNextLine(1)
						//updateCursorPosY(1)
						sbAddLineToBuffer(myState.cursorPos.y, myState.cursorPos.x)
						updateCursorPosY(1)
						easyterm.CursorPos(myState.cursorPos.y, 0)
						myState.currentLine = sbGetLine(myState.cursorPos.y)
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
						easyterm.CursorPos(0,0)
						easyterm.End()
						return
					case letter > 0 && letter <= 31:
						// Do nothing
	
					default:
						//fmt.Print(letter)
						///fmt.Printf("BytesRead: %d\n", bytesRead)
						fmt.Printf("%s", string(letter))
						updateCursorPosX(1)
						writeTextToBuffer(letter)
				}

			}

		} else {
			easyterm.Clear()
			easyterm.CursorPos(0,0)
			easyterm.End()
			return
		}
	}

}
