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
	temp.index = 0
	temp.line = "cock"
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
		easyterm.CursorPos(i,0)
		i++
	}
	// Move cursor to next line	
	easyterm.CursorPos(i+1,0)

	// Fill the rest of the screen with ~ if some space available
	for x:= i; x < DEFAULT_HEIGHT; x++ {
		fmt.Print("~")
		easyterm.CursorPos(x,0)
	}
	easyterm.CursorPos(0,0)
	myState.cursorPos.x = 0
	myState.cursorPos.y = 0
	//fmt.Println("lenght: " + strconv.Itoa(buffer.length))
}

func sbGetLineLength(line int) int {
	i := 0
	for traveler := buffer.head; traveler != nil; traveler = traveler.next {
		if traveler.index == line {
			i = traveler.length
			break
		}
	}
	return i
}

func sbGetBufferLength() int {
	return buffer.length
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
	if col == 0 {
		return UP
	} else if col > 0 && col < length {
		return SPLIT
	} else if col == length {
		return DOWN
	} else {
		return -1
	}
}

func sbReprintBuffer() {
	i := 0
	for traveler := buffer.head; traveler != nil; traveler = traveler.next {
		easyterm.CursorPos(i, 0)
		easyterm.ClearLine()
		easyterm.CursorPos(i,0)
		fmt.Printf("%s\n", traveler.line)
		i++
	}
}

func sbAddLineToBuffer(line, column int) {
	
	for traveler := buffer.head; traveler != nil; traveler = traveler.next {
		if traveler.index == line {

			var temp = &BufferNode{}
			temp.line = ""
			temp.length = 0

			var insertWhere = sbManageNewLineString(column, buffer.head.length)
			switch insertWhere {
				case UP:
					var prev *BufferNode
					prev = traveler.prev
					temp.index = traveler.index
					temp.next = traveler
					if prev != nil {
						temp.prev = prev
					} else {
						temp.prev = nil
					}
					traveler.prev = temp
					traveler.index += 1
				case DOWN:
					var next *BufferNode
					next = traveler.next
					temp.prev = traveler
					temp.index = traveler.index + 1
					if next != nil {
						next.prev = temp
						temp.next = next
					}
				case SPLIT:
					var next *BufferNode
					next = traveler.next
					temp.prev = traveler
					traveler.next = temp
					temp.index = traveler.index + 1
					if next != nil {
						next.prev = temp
						temp.next = next
					}
					// Split text
					origText := []rune(traveler.line)
					temp.line = string(origText[column:len(origText)])
					temp.length = len(temp.line)

					traveler.line = string(origText[0:column])
					traveler.length = len(traveler.line)
			}
			// TODO: Reindex all nodes on screen
			// TODO: Work with slices to make new ones with make
			sbReprintBuffer()
		}
	}

}
