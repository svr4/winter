package main

import (
	"fmt"
	"easyterm"
	//"strconv"
)

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
}

/* The buffer */
var buffer ScreenBuffer


func sbInitNewFile() {
	//var traveler = &BufferNode{}
	//buffer.length = 0;
	/*if w, h, err := easyterm.GetSize(); err == nil {
		DEFAULT_HEIGHT = h
		DEFAULT_WIDTH = w
	}*/
	// Insert the first line in new file
	var temp = &BufferNode{}
	temp.index = 1
	temp.line = "test"
	temp.length = 4
	temp.prev = nil
	temp.next = nil
	buffer.head = temp
	buffer.length = 1;

	/*for i := 0; i < DEFAULT_HEIGHT; i++ {
		var temp = &BufferNode{}
		temp.index = i
		temp.line = "~"
		temp.length = 1
		if i == 0 {
			temp.prev = nil
			temp.next = nil
			buffer.head = temp
			traveler = buffer.head
		} else {
			traveler.next = temp
			temp.prev = traveler
			temp.next = nil
			traveler = temp
		}
		buffer.length += 1
	}*/

}

/*func sbInitExistingFile(fileName string) {


}*/

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
// TODO: spliting a line and pressing enter again and again not moving split line down
func sbAddLineToBuffer(line, column int) {

	for traveler := buffer.head; traveler != nil; traveler = traveler.next {
		if traveler.index == line {

			var temp = &BufferNode{}
			temp.line = ""
			temp.length = 0
			// col - 1 because screen is 1 based and strings are 0 based
			//fmt.Printf("%v\n", temp.prev)
			var insertWhere = sbManageNewLineString(column, buffer.head.length)
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
