package winter

import (
	"fmt"
	"easyterm"
	"bufio"
	"io"
	"bytes"
	"os"
	//"strconv"
)
type Buffer = bytes.Buffer
type File = os.File
var DEFAULT_HEIGHT, DEFAULT_WIDTH = 25, 80
// Consts represent what to do with the
// content when the user presses enter on a line
// depending on the column index
const (
	UP = iota
	SPLIT = iota
	DOWN = iota
)

type BufferNode struct {
	index int
	line string
	length int
	next *BufferNode
	prev *BufferNode
}

type ScreenBuffer struct {
	head *BufferNode
	length int
	maxLength int
	filePtrIndex int64
	filePtr *File
	fileRW *ReadWriter
}

/* The buffer */
var buffer ScreenBuffer


func sbLoadFile(file *File) {
	// init buffer length
	buffer.length = 0;
	buffer.filePtrIndex = -1
	buffer.filePtr = file
	// get current window dimensions
	if w, h, err := easyterm.GetSize(); err == nil {
		DEFAULT_HEIGHT = h
		DEFAULT_WIDTH = w
		buffer.maxLength = h // set the max length of list to screen height
	} else {
		buffer.maxLength = DEFAULT_HEIGHT
	}
	// read/writer for the file
	buffer.fileRW = bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))
	var temp = &BufferNode{}
	var traveler = &BufferNode{}
	for i := 1; i < DEFAULT_HEIGHT; i ++ {
		lineBytes, err := sbReadLine(buffer.fileRW);
		//fmt.Print(string(lineBytes))
		// already at the end
		if err == io.EOF && buffer.length == 0 {
			temp.index = i
			temp.line = ""
			temp.length = len(temp.line)
			temp.prev = nil
			temp.next = nil
			buffer.head = temp
			buffer.length++;
			buffer.filePtrIndex++
			break
		}
		// Not at EOF
		if err == nil {
			if i == 1 {
				temp.index = i
				temp.line = string(lineBytes)
				temp.length = len(temp.line)
				temp.prev = nil
				temp.next = nil
				buffer.head = temp
				buffer.length++;
				traveler = buffer.head
				buffer.filePtrIndex++
			} else {
				temp = &BufferNode{}
				temp.index = i
				temp.line = string(lineBytes)
				temp.length = len(temp.line)
				temp.prev = traveler
				temp.next = nil

				traveler.next = temp
				traveler = traveler.next
				buffer.length++
				buffer.filePtrIndex++
			}
		}
	}

}

func sbReadLine(fileRW *ReadWriter) ([]byte, error) {

	lineBytes, isPrefix, err := fileRW.ReadLine()
	if err == nil {

		if isPrefix {
			var buffer Buffer
			buffer.Write(lineBytes)
			for isPrefix {
				lineBytes, isPrefix, err = fileRW.ReadLine()
				if err == nil {
						buffer.Write(lineBytes)
				}
			}
			copyOfBuffer := make([]byte, buffer.Len())
			copy(copyOfBuffer, buffer.Bytes())
			return copyOfBuffer, nil
		} else {
			return lineBytes, nil
		}

	} else {
		return make([]byte, 0), err
	}

}

func sbEnqueueLine(line []byte, where int) {
	// add line via reading or add line via enter
	traveler := sbGetLine(buffer.length)
	switch where {
	case UP:

	case DOWN:
		buffer.length++
		var temp = &BufferNode{}
		temp.index = buffer.length
		temp.line = string(line)
		temp.length = len(temp.line)
		temp.prev = traveler
		temp.next = nil

		if traveler.next == nil {
			traveler.next = temp
		}
	}
}

func sbDequeueLine() {

}

func sbLoadLine(fromWhere int) {
	switch fromWhere {
	case UP:

	case DOWN:
		newPtrIndex, err := buffer.filePtr.Seek(buffer.filePtrIndex+1, 0)
		if err == nil {
			buffer.filePtrIndex = newPtrIndex
			lineBytes, err2 := sbReadLine(buffer.fileRW)
			if err2 == nil {
				// We got a line, lets put it on the screen
				sbEnqueueLine(lineBytes, DOWN)
				sbPrintBuffer()
			}
		}
	}
}

func sbPrintBuffer() {
	if w, h, err := easyterm.GetSize(); err == nil {
		DEFAULT_HEIGHT = h
		DEFAULT_WIDTH = w
	}

	var i int = 1
	for traveler := buffer.head; traveler != nil; traveler = traveler.next {
		fmt.Printf("%s\n", traveler.line)
		easyterm.CursorPos(i,1)
		i++
	}
	// Move cursor to next line
	easyterm.CursorPos(i+1,1)

	// Fill the rest of the screen with ~ if some space available
	for x:= i; x < DEFAULT_HEIGHT; x++ {
		fmt.Printf("~\n")
		easyterm.CursorPos(x,1)
	}
	easyterm.CursorPos(1,1)
	myState.cursorPos.x = 1
	myState.cursorPos.y = 1
	//fmt.Println("lenght: " + strconv.Itoa(buffer.length))
}

func sbGetLineLength(line int) int {
	i := 1
	for traveler := buffer.head; traveler != nil; traveler = traveler.next {
		if traveler.index == line {
			i = traveler.length
			break
		}
	}
	return i
}

func sbGetFilePtrIndex() int64 {
	return buffer.filePtrIndex
}

func sbGetBufferLength() int {
	size := 0
	for traveler := buffer.head; traveler != nil; traveler = traveler.next {
		size++
	}
	return size
	//return buffer.length
}

func sbUpdateBufferIndexes() {
	i := 1
	for traveler := buffer.head; traveler != nil; traveler = traveler.next {
		traveler.index = i
		i++
	}
}

func sbPrintLine(line int) {
	for traveler := buffer.head; traveler != nil; traveler = traveler.next {
		if traveler.index == line {
			fmt.Print(traveler.line)
			break
		}
	}
}

func sbGetLine(line int) *BufferNode {
	var lineNode *BufferNode
	for traveler := buffer.head; traveler != nil; traveler = traveler.next {
		if traveler.index == line {
			lineNode = traveler
			break
		}
	}
	return lineNode
}

func sbManageNewLineString(col, length int) int {
	if col == 1 {
		return UP
	} else if col > 1 && col < length {
		return SPLIT
	} else if col >= length {
		return DOWN
	} else {
		return -1
	}
}

func sbReprintBuffer() {
	i := 1
	for traveler := buffer.head; traveler != nil; traveler = traveler.next {
		easyterm.CursorPos(i,1)
		easyterm.ClearLine()
		easyterm.CursorPos(i,1)
		if traveler.next != nil {
			fmt.Printf("%s\n", traveler.line)
		} else {
			fmt.Printf("%s", traveler.line)
		}
		//fmt.Printf("%d\n", traveler.index)
		i++
	}

	for x:= i; x < DEFAULT_HEIGHT; x++ {
		easyterm.CursorPos(i,1)
		easyterm.ClearLine()
		easyterm.CursorPos(i,1)
		fmt.Printf("~\n")
		easyterm.CursorPos(x,1)
	}

}

func sbAddLineToBuffer(line, column int) {

	for traveler := buffer.head; traveler != nil; traveler = traveler.next {
		if traveler.index == line {

			var temp = &BufferNode{}
			temp.line = ""
			temp.length = 0
			// col - 1 because screen is 1 based and strings are 0 based
			//fmt.Printf("%v\n", temp.prev)
			var insertWhere = sbManageNewLineString(column, traveler.length)
			//fmt.Printf("%d\n",insertWhere)
			switch insertWhere {
				case UP:
					var prev *BufferNode
					prev = traveler.prev
					//temp.index = traveler.index
					temp.next = traveler
					if prev != nil {
						temp.prev = prev
						prev.next = temp
					} else {
						temp.prev = nil
					}
					traveler.prev = temp
					if traveler.index == 1 {
						buffer.head = temp
					}
					//traveler.index += 1
					buffer.length += 1
				case DOWN:
					var next *BufferNode
					next = traveler.next
					temp.prev = traveler
					//temp.index = traveler.index + 1
					if next != nil {
						next.prev = temp
						temp.next = next
					}
					traveler.next = temp
					buffer.length += 1
				case SPLIT:
					var next *BufferNode
					next = traveler.next
					temp.prev = traveler
					if next != nil {
						next.prev = temp
						temp.next = next
					}
					traveler.next = temp
					buffer.length += 1
					// Split text
					origText := make([]rune, traveler.length)
					// Copy original text to origText
					copy(origText, []rune(traveler.line))
					// New original lines text
					origRune := make([]rune, len(origText[0:(column - 1)]))
					// Copying original lines new text to rune
					copy(origRune, origText[0:column - 1])
					// slice with new line text
					newRune := origText[(column - 1):traveler.length]
					newText := make([]rune, len(newRune))
					copy(newText, newRune)
					// set the string on the new line
					temp.line = string(newText)
					temp.length = len(temp.line)

					// update the old lines text
					traveler.line = string(origRune)
					traveler.length = len(traveler.line)

					/*traveler.line = string(origText[0:column])
					traveler.length = len(traveler.line)*/
			}

			sbUpdateBufferIndexes()
			sbReprintBuffer()
			//easyterm.CursorPos(line,column)
		}
	}

}
