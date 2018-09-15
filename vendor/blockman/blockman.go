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
	ammountReadInLoadedBlock int
	loadedBlock []byte
	realBlockSize int // in bytes
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
	bm.ammountReadInLoadedBlock = 0
	bm.loadedBlock = nil
	bm.realBlockSize = 0
	return nil
}

func (bm *BlockMan) Read() ([]byte, error) {
	if bm == nil {
		return nil, &BlockManError{"BlockMan object is nil."}
	}

	var (
		line []byte
		err error
	)

	if bm.loadedBlock == nil {
		err = loadBlock(bm);
		// All good, block should be in bm.loadedBlock
		// Lets read from it
		if err == nil {
			line, err = readHelper(bm)
		}
	} else if bm.ammountReadInLoadedBlock < bm.realBlockSize {
		// Lets keep reading until block is read fully
		line, err = readHelper(bm)
	} else {
		// Read fully, lets load more bytes if any
		err = loadBlock(bm)
		if err == nil {
			line, err = readHelper(bm)
		}
	}
	return line, err
}

func loadBlock(bm *BlockMan) error {
	totalBytesRead := 0
	// A blockSize'd buffer
	buffer := bytes.NewBuffer(make([]byte, bm.blockSize))
	bytesRead, err := bm.fileRW.Read(buffer.Bytes())
	totalBytesRead += bytesRead

	if bytesRead == 0 && err == io.EOF {
		return err // Empty file
	} else {
		for totalBytesRead < bm.blockSize {
			bytesRead, err = bm.fileRW.Read(buffer.Bytes())
			totalBytesRead += bytesRead
			if bytesRead == 0 && err == io.EOF {
				break // reached EOF
			}
		}
		bm.loadedBlock = buffer.Bytes()
		bm.realBlockSize = totalBytesRead
		return nil
	}
}

func readHelper(bm *BlockMan) ([]byte, error) {
	buffer := bytes.NewBuffer(bm.loadedBlock)
	line, readingErr := buffer.ReadBytes('\n')
	if readingErr == nil {
		bm.ammountReadInLoadedBlock += len(line)
		return line, nil
	}
	if len(line) > 0 {
		return line, readingErr
	}
	return nil, readingErr
}






