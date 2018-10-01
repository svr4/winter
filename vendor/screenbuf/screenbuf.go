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
	"unsafe"
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
	MaxSizeInBytes int64
	FilePtr *File
	IndexOfFirstVisibleLine int
	IndexOfLastVisisbleLine int
	Blockman *BlockMan
	DefaultHeight int
	DefaultWidth int
	TabFiller string
	TabSpace int
	TabStops []int
}

func NewScreenBuffer(file *File)(*ScreenBuffer) {
	var sb = &ScreenBuffer{}
	sb.Length = 0;
	sb.FilePtr = file

	// Set the max size in bytes to a third of the size of the file or OS page size
	if fileInfo, errr := sb.FilePtr.Stat(); errr == nil && fileInfo.Size() > 0 {

		if gigs := fileInfo.Size() / 1000000000; gigs > 0 {
			sb.MaxSizeInBytes = int64(gigs / 8) // only one eight of the file
		} else if megs := fileInfo.Size() / 1000000; megs > 0  { // size in megs or lower
			sb.MaxSizeInBytes = int64(megs / 6)
		} else if kilos := fileInfo.Size() / 1000; kilos > 0 {
			sb.MaxSizeInBytes = int64(kilos / 4)
		} else {
			sb.MaxSizeInBytes = int64(os.Getpagesize())
		}

	} else {
		sb.MaxSizeInBytes = int64(os.Getpagesize())
	}

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
	var numOfStops int = (sb.DefaultWidth / sb.TabSpace)
	sb.TabStops = make([]int, numOfStops)

	var tabIdx int = 0
	for ii:= 1; ii <= sb.DefaultWidth; ii++ {
		if ii % sb.TabSpace == 0 {
			if tabIdx < len(sb.TabStops) {
				sb.TabStops[tabIdx] = ii
				tabIdx++
			} else {
				break;
			}
		}
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
				temp.RealLine = buffer.PackTabs(temp.Line) // pad with 8 spaces the line //"\t       "
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
				temp.RealLine = buffer.PackTabs(temp.Line) // pad with 8 spaces the line
				temp.Length = len(temp.RealLine)
				temp.Prev = traveler
				temp.Next = nil

				traveler.Next = temp
				traveler = traveler.Next
				buffer.Length++
			}
		} else if err == io.EOF && len(lineBytes) > 0 {
			temp = &BufferNode{}
			temp.Index = i
			temp.Line = ""
			for err == io.EOF && len(lineBytes) > 0 {
				temp.Line += string(lineBytes)
				lineBytes, err = sbReadLine(buffer.Blockman);
			}
			if err == nil && len(lineBytes) > 0 {
				temp.Line += string(lineBytes)
			}
			temp.Line = strings.Trim(temp.Line, "\n")
			temp.RealLine = buffer.PackTabs(temp.Line) // pad with 8 spaces the line
			temp.Length = len(temp.RealLine)
			temp.Prev = traveler
			temp.Next = nil

			traveler.Next = temp
			traveler = traveler.Next
			buffer.Length++
		}
	}
	buffer.IndexOfLastVisisbleLine = buffer.Length
}

func (sb *ScreenBuffer) LoadLine(fromWhere int, currentLineIndex int) {
	currentLine := sb.GetLine(currentLineIndex)

	switch fromWhere {
	case UP:

		if currentLine.Prev != nil {
			screenUpReAdjustment(sb)
		}


	case DOWN:

		/*if sb.screenBuffByteSize() >= uintptr(sb.MaxSizeInBytes) {
			// need to move invisible top lines to the .~filename
		}*/

		if currentLine.Next == nil {
			line, err := sbReadLine(sb.Blockman)
			if err == nil || (err == io.EOF && len(line) > 0) {
				sbEnqueueLine(sb,line, DOWN)
				screenDownReAdjustment(sb)
			}
		} else {
			screenDownReAdjustment(sb)
		}
	}
}

func (buffer *ScreenBuffer) PrintBuffer() {
	var i int = 1
	easyterm.ShowCursor(false)
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
	easyterm.ShowCursor(true)
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

func (buffer *ScreenBuffer) Size() int {
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
	easyterm.ShowCursor(false)
	for traveler := buffer.GetLine(buffer.IndexOfFirstVisibleLine);
	traveler != nil && traveler.Index <= buffer.IndexOfLastVisisbleLine; traveler = traveler.Next {
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
	easyterm.ShowCursor(true)
}

func (buffer *ScreenBuffer) AddLineToBuffer(line, column int) {

	for traveler := buffer.Head; traveler != nil; traveler = traveler.Next {
		if traveler.Index == line {

			var temp = &BufferNode{}
			temp.Line = ""
			temp.RealLine = ""
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
					temp.Line = buffer.UnpackTabs(string(newText))
					temp.RealLine = buffer.PackTabs(temp.Line)
					temp.Length = len(temp.RealLine)

					// update the old lines text
					traveler.Line = buffer.UnpackTabs(string(origRune))
					traveler.RealLine = buffer.PackTabs(traveler.Line)
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

func (sb *ScreenBuffer) NextTabStop(index int) int {
	var tabStopsLen = len(sb.TabStops)
	var nextStop int = 0
	for i:=0; i < tabStopsLen; i ++ {
		if (i + 1) < tabStopsLen {
			if index < sb.TabStops[i] {
				nextStop = sb.TabStops[i]
				break
			} else if index == sb.TabStops[i] {
				nextStop = sb.TabStops[i+1]
				break
			} else if index > sb.TabStops[i] && index < sb.TabStops[i + 1] {
				nextStop = sb.TabStops[i + 1]
				break
			}
		}
	}
	return nextStop
}

func (sb *ScreenBuffer) PrevTabStop(index int) int {
	var tabStopsLen = len(sb.TabStops)
	var prevStop int = 0
	for i:=tabStopsLen - 1; i >= 0; i-- {
		// Could be used to better the NextTabStop search
		if (i - 1) >= 0 {
			if index >= sb.TabStops[i] {
				prevStop = sb.TabStops[i - 1]
				break
			}
		}
	}
	return prevStop
}

func (sb *ScreenBuffer) IsATabStop(index int) bool {
	for i:=0; i < len(sb.TabStops); i++ {
		if sb.TabStops[i] == index {
			return true
		}
	}
	return false
}

func (sb *ScreenBuffer) PackTabs(line string) string {
	var filler string = ""
	var nextStop int
	origString := make([]rune, len(line))
	copy(origString, []rune(line))

	var workingFiller []rune
	var firstHalf, secondHalf []rune
	
	for i:=0; i < len(origString); i++ {
		if origString[i] == '\t' {

			nextStop = sb.NextTabStop(i)

			for ii:=i; ii < nextStop - 1; ii++ {
				filler += " "
			}
			
			workingFiller = make([]rune, len(filler))
			copy(workingFiller, []rune(filler))

			firstHalf = make([]rune, len(origString[0:i]))
			copy(firstHalf, origString[0:i])

			if (i + 1) < len(origString) {
				secondHalf = make([]rune, len(origString[i+1:len(origString)]))
				copy(secondHalf, origString[i+1:len(origString)])
			} else {
				secondHalf = make([]rune, 0)
			}
			
			newLine := string(firstHalf) + "\t" + string(workingFiller) + string(secondHalf)
			origString = make([]rune, len(newLine))
			copy(origString, []rune(newLine))
			
			i += len(workingFiller)
			filler = ""
		}
	}
	return string(origString)
	
}

func (sb *ScreenBuffer) UnpackTabs(line string) string {
	var nextStop int
	origString := make([]rune, len(line))
	copy(origString, []rune(line))

	var firstHalf, secondHalf []rune

	for i:=len(origString) - 1; i >= 0 ; i-- {
		if origString[i] == '\t' {
			
			nextStop = sb.NextTabStop(i)
			
			if nextStop > len(origString) || len(origString[i:nextStop]) == 0 {
				continue
			} else {
				var tabFill []rune = origString[i+1:nextStop]
				if (!strings.Contains(string(tabFill), " ")){
					continue
				}
			}

			firstHalf = make([]rune, len(origString[0:i+1]))
			copy(firstHalf, origString[0:i+1])

			secondHalf = make([]rune, len(origString[nextStop:len(origString)]))
			copy(secondHalf, origString[nextStop:len(origString)])

			newLine := string(firstHalf) + string(secondHalf)
			origString = make([]rune, len(newLine))

			copy(origString, []rune(newLine))
		}
	}
	return string(origString)
	
}

/* Private functions */

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
		temp.RealLine = buffer.PackTabs(temp.Line)
		temp.Length = len(temp.RealLine)
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

func hasNodeAtIndex(sb *ScreenBuffer, index int) bool {
	var containsIndex bool = false
	for traveler := sb.Head; traveler != nil; traveler = traveler.Next {
		if traveler.Index == index {
			containsIndex = true
			break
		}
	}
	return containsIndex
}

func (sb *ScreenBuffer) screenBuffByteSize () uintptr {
	var size uintptr = 0
	for traveler := sb.Head; traveler != nil; traveler = traveler.Next {
		size += unsafe.Sizeof(traveler.RealLine)
	}
	return size
}

func reprintBufferWindow(sb *ScreenBuffer) {
	i := 1
	easyterm.ShowCursor(false)
	easyterm.Clear()
	for traveler := sb.GetLine(sb.IndexOfFirstVisibleLine);
	traveler != nil && traveler.Index <= sb.IndexOfLastVisisbleLine; traveler = traveler.Next {
		easyterm.CursorPos(i,1)
		if traveler.Next != nil {
			fmt.Printf("%s\n", traveler.Line)
		} else {
			fmt.Printf("%s", traveler.Line)
		}		
		i++
	}
	easyterm.ShowCursor(true)
}

func screenDownReAdjustment(sb *ScreenBuffer) {
	// If this is called we know we want the screen to move up one line
	// by changing the id for the first visible line to the next id down
	// and reprinting the file
	var firstNodeIndex int = sb.IndexOfFirstVisibleLine + 1
	var lastNodeIndex = sb.IndexOfLastVisisbleLine + 1
	if hasNodeAtIndex(sb, firstNodeIndex) {
		sb.IndexOfFirstVisibleLine = firstNodeIndex
	}
	if hasNodeAtIndex(sb, lastNodeIndex) {
		sb.IndexOfLastVisisbleLine = lastNodeIndex
	}
	reprintBufferWindow(sb)

	// check that buffer is small enough. If large move top lines to temp file ~.filename
	// and load lines later from there

}

func screenUpReAdjustment(sb *ScreenBuffer) {
	var firstNodeIndex = sb.IndexOfFirstVisibleLine - 1
	var lastNodeIndex = sb.IndexOfLastVisisbleLine - 1
	if hasNodeAtIndex(sb, firstNodeIndex) {
		sb.IndexOfFirstVisibleLine = firstNodeIndex
	}
	if hasNodeAtIndex(sb, lastNodeIndex) {
		sb.IndexOfLastVisisbleLine = lastNodeIndex
	}
	reprintBufferWindow(sb)
}