package screenbuf

import (
	"fmt"
	"easyterm"
	"bufio"
	"io"
	"bytes"
	"os"
	"blockman"
	"strings"
	//"strconv"
)
type Buffer = bytes.Buffer
type File = os.File
type ReadWriter = bufio.ReadWriter
type BlockMan = blockman.BlockMan

// Consts represent what to do with the
// content when the user presses enter on a line
// depending on the column index
const TAB_SPACE int = 8

const (
	UP = iota
	SPLIT = iota
	DOWN = iota
)

type BufferNode struct {
	Index int
	Line string
	RealLine string
	Length int
	Next *BufferNode
	Prev *BufferNode
}

type ScreenBuffer struct {
	Head *BufferNode
	Length int
	MaxSizeInBytes int
	FilePtr *File
	IndexOfFirstVisibleLine int
	Blockman *BlockMan
	DefaultHeight int
	DefaultWidth int
	TabFiller string
	TabSpace int
}

func NewScreenBuffer(file *File)(*ScreenBuffer) {
	var sb = &ScreenBuffer{}
	sb.Length = 0;
	sb.FilePtr = file
	// Init blockman
	var bm = blockman.NewBlockMan(sb.FilePtr)
	sb.Blockman = bm
	sb.IndexOfFirstVisibleLine = 1
	sb.DefaultHeight = 25
	sb.DefaultWidth = 80

	// get current window dimensions
	if w, h, err := easyterm.GetSize(); err == nil {
		sb.DefaultHeight = h
		sb.DefaultWidth = w
	}

	sb.TabSpace = TAB_SPACE
	for i:=0; i < TAB_SPACE - 1; i++ {
		sb.TabFiller += " "		
	}

	return sb
}

func (buffer *ScreenBuffer) LoadFile() {
	// Lets load the file
	var temp = &BufferNode{}
	var traveler = &BufferNode{}

	for i := 1; i < buffer.DefaultHeight; i ++ {
		lineBytes, err := sbReadLine(buffer.Blockman); // Send blockmanager to read
		//easyterm.CursorPos(1,1)
		//fmt.Print(string(lineBytes))
		//fmt.Print(err)
		//os.Exit(1)
		// already at the end
		if err == io.EOF && buffer.Length == 0 {
			temp.Index = i
			temp.Line = ""
			temp.RealLine = ""
			temp.Length = len(temp.RealLine)
			temp.Prev = nil
			temp.Next = nil
			buffer.Head = temp
			buffer.Length++;
			break
		}
		// Not at EOF
		if err == nil {
			if i == 1 {
				temp.Index = i
				temp.Line = strings.Trim(string(lineBytes), "\n")
				temp.RealLine = packTabs(buffer, temp.Line) // pad with 8 spaces the line //"\t       "
				temp.Length = len(temp.RealLine)
				temp.Prev = nil
				temp.Next = nil
				buffer.Head = temp
				buffer.Length++;
				traveler = buffer.Head
			} else {
				temp = &BufferNode{}
				temp.Index = i
				temp.Line = strings.Trim(string(lineBytes), "\n")
				temp.RealLine = packTabs(buffer, temp.Line) // pad with 8 spaces the line
				temp.Length = len(temp.RealLine)
				temp.Prev = traveler
				temp.Next = nil

				traveler.Next = temp
				traveler = traveler.Next
				buffer.Length++
			}
		}
	}

}

func (sb *ScreenBuffer) LoadLine(fromWhere int) {
	switch fromWhere {
	case UP:

	case DOWN:
		line, err := sbReadLine(sb.Blockman)
		if err == nil {
			sbEnqueueLine(sb,line, DOWN)
		}
	}
}

func (buffer *ScreenBuffer) PrintBuffer() {
	var i int = 1
	for traveler := buffer.Head; traveler != nil; traveler = traveler.Next {
		fmt.Printf("%s\n", traveler.Line)
		i++
		easyterm.CursorPos(i,1)
	}
	// Move cursor to next line
	easyterm.CursorPos(i+1,1)

	// Fill the rest of the screen with ~ if some space available
	for x:= i; x < buffer.DefaultHeight; x++ {
		fmt.Printf("~\n")
		easyterm.CursorPos(x,1)
	}
	easyterm.CursorPos(1,1)
	//fmt.Println("lenght: " + strconv.Itoa(buffer.length))
}

func (buffer *ScreenBuffer) GetLineLength(line int) int {
	i := 1
	for traveler := buffer.Head; traveler != nil; traveler = traveler.Next {
		if traveler.Index == line {
			i = traveler.Length
			break
		}
	}
	return i
}

func (buffer *ScreenBuffer) GetBufferLength() int {
	size := 0
	for traveler := buffer.Head; traveler != nil; traveler = traveler.Next {
		size++
	}
	return size
	//return buffer.length
}

func (buffer *ScreenBuffer) UpdateBufferIndexes() {
	i := 1
	for traveler := buffer.Head; traveler != nil; traveler = traveler.Next {
		traveler.Index = i
		i++
	}
}

func (buffer *ScreenBuffer) PrintLine(line int) {
	for traveler := buffer.Head; traveler != nil; traveler = traveler.Next {
		if traveler.Index == line {
			fmt.Print(traveler.Line)
			break
		}
	}
}

func (buffer *ScreenBuffer) GetLine(line int) *BufferNode {
	var lineNode *BufferNode
	for traveler := buffer.Head; traveler != nil; traveler = traveler.Next {
		if traveler.Index == line {
			lineNode = traveler
			break
		}
	}
	return lineNode
}

func (buffer *ScreenBuffer) ReprintBuffer()  {
	i := 1
	for traveler := buffer.Head; traveler != nil; traveler = traveler.Next {
		easyterm.CursorPos(i,1)
		easyterm.ClearLine()
		easyterm.CursorPos(i,1)
		if traveler.Next != nil {
			fmt.Printf("%s\n", traveler.Line)
		} else {
			fmt.Printf("%s", traveler.Line)
		}
		//fmt.Printf("%d\n", traveler.index)
		i++
	}

	for x:= i; x < buffer.DefaultHeight; x++ {
		easyterm.CursorPos(i,1)
		easyterm.ClearLine()
		easyterm.CursorPos(i,1)
		fmt.Printf("~\n")
		easyterm.CursorPos(x,1)
	}

}

func (buffer *ScreenBuffer) AddLineToBuffer(line, column int) {

	for traveler := buffer.Head; traveler != nil; traveler = traveler.Next {
		if traveler.Index == line {

			var temp = &BufferNode{}
			temp.Line = ""
			temp.Length = 0
			// col - 1 because screen is 1 based and strings are 0 based
			//fmt.Printf("%v\n", temp.prev)
			var insertWhere = manageNewLineString(column, traveler.Length)
			//fmt.Printf("%d\n",insertWhere)
			switch insertWhere {
				case UP:
					var prev *BufferNode
					prev = traveler.Prev
					//temp.index = traveler.index
					temp.Next = traveler
					if prev != nil {
						temp.Prev = prev
						prev.Next = temp
					} else {
						temp.Prev = nil
					}
					traveler.Prev = temp
					if traveler.Index == 1 {
						buffer.Head = temp
					}
					//traveler.index += 1
					buffer.Length += 1
				case DOWN:
					var next *BufferNode
					next = traveler.Next
					temp.Prev = traveler
					//temp.index = traveler.index + 1
					if next != nil {
						next.Prev = temp
						temp.Next = next
					}
					traveler.Next = temp
					buffer.Length += 1
				case SPLIT:
					var next *BufferNode
					next = traveler.Next
					temp.Prev = traveler
					if next != nil {
						next.Prev = temp
						temp.Next = next
					}
					traveler.Next = temp
					buffer.Length += 1
					// Split text
					origText := make([]rune, traveler.Length)
					// Copy original text to origText
					copy(origText, []rune(traveler.RealLine))
					// New original lines text
					origRune := make([]rune, len(origText[0:(column - 1)]))
					// Copying original lines new text to rune
					copy(origRune, origText[0:column - 1])
					// slice with new line text
					newRune := origText[(column - 1):traveler.Length]
					newText := make([]rune, len(newRune))
					copy(newText, newRune)
					// set the string on the new line
					temp.Line = unpackTabs(buffer, string(newText))
					temp.RealLine = packTabs(buffer, temp.Line)
					temp.Length = len(temp.RealLine)

					// update the old lines text
					traveler.Line = unpackTabs(buffer, string(origRune))
					traveler.RealLine = packTabs(buffer, traveler.Line)
					traveler.Length = len(traveler.RealLine)

					/*traveler.line = string(origText[0:column])
					traveler.length = len(traveler.line)*/
			}

			buffer.UpdateBufferIndexes()
			buffer.ReprintBuffer()
			//easyterm.CursorPos(line,column)
		}
	}

}

func manageNewLineString(col, length int) int {
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

func packTabs(sb *ScreenBuffer, line string) string {
	return strings.Replace(line, "\t", "\t" + sb.TabFiller, -1)
}

func unpackTabs(sb *ScreenBuffer, line string) string {
	return strings.Replace(line, "\t" + sb.TabFiller, "\t", -1)
}

func sbReadLine(bm *BlockMan) ([]byte, error) {
	return bm.Read()
}

func sbEnqueueLine(buffer *ScreenBuffer, line []byte, where int) {
	// add line via reading or add line via enter
	traveler := buffer.GetLine(buffer.Length)
	switch where {
	case UP:

	case DOWN:
		var temp = &BufferNode{}
		temp.Index = buffer.Length + 1
		temp.Line = string(line)
		temp.Length = len(temp.Line)
		temp.RealLine = strings.Replace(temp.Line, "\t", "\t" + buffer.TabFiller, -1)
		temp.Prev = traveler
		temp.Next = nil

		if traveler.Next == nil {
			traveler.Next = temp
		}
		buffer.Length++
	}
	// Call function that handles showing data on screen, hiding/moving lines when sb is too large
}

func sbDequeueLine() {

}