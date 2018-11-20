package main

import (
	"bufio"
	"easyterm"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"screenbuf"
	"strings"
)

/* Alias ReadWriter */
type File = os.File
type ReadWriter = bufio.ReadWriter
type ScreenBuffer = screenbuf.ScreenBuffer
type BufferNode = screenbuf.BufferNode

var (
	UP    = screenbuf.UP
	SPLIT = screenbuf.SPLIT
	DOWN  = screenbuf.DOWN
)

type Cursor struct {
	x int
	y int
}

/* Contains the state of the editor session */
type WinterState struct {
	cursorPos   Cursor
	currentLine *BufferNode
	fileName string
	filePath string
}

type WinterError struct {
	message string
	newFile bool
}

type IWinterError interface {
	Error()
	IsNewFile()
}

func (e *WinterError) Error() string {
	return fmt.Sprintf("%s", e.message)
}

func (e *WinterError) IsNewFile() bool {
	return e.newFile
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
	if (myState.cursorPos.x - 1) < line_length {
		if num < 0 {
			if sb.IsATabStop(myState.cursorPos.x-1) {
				if strings.ContainsRune(myState.currentLine.RealLine[0:myState.cursorPos.x-1], 9){
					for i:= myState.cursorPos.x-1; i >= 0; i-- {
						if myState.currentLine.RealLine[i] == '\t' {
							num *= ((myState.cursorPos.x - i) - 1) // to land on the char before the stop
							break
						}
					}
				}
			}
		}
		if myState.currentLine.RealLine[myState.cursorPos.x-1] == '\t' {
			if num > 0 {
				var nextTab int = sb.NextTabStop(myState.cursorPos.x-1)
				num *= (nextTab - myState.cursorPos.x) + 1 // Tab +1 to land on next editable char
			}
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
		currentLine      *BufferNode
		nextLine         *BufferNode
		currentLineIndex int
	)
	currentLineIndex = myState.cursorPos.y + num
	//easyterm.CursorPos(29, 80)
	//fmt.Print("Y-CurrentLineIndex: ")
	//fmt.Print(currentLineIndex)
	if currentLineIndex == 0 {
		// trying to go up in the file, load lines from above if any
		sb.LoadLine(UP, myState.currentLine.Index)

		currentLine = myState.currentLine
		nextLine = sb.GetLine(currentLine.Index + num) // Get the previous node

		if nextLine == nil {
			return
		}

		if nextLine.Length < currentLine.Length {
			easyterm.CursorPos(myState.cursorPos.y, nextLine.Length)
			setCursorPos(myState.cursorPos.y, nextLine.Length+1)
		}

		myState.currentLine = nextLine

	} else if currentLineIndex > 0 && currentLineIndex < sb.DefaultHeight {
		currentLine = myState.currentLine
		//updateCursorPosY(num)
		nextLine = sb.GetLine(currentLine.Index + num)

		if nextLine == nil {
			return
		}

		fn(int(math.Abs(float64(num)))) // Calls the easyterm func with a positive number
		updateCursorPosY(num)

		if nextLine.Length < currentLine.Length {
			//easyterm.CursorPos(0,1)
			easyterm.CursorPos(myState.cursorPos.y, nextLine.Length)
			setCursorPos(myState.cursorPos.y, nextLine.Length+1)
		}

		myState.currentLine = nextLine

	} else if currentLineIndex == sb.DefaultHeight {
		// load the lines to bottom
		sb.LoadLine(DOWN, myState.currentLine.Index)

		currentLine = myState.currentLine
		nextLine = sb.GetLine(currentLine.Index + num)

		if nextLine == nil {
			return
		}

		if nextLine.Length < currentLine.Length {
			easyterm.CursorPos(myState.cursorPos.y, nextLine.Length)
			setCursorPos(myState.cursorPos.y, nextLine.Length+1)
		}

		myState.currentLine = nextLine
	}
	showEditorData()
}

func showEditorData() {
	pos := 30
	pos2 := 80
	easyterm.CursorPos(pos, pos2)
	if myState.currentLine != nil {
		fmt.Printf("Current Line Index: %v", myState.currentLine.Index)
	}
	easyterm.CursorPos(pos+1, pos2)
	fmt.Printf("Buffer Length: %v", sb.Size())
	easyterm.CursorPos(pos+2, pos2)
	fmt.Print("Position:")
	easyterm.CursorPos(pos+3, pos2)
	fmt.Printf("X: %v Y: %v", myState.cursorPos.x, myState.cursorPos.y)
	easyterm.CursorPos(pos+4, pos2)
	fmt.Print("DEFAULT_HEIGHT: ")
	fmt.Print(sb.DefaultHeight)
	easyterm.CursorPos(pos+5, pos2)
	fmt.Printf("Index of first visible line: %v", sb.IndexOfFirstVisibleLine)
	easyterm.CursorPos(pos+6, pos2)
	fmt.Printf("Index of last visible line: %v", sb.IndexOfLastVisisbleLine)
	easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
}

func writeTextToBuffer(letter byte) {

	switch {
	// begining
	case myState.cursorPos.x == 1:
		myState.currentLine.Line = string(letter) + myState.currentLine.Line
		myState.currentLine.RealLine = packTabs(myState.currentLine.Line)
	// end
	case myState.cursorPos.x == (myState.currentLine.Length + 1):
		myState.currentLine.Line += string(letter)
		myState.currentLine.RealLine = packTabs(myState.currentLine.Line)
	// middle
	case myState.cursorPos.x < (myState.currentLine.Length + 1):
		total := make([]rune, myState.currentLine.Length+1)
		firstHalf := make([]rune, len(myState.currentLine.RealLine[0:myState.cursorPos.x-1]))
		copy(firstHalf, []rune(myState.currentLine.RealLine[0:myState.cursorPos.x-1]))
		secondHalf := make([]rune, len(myState.currentLine.RealLine[myState.cursorPos.x-1:myState.currentLine.Length]))
		copy(secondHalf, []rune(myState.currentLine.RealLine[myState.cursorPos.x-1:myState.currentLine.Length]))

		copy(total, []rune(string(firstHalf)+string(letter)+string(secondHalf)))
		myState.currentLine.Line = unpackTabs(string(total))
		myState.currentLine.RealLine = packTabs(myState.currentLine.Line)
	}
	easyterm.ClearLine()
	easyterm.CursorPos(myState.cursorPos.y, 1)
	fmt.Print(myState.currentLine.Line)
	easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
	myState.currentLine.Length = len(myState.currentLine.RealLine)
	if letter == 9 {
		curpos := sb.NextTabStop(myState.cursorPos.x - 1)
		easyterm.CursorRight(curpos)
		updateCursorPosX(curpos - (myState.cursorPos.x - 1))
	} else {
		easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x+1)
		updateCursorPosX(1)
	}
	showEditorData()
}

func backspaceLine() {

	// Handling tab backspace
	// Checking a TabSpace size window backwards in the line for a \t
	// var eightBack int = (myState.cursorPos.x - 1) - sb.TabSpace
	// if eightBack < 0 {
	// 	eightBack = 0
	// }
	// var prevTabStop int = sb.PrevTabStop(myState.cursorPos.x - 1)
	// var backSpacePos int = myState.cursorPos.x

	// if backSpacePos > myState.currentLine.Length && (myState.currentLine.Length > 0) {
	// 	backSpacePos -= 2
	// } else {
	// 	backSpacePos -= 1
	// }

	// if strings.ContainsRune(myState.currentLine.RealLine[prevTabStop:backSpacePos], 9) {
	// 	var spacesBack int = 0
	// 	var total []rune
	// 	var firstHalf []rune
	// 	var indexToDelete int = myState.cursorPos.x - 2

	// 	for i := myState.currentLine.Length - 1; i >= prevTabStop; i-- {
	// 		if myState.currentLine.RealLine[i] == '\t' {

	// 			if i == 0 {
	// 				firstHalf = make([]rune, len(myState.currentLine.RealLine[backSpacePos:myState.currentLine.Length]))
	// 				copy(firstHalf, []rune(myState.currentLine.RealLine[myState.cursorPos.x-1:myState.currentLine.Length]))

	// 				total = firstHalf

	// 			} else if i > 0 {
	// 				firstHalf = make([]rune, len(myState.currentLine.RealLine[0:i]))
	// 				copy(firstHalf, []rune(myState.currentLine.RealLine[0:i]))

	// 				secondHalf := make([]rune, len(myState.currentLine.RealLine[backSpacePos:myState.currentLine.Length]))
	// 				copy(secondHalf, []rune(myState.currentLine.RealLine[backSpacePos:myState.currentLine.Length]))

	// 				total = make([]rune, (len(firstHalf) + len(secondHalf)))
	// 				copy(total, []rune(string(firstHalf)+string(secondHalf)))
	// 			}

	// 			myState.currentLine.Line = unpackTabs(string(total))
	// 			myState.currentLine.RealLine = packTabs(myState.currentLine.Line)
	// 			myState.currentLine.Length = len(myState.currentLine.RealLine)

	// 			easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
	// 			easyterm.ClearLine()
	// 			easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
	// 			fmt.Print(myState.currentLine.Line)        // write updated line again
	// 			easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)

	// 			easyterm.CursorLeft(spacesBack)
	// 			updateCursorPosX(-1 * (spacesBack))

	// 			return
	// 		} else {
	// 			if i == 0 {
	// 				firstHalf = make([]rune, len(myState.currentLine.RealLine[backSpacePos:myState.currentLine.Length]))
	// 				copy(firstHalf, []rune(myState.currentLine.RealLine[myState.cursorPos.x-1:myState.currentLine.Length]))

	// 				total = firstHalf

	// 			} else if i > 0 {
	// 				firstHalf = make([]rune, len(myState.currentLine.RealLine[0:i]))
	// 				copy(firstHalf, []rune(myState.currentLine.RealLine[0:i]))

	// 				secondHalf := make([]rune, len(myState.currentLine.RealLine[backSpacePos:myState.currentLine.Length]))
	// 				copy(secondHalf, []rune(myState.currentLine.RealLine[backSpacePos:myState.currentLine.Length]))

	// 				total = make([]rune, (len(firstHalf) + len(secondHalf)))
	// 				copy(total, []rune(string(firstHalf)+string(secondHalf)))
	// 			}

	// 			myState.currentLine.Line = unpackTabs(string(total))
	// 			myState.currentLine.RealLine = packTabs(myState.currentLine.Line)
	// 			myState.currentLine.Length = len(myState.currentLine.RealLine)

	// 			easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
	// 			easyterm.ClearLine()
	// 			easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
	// 			fmt.Print(myState.currentLine.Line)        // write updated line again
	// 			easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)

	// 			easyterm.CursorLeft(1)
	// 			updateCursorPosX(-1)
	// 		}
	// 		spacesBack++ // the amount of spaces to move cursor back if necesary
	// 	}

	// }

	/*var tabPos int = (sb.PrevTabStop(myState.cursorPos.x - 1) - 1)
	if tabPos >= 0 && tabPos < myState.currentLine.Length {
		// This means we can check for tabs backwards
		if myState.currentLine.RealLine[tabPos] == '\t' {
			total := make([]rune, myState.currentLine.Length - sb.TabSpace)
			var firstHalf []rune
			if tabPos == 0 {
				firstHalf = make([]rune, len(myState.currentLine.RealLine[(myState.cursorPos.x - 1):myState.currentLine.Length]))
				copy(firstHalf, []rune(myState.currentLine.RealLine[(myState.cursorPos.x - 1):myState.currentLine.Length]))
				copy(total, []rune(string(firstHalf)))

			} else if tabPos > 0 {
				firstHalf = make([]rune, len(myState.currentLine.RealLine[0:tabPos]))
				copy(firstHalf, []rune(myState.currentLine.RealLine[0:tabPos]))

				secondHalf := make([]rune, len(myState.currentLine.RealLine[(myState.cursorPos.x - 1):myState.currentLine.Length]))
				copy(secondHalf, []rune(myState.currentLine.RealLine[(myState.cursorPos.x - 1):myState.currentLine.Length]))

				copy(total, []rune(string(firstHalf) + string(secondHalf)))
			}

			myState.currentLine.Line = unpackTabs(string(total))
			myState.currentLine.RealLine = packTabs(myState.currentLine.Line)
			myState.currentLine.Length = len(myState.currentLine.RealLine)

			easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
			easyterm.ClearLine()
			easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
			fmt.Print(myState.currentLine.Line) // write updated line again
			easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)

			easyterm.CursorLeft(sb.TabSpace)
			updateCursorPosX(-1 * sb.TabSpace)

			return
		}
	}*/

	if myState.currentLine.Length == 0 {
		// delete node, rehook list ends and move cursor to end of previous line
		var (
			prev *BufferNode
			next *BufferNode
		)
		prev = myState.currentLine.Prev
		next = myState.currentLine.Next

		if prev != nil {
			// verify that if we are working with the first visible line we update the buffer
			if myState.currentLine.Index == sb.IndexOfFirstVisibleLine {
				sb.IndexOfFirstVisibleLine -= 1
			}
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
		// When we delete a line we can load the next one on screen if any in sb, else load if we have any
		lastLineOnScreen := sb.GetLine(sb.IndexOfLastVisisbleLine - 1)
		if lastLineOnScreen != nil {
			if lastLineOnScreen.Next == nil {
				sb.LoadLine(DOWN, lastLineOnScreen.Index)
			} else {
				sb.IndexOfLastVisisbleLine = lastLineOnScreen.Next.Index
			}
		}
		sb.ReprintBuffer() // reprints complete buffer
		if myState.cursorPos.y > sb.DefaultHeight {
			var ypos int = sb.DefaultHeight - (myState.cursorPos.y - sb.DefaultHeight)
			easyterm.CursorPos(ypos, myState.currentLine.Length+1)
			setCursorPos(ypos, myState.currentLine.Length+1)
		} else {
			easyterm.CursorPos(myState.cursorPos.y, myState.currentLine.Length+1)
			setCursorPos(myState.cursorPos.y, myState.currentLine.Length+1)
		}
		showEditorData()
		return
	}

	// TODO: single line wrapping handling

	switch {
	// Between the first and last characters of a line
	case myState.cursorPos.x > 1 && myState.cursorPos.x < (myState.currentLine.Length+1):

		var firstHalf []rune
		var total []rune
		var pos int = myState.cursorPos.x - 1
		var moveCursor int = 0

		if sb.IsATabStop(pos) {

			if strings.ContainsRune(myState.currentLine.RealLine[0:pos], 9){
				for i:= pos; i >= 0; i-- {
					if myState.currentLine.RealLine[i] == '\t' {
						pos = (pos - i) // to land on the char before the stop
						break
					}
				}
				total = make([]rune, myState.currentLine.Length-pos)
				firstHalf = make([]rune, len(myState.currentLine.RealLine[0:pos-2]))
				copy(firstHalf, []rune(myState.currentLine.RealLine[0:pos-2]))
				moveCursor = pos
			} else {
				total = make([]rune, myState.currentLine.Length-1)
				firstHalf = make([]rune, len(myState.currentLine.RealLine[0:myState.cursorPos.x-2]))
				copy(firstHalf, []rune(myState.currentLine.RealLine[0:myState.cursorPos.x-2]))
				moveCursor = 1
			}
			
		} else {
			total = make([]rune, myState.currentLine.Length-1)
			firstHalf = make([]rune, len(myState.currentLine.RealLine[0:myState.cursorPos.x-2]))
			copy(firstHalf, []rune(myState.currentLine.RealLine[0:myState.cursorPos.x-2]))
			moveCursor = 1
		}
		
		secondHalf := make([]rune, len(myState.currentLine.RealLine[myState.cursorPos.x-1:myState.currentLine.Length]))
		copy(secondHalf, []rune(myState.currentLine.RealLine[myState.cursorPos.x-1:myState.currentLine.Length]))

		copy(total, []rune(string(firstHalf)+string(secondHalf)))
		myState.currentLine.Line = unpackTabs(string(total))
		myState.currentLine.RealLine = packTabs(myState.currentLine.Line)
		myState.currentLine.Length = len(myState.currentLine.RealLine)

		// reprint line
		easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
		easyterm.ClearLine()
		easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
		fmt.Print(myState.currentLine.Line)        // write updated line again
		easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
		// move & update cursor
		easyterm.CursorLeft(moveCursor)
		updateCursorPosX(-1 * moveCursor)

	// At the start of a line
	case myState.cursorPos.x == 1:

		// check if we have line on top, if not do nothing, else move it up
		var prev *BufferNode
		prev = myState.currentLine.Prev
		if prev != nil {
			// verify that if we are working with the first visible line we update the buffer
			if myState.currentLine.Index == sb.IndexOfFirstVisibleLine {
				sb.IndexOfFirstVisibleLine -= 1
				sb.IndexOfLastVisisbleLine -= 1
			}
			// move current line up
			origPrevLineLength := prev.Length // to move the cursor later
			currentText := make([]rune, myState.currentLine.Length)
			copy(currentText, []rune(myState.currentLine.RealLine))
			prev.Line += unpackTabs(string(currentText))
			prev.RealLine = packTabs(prev.Line)
			prev.Length = len(prev.RealLine)

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
			// When we delete a line we can load the next one on screen if any in sb, else load if we have any
			lastLineOnScreen := sb.GetLine(sb.IndexOfLastVisisbleLine - 1)
			if lastLineOnScreen != nil {
				if lastLineOnScreen.Next == nil {
					sb.LoadLine(DOWN, lastLineOnScreen.Index)
				} else {
					sb.IndexOfLastVisisbleLine = lastLineOnScreen.Next.Index
				}
			}
			sb.ReprintBuffer() // reprints complete buffer
			easyterm.CursorPos(myState.cursorPos.y, origPrevLineLength+1)
			setCursorPos(myState.cursorPos.y, origPrevLineLength+1)
			//easyterm.CursorPos(20, 1)
			//fmt.Print(sbGetBufferLength())

		}
		// At the very end of a line
	case myState.cursorPos.x == (myState.currentLine.Length + 1):

		var newLine []rune
		var pos int = myState.cursorPos.x - 1
		var moveCursor int = 0

		if sb.IsATabStop(pos) {

			if strings.ContainsRune(myState.currentLine.RealLine[0:pos], 9){
				for i:= pos - 1; i >= 0; i-- {
					if myState.currentLine.RealLine[i] == '\t' {
						pos = (pos - i) // to land on the char before the stop
						break
					}
				}
				moveCursor = pos
				newLine = make([]rune, myState.currentLine.Length-pos)
				copy(newLine, []rune(myState.currentLine.RealLine[0:myState.currentLine.Length-pos]))
			} else {
				moveCursor = 1
				newLine = make([]rune, myState.currentLine.Length-1)
				copy(newLine, []rune(myState.currentLine.RealLine[0:myState.currentLine.Length-1]))
			}
		} else {
			moveCursor = 1
			newLine = make([]rune, myState.currentLine.Length-1)
			copy(newLine, []rune(myState.currentLine.RealLine[0:myState.currentLine.Length-1]))
		}

		//newLine := make([]rune, myState.currentLine.Length-1)
		//copy(newLine, []rune(myState.currentLine.RealLine[0:myState.currentLine.Length-1]))
		myState.currentLine.Line = unpackTabs(string(newLine))
		myState.currentLine.RealLine = packTabs(myState.currentLine.Line)
		myState.currentLine.Length = len(myState.currentLine.RealLine)

		// reprint line
		easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
		easyterm.ClearLine()
		easyterm.CursorPos(myState.cursorPos.y, 1) // move cursor to start
		fmt.Print(myState.currentLine.Line)        // write updated line again
		easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
		// move & update cursor
		easyterm.CursorLeft(moveCursor)
		updateCursorPosX(-1 * moveCursor)
	}
	showEditorData()
}

func packTabs(line string) string {
	return sb.PackTabs(line)
}

func unpackTabs(line string) string {
	return sb.UnpackTabs(line)
}

func saveFile() {
	sb.Save()
	easyterm.CursorPos(myState.cursorPos.y, myState.cursorPos.x)
}

// TODO: When the file is new or temp, don't actually create the file until it's saved by user
func handleArguments(args []string) (*File, error) {
	var (
		fileName string
		filePath string
	)
	switch len(args) {

	case 1:
		myState.fileName = ""
		myState.filePath = "."
		return nil, &WinterError{"No file name entered.", true}
	case 2:
		fileName = filepath.Base(os.Args[1])
		filePath = filepath.Dir(os.Args[1])
		myState.fileName = fileName
		myState.filePath = filePath
		//pwd, _ := os.Getwd()
		if fileInfo, existsErr := os.Stat(filePath + "/" + fileName); !os.IsNotExist(existsErr) {
			if file, err := os.OpenFile(filePath+"/"+fileName, os.O_RDWR, fileInfo.Mode()); err == nil {
				// Found the file, lets load it after
				return file, nil

			} else {
				// Something happened, throw error
				return nil, err
			}
		} else {
			return nil, &WinterError{"New file.", true}
		}
	default:
		return nil, &WinterError{"Argument Error.", false}
	}
}

func main() {
	/* Terminal in raw mode */
	easyterm.Init()
	easyterm.Clear()
	easyterm.CursorPos(1, 1)

	/* Init my position */
	myState.cursorPos = Cursor{1, 1}

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
		// If the file doesn't exist or if something went wrong
		if nwerr, ok := err.(*WinterError); ok && nwerr.IsNewFile() {
			sb = screenbuf.NewScreenBuffer(nil)
			sb.FileName = myState.fileName
			sb.FilePath = myState.filePath
			sb.LoadFile()
		} else {
			fmt.Print("-winter: ")
			fmt.Println(err)
			easyterm.CursorPos(2, 1)
			easyterm.End()
			os.Exit(1)
		}
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
					var oldLineIndex int = myState.currentLine.Index
					sb.AddLineToBuffer(myState.currentLine.Index, myState.cursorPos.x, myState.cursorPos.y)
					//myState.currentLine = sb.GetLine(newIndex)
					sb.Dirty = true
					if myState.cursorPos.y < sb.DefaultHeight {
						updateCursorPosY(1)
					}
					easyterm.CursorPos(myState.cursorPos.y, 1)
					setCursorPos(myState.cursorPos.y, 1)
					myState.currentLine = sb.GetLine(oldLineIndex + 1)
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
					sb.Dirty = true
				case letter == 27:
					// Do Nothing for Esc for now

				case letter == 19:
					// Save
					saveFile()

				case letter == 17:
					// Ctrl-Q
					easyterm.Clear()
					easyterm.CursorPos(1, 1)
					easyterm.End()
					return

				case letter == 9:
					// Tab
					writeTextToBuffer(letter)
					sb.Dirty = true
				case letter > 0 && letter <= 31:
					// Do nothing

				default:
					//fmt.Print(letter)
					///fmt.Printf("BytesRead: %d\n", bytesRead)
					//fmt.Printf("%s", string(letter))
					//updateCursorPosX(1)
					writeTextToBuffer(letter)
					sb.Dirty = true
				}

			}

		} else {
			easyterm.Clear()
			easyterm.CursorPos(1, 1)
			easyterm.End()
			return
		}
	}

}
