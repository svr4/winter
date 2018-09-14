package blockman

import (
	"bufio"
	"bytes"
	"os"
	"fmt"
	"io"
)

/* Type Alias */
type Buffer = bytes.Buffer
type File = os.File
type ReadWriter = bufio.ReadWriter

type BlockManager interface {
	Read() ([]byte, error)
}

type BlockMan struct {
	file *File
	fileRW *ReadWriter
	blockSize int
}
/* Custom Errors */
type BlockManError struct {
	message string
}

func (e *BlockManError) Error() string {
	return fmt.Sprintf("%s", e.message)
}

/* "Constructor" */

func (bm *BlockMan) Init(file *File) error {
	if bm == nil {
		return &BlockManError{"BlockMan object is nil."}
	}
	bm.file = file
	bm.blockSize = os.Getpagesize()
	bm.fileRW = bufio.NewReadWriter(bufio.NewReader(bm.file), bufio.NewWriter(file))
	return nil
}

func (bm *BlockMan) Read() ([]byte, error) {
	if bm == nil {
		return nil, &BlockManError{"BlockMan object is nil."}
	}

	totalBytesRead := 0
	// A blockSize'd buffer
	buffer := bytes.NewBuffer(make([]byte, bm.blockSize))
	bytesRead, err := bm.fileRW.Read(buffer.Bytes())
	totalBytesRead += bytesRead

	if bytesRead == 0 && err == io.EOF {
		return nil, err
	} else {
		for totalBytesRead < bm.blockSize {
			bytesRead, err = bm.fileRW.Read(buffer.Bytes())
			totalBytesRead += bytesRead
			if bytesRead == 0 && err == io.EOF {
				return buffer.Bytes(), err
			}
		}
		return buffer.Bytes(), nil
	}
}






