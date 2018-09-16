package main

import (
	"easyterm"
	"bufio"
	"os"
	"fmt"
	"math"
	"log"
	"screenbuf"
	"strings"
)
/* Alias ReadWriter */
type File = os.File
type ReadWriter = bufio.ReadWriter
type ScreenBuffer = screenbuf.ScreenBuffer
type BufferNode = screenbuf.BufferNode

var (
	UP = screenbuf.UP
	SPLIT = screenbuf.SPLIT
	DOWN = screenbuf.DOWN
)
type Cursor struct {
	x int
	y int
}

/* Contains the state of the editor session */
type WinterState struct {
	cursorPos Cursor
	currentLine *BufferNode
	fileName string
	//filePtr *File
}

type WinterError struct {
	message string
}

func (e *WinterError) Error() string {
	return fmt.Sprintf("%s", e.message)
}

/* Global State */
var myState WinterState

/* stdin/stdout RW */
var termRW *ReadWriter

/* the ScreenBuffer*/
var sb *ScreenBuffer

func clearBuffer(buff []byte) {
	buff[0] = 0
	buff[1] = 0
	buff[2] = 0
	buff[3] = 0
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
	line_length := myState.currentLine.Length
	// Handling tab movement
	if ((myState.cursorPos.x - 1) < line_length) {
		if myState.currentLine.RealLine[myState.cursorPos.x - 1] == '\t' {
			num *= 8 // Tab
		}
	}
	// Current pos + the next column < the length of the line
	if ((myState.cursorPos.x - 1) + num) <= line_length {
		fn(int(math.Abs(float64(num)))) // Calls the easyterm func with a positive number
		updateCursorPosX(num)
	}
	showEditorData()
}

func moveCursorY(num int, fn func(int)) {
	//buffer_length := sbGetBufferLength()
	var (
		currentLine *BufferNode
		nextLine *BufferNode
		currentLineIndex int
	)
	currentLineIndex = myState.cursorPos.y + num
	//easyterm.CursorPos(29, 80)
	//fmt.Print("Y-CurrentLineIndex: ")
	//fmt.Print(currentLineIndex)
	if currentLineIndex < 0 {
		// trying to go up in the file, load lines from above if any

	} else if currentLineIndex > 0 && currentLineIndex < sb.DefaultHeight {
		currentLine = sb.GetLine(myState.cursorPos.y)
		//updateCursorPosY(num)
		nextLine = sb.GetLine(myState.cursorPos.y + num)

		if nextLine == nil {
			return
		}

		fn(int(math.Abs(float64(num)))) // Calls the easyterm func with a positive number
		updateCursorPosY(num)

		if nextLine.Length < currentLine.Length {
			//easyterm.CursorPos(0,1)
			easyterm.CursorPos(myState.cursorPos.y, nextLine.Length)
			setCursorPos(myState.cursorPos.y, nextLine.Length + 1)
		}

		myState.currentLine = nextLine
		showEditorData()
	} else if currentLineIndex == sb.DefaultHeight {
		// load the lines to bottom
		sb.LoadLine(DOWN)
	}
}

func showEditorData() {
	pos := 30
	pos2 := 80
	easyterm.CursorPos(pos, pos2)
	if myState.currentLine != nil {
		fmt.Printf("Current Line Index: %v", myState.currentLine.Index)
	}
	easyterm.CursorPos(pos+1, pos2)
	fmt.Printf("Buffer Length: %v", sb.GetBufferLength())
	easyterm.CursorPos(pos+2, pos2)
	fmt.Print("Position:")
	easyterm.CursorPos(pos+3, pos2)
	fmt.Printf("X: %v Y: %v", myState.cursorPos.x, myState.cursorPos.y)
	easyterm.CursorPos(pos+4, pos2)
	fmt.Print("DEFAULT_HEIGHT: ")
	fmt.Print(sb.DefaultHeight)
	easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
}

func writeTextToBuffer(letter byte) {

	switch {
	// begining
	case myState.cursorPos.x == 1:
		myState.currentLine.Line = string(letter) + myState.currentLine.Line
		myState.currentLine.RealLine = strings.Replace(myState.currentLine.Line, "\t", "\t" + sb.TabFiller, -1)
	// end
	case myState.cursorPos.x == (myState.currentLine.Length + 1):
		myState.currentLine.Line += string(letter)
		myState.currentLine.RealLine = strings.Replace(myState.currentLine.Line, "\t", "\t" + sb.TabFiller, -1)
	// middle
	case myState.cursorPos.x < (myState.currentLine.Length + 1):
		total := make([]rune, myState.currentLine.Length + 1)
		firstHalf := make([]rune, len(myState.currentLine.RealLine[0:myState.cursorPos.x - 1]))
		copy(firstHalf, []rune(myState.currentLine.RealLine[0:myState.cursorPos.x - 1]))
		secondHalf := make([]rune, len(myState.currentLine.RealLine[myState.cursorPos.x-1:myState.currentLine.Length]))
		copy(secondHalf, []rune(myState.currentLine.RealLine[myState.cursorPos.x-1:myState.currentLine.Length]))

		copy(total, []rune(string(firstHalf) + string(letter) + string(secondHalf)))
		myState.currentLine.Line = unpackTabs(string(total))
		myState.currentLine.RealLine = packTabs(myState.currentLine.Line)
	}
	easyterm.ClearLine()
	easyterm.CursorPos(myState.cursorPos.y, 1)
	fmt.Print(myState.currentLine.Line)
	easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
	myState.currentLine.Length = len(myState.currentLine.RealLine)
	if letter == 9 {
		easyterm.CursorRight(sb.TabSpace)
		myState.cursorPos.x += sb.TabSpace
	} else {
		easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x + 1)
	}
}

func backspaceLine() {

	if myState.currentLine.Length == 0 {
		// delete node, rehook list ends and move cursor to end of previous line
		var (
			prev *BufferNode
			next *BufferNode
		)
		prev = myState.currentLine.Prev
		next = myState.currentLine.Next

		if prev != nil {
			prev.Next = next
		} else {
			return
		}

		if next != nil {
			next.Prev = prev
		}

		myState.currentLine.Prev = nil
		myState.currentLine.Next = nil

		myState.currentLine = prev

		updateCursorPosY(-1) // move up one line in state
		sb.UpdateBufferIndexes()
		sb.ReprintBuffer() // reprints complete buffer
		easyterm.CursorPos(myState.cursorPos.y, myState.currentLine.Length + 1)
		setCursorPos(myState.cursorPos.y, myState.currentLine.Length + 1)
		showEditorData()
		return
	}


	// TODO: single line wrapping handling

	switch {
		// Between the first and last characters of a line
	case myState.cursorPos.x > 1 && myState.cursorPos.x < (myState.currentLine.Length + 1):

		total := make([]rune, myState.currentLine.Length - 1)
		firstHalf := make([]rune, len(myState.currentLine.RealLine[0:myState.cursorPos.x-2]))
		copy(firstHalf, []rune(myState.currentLine.RealLine[0:myState.cursorPos.x-2]))
		secondHalf := make([]rune, len(myState.currentLine.RealLine[myState.cursorPos.x-1:myState.currentLine.Length]))
		copy(secondHalf, []rune(myState.currentLine.RealLine[myState.cursorPos.x-1:myState.currentLine.Length]))

		copy(total, []rune(string(firstHalf) + string(secondHalf)))
		myState.currentLine.Line = unpackTabs(string(total))
		myState.currentLine.RealLine = packTabs(myState.currentLine.Line)
		myState.currentLine.Length = len(myState.currentLine.RealLine)

		// reprint line
		easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
		easyterm.ClearLine()
		easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
		fmt.Print(myState.currentLine.Line) // write updated line again
		easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
		// move & update cursor
		easyterm.CursorLeft(1)
		updateCursorPosX(-1)

	case myState.cursorPos.x == 1:

		// check if we have line on top, if not do nothing, else move it up
		var prev *BufferNode
		prev = myState.currentLine.Prev
		if prev != nil {
			// move current line up
			origPrevLineLength := prev.Length // to move the cursor later
			currentText := make([]rune, myState.currentLine.Length)
			copy(currentText, []rune(myState.currentLine.RealLine))
			prev.Line += unpackTabs(string(currentText))
			prev.RealLine = packTabs(prev.Line)
			prev.Length = len(prev.Line)

			// rehook list, sans the soon to be destroyed node and update currentLine state
			var next *BufferNode
			next = myState.currentLine.Next
			if next != nil {
				prev.Next = next
				next.Prev = prev
				myState.currentLine.Prev = nil
				myState.currentLine.Next = nil
			} else {
				prev.Next = nil
				myState.currentLine.Prev = nil
			}

			myState.currentLine = prev
			updateCursorPosY(-1) // move up one line in state
			/*easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start of new line
			easyterm.ClearLine()
			easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
			fmt.Print(myState.currentLine.line) // write updated line again*/
			sb.UpdateBufferIndexes()
			sb.ReprintBuffer() // reprints complete buffer
			easyterm.CursorPos(myState.cursorPos.y, origPrevLineLength+1)
			setCursorPos(myState.cursorPos.y, origPrevLineLength + 1)
			//easyterm.CursorPos(20, 1)
			//fmt.Print(sbGetBufferLength())

		}

	case myState.cursorPos.x == (myState.currentLine.Length + 1):

		newLine := make([]rune, myState.currentLine.Length-1)
		copy(newLine, []rune(myState.currentLine.RealLine[0:myState.currentLine.Length-1]))
		myState.currentLine.Line = unpackTabs(string(newLine))
		myState.currentLine.RealLine = packTabs(myState.currentLine.Line)
		myState.currentLine.Length = len(myState.currentLine.Line)

		// reprint line
		easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
		easyterm.ClearLine()
		easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
		fmt.Print(myState.currentLine.Line) // write updated line again
		easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
		// move & update cursor
		easyterm.CursorLeft(1)
		updateCursorPosX(-1)
	}
	showEditorData()
}

func packTabs(line string) string {
	return strings.Replace(line, "\t", "\t" + sb.TabFiller, -1)
}

func unpackTabs(line string) string {
	return strings.Replace(line, "\t" + sb.TabFiller, "\t", -1)
}

// TODO: When the file is new or temp, don't actually create the file until it's saved by user
func handleArguments(args []string) (*File, error) {

	switch len(args) {

	case 1:
		// new file with no name, when saved do it with random name
		myState.fileName = "winterTemp"
		if file, err := os.OpenFile(myState.fileName, os.O_RDWR|os.O_CREATE, 0666); err == nil {
			return file, nil
		} else {
			return nil, err
		}
	case 2:
		myState.fileName = os.Args[1]
		if file, err := os.OpenFile(myState.fileName, os.O_RDWR|os.O_CREATE, 0666); err == nil {
			// Found the file, lets load it after
			return file, nil

		} else {
			// Something happened, throw error
			return nil, err
		}
	default:
		return nil, &WinterError{"Argument Error."}
	}
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

	// Handles file name, myState.filePtr should have been set
	file, err := handleArguments(os.Args)
	if err == nil {
		/* Init the ScreenBuffer */
		sb = screenbuf.NewScreenBuffer(file)
		sb.LoadFile()
	} else {
		log.Fatal(err)
	}
	sb.PrintBuffer()
	myState.cursorPos.x = 1
	myState.cursorPos.y = 1
	/* Get first line node to have something to write to */
	myState.currentLine = sb.GetLine(1)
	showEditorData()

	for {
		if bytesRead, err := termRW.Reader.Read(buffer); err == nil {
			//fmt.Printf("%s", string(letter))
			/* Means that the arrow keys where pressed */
			/* This will send 3 bytes: Esc, [ and (A or B or C or D)  */
			if bytesRead > 1 {
				//fmt.Print(string(buffer[0]))
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
				//fmt.Print(letter)
				switch {
					case letter == 13:
						// Enter
						//easyterm.CursorNextLine(1)
						//updateCursorPosY(1)
						sb.AddLineToBuffer(myState.cursorPos.y, myState.cursorPos.x)
						updateCursorPosY(1)
						easyterm.CursorPos(myState.cursorPos.y, 1)
						setCursorPos(myState.cursorPos.y, 1)
						myState.currentLine = sb.GetLine(myState.cursorPos.y)
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

					case letter == 9:
						// Tab
						writeTextToBuffer(letter)
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
