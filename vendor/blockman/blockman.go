package blockman

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

/* Type Alias */
type Buffer = bytes.Buffer
type File = os.File
type Reader = bufio.Reader

type BlockManager interface {
	Read() ([]byte, error)
	Write([]byte) (int, error)
}

type BlockMan struct {
	File                     *File
	blockSize                int
	ammountReadInLoadedBlock int
	loadedBlock              []byte
	realBlockSize            int // in bytes
	workingBuffer            *Buffer
	TotalBytesRead           int64
	writingBuffer            *Buffer
	totalBytesWritten        int64
}

/* Custom Errors */
type BlockManError struct {
	message string
	hasFile bool
}

type IBlockManError interface {
	Error()
	HasFile()
}

func (e *BlockManError) Error() string {
	return fmt.Sprintf("%s", e.message)
}

func (e *BlockManError) HasFile() bool {
	return e.hasFile
}

/* "Constructor" */

func NewBlockMan(file *File) *BlockMan {
	var bm = &BlockMan{}
	bm.blockSize = os.Getpagesize() //100
	if file != nil {
		bm.File = file
	} else {
		bm.File = nil
	}
	bm.ammountReadInLoadedBlock = 0
	bm.loadedBlock = nil
	bm.realBlockSize = 0
	bm.workingBuffer = nil
	bm.writingBuffer = nil
	bm.TotalBytesRead = 0
	bm.totalBytesWritten = 0
	return bm
}

func (bm *BlockMan) Read() ([]byte, error) {
	if bm == nil {
		return nil, &BlockManError{"BlockMan object is nil.", false}
	}

	if bm.File == nil {
		return nil, &BlockManError{"No file in Block Manager.", false}
	}

	var (
		line []byte
		err  error
	)

	if bm.loadedBlock == nil {
		err = loadBlock(bm)
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
		close(bm) // munmap
		err = loadBlock(bm)
		if err == nil {
			line, err = readHelper(bm)
		}
	}
	return line, err
}

func (bm *BlockMan) Write(p []byte) (int, error) {

	// Will recieve bytes until it's full then it will write
	// if recv bytes are greater than its size, the remainder will be stored into the block
	// until either a subsequent write call writes the data or flush is called

	if bm.writingBuffer == nil {
		bm.writingBuffer = bytes.NewBuffer(make([]byte, 0))
	}

	//var bw int
	// Length available after write
	var bytesLeft int = (os.Getpagesize() - bm.writingBuffer.Len()) - len(p)
	if bytesLeft >= 0 {
		n, err := bm.writingBuffer.Write(p)
		if err != nil {
			return n, err
		}
		return n, nil
	} else {
		// We have a remainder only write part of if
		s, serr := bm.getFileSize()
		// Error getting the file size
		if serr != nil {
			return 0, serr
		}
		var (
			data []byte
			derr error
		)
		if (bm.totalBytesWritten + int64(os.Getpagesize())) <= s {
			data, derr = unix.Mmap(int(bm.File.Fd()), bm.totalBytesWritten, os.Getpagesize(),
				unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
		} else {
			// file needs to grow via ftruncate
			unix.Ftruncate(int(bm.File.Fd()), int64(bm.writingBuffer.Len()))

			data, derr = unix.Mmap(int(bm.File.Fd()), bm.totalBytesWritten, bm.writingBuffer.Len(),
				unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
		}

		if derr == nil {

			n := copy(data, bm.writingBuffer.Bytes()[:])
			// Sum what was written with what is remaining that will be flushed next
			bm.totalBytesWritten += int64(n)
			// underlying []byte will be emptied
			bm.writingBuffer.Reset()
			// write remainder
			unix.Msync(data, unix.MS_ASYNC|unix.MS_INVALIDATE)
			unix.Munmap(data)         // Close mmap
			bm.writingBuffer.Write(p) // write remainder
			return n, nil
		}
		return 0, derr
	}
}

func (bm *BlockMan) Flush() error {

	if bm.writingBuffer == nil {
		if bm.File != nil {
			return &BlockManError{"Nothing to Flush.", false}
		} else {
			return &BlockManError{"Nothing to Flush.", true}
		}
	}

	s, serr := bm.getFileSize()
	// Error getting the file size
	if serr != nil {
		return serr
	}
	var (
		data []byte
		derr error
	)
	if bm.writingBuffer.Len() > 0 {

		if (bm.totalBytesWritten + int64(os.Getpagesize())) <= s {
			data, derr = unix.Mmap(int(bm.File.Fd()), bm.totalBytesWritten, os.Getpagesize(),
				unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
		} else {
			// file needs to grow via ftruncate
			unix.Ftruncate(int(bm.File.Fd()), int64(bm.writingBuffer.Len()))

			data, derr = unix.Mmap(int(bm.File.Fd()), bm.totalBytesWritten, bm.writingBuffer.Len(),
				unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
		}

		if derr == nil {
			copy(data, bm.writingBuffer.Bytes()[:])
			bm.totalBytesWritten = 0
			bm.writingBuffer.Reset()
			unix.Msync(data, unix.MS_ASYNC|unix.MS_INVALIDATE)
			unix.Munmap(data) // Close mmap
			return nil
		}

		return derr
	}

	return nil
}

func (bm *BlockMan) BytesRead() int64 {
	return bm.TotalBytesRead
}

func loadBlock(bm *BlockMan) error {

	s, serr := bm.getFileSize()
	if serr != nil {
		return serr
	}
	var (
		data []byte
		derr error
	)

	if (bm.TotalBytesRead + int64(os.Getpagesize())) <= s {
		data, derr = unix.Mmap(int(bm.File.Fd()), bm.TotalBytesRead, os.Getpagesize(),
			unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	} else {
		data, derr = unix.Mmap(int(bm.File.Fd()), bm.TotalBytesRead, int(s-bm.TotalBytesRead),
			unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	}

	if derr == nil {
		//buffer := bytes.NewBuffer(make([]byte, bm.blockSize))
		bm.loadedBlock = data
		bm.realBlockSize = len(data)
		bm.TotalBytesRead += int64(len(data))
		bm.ammountReadInLoadedBlock = 0
		return nil
	}
	return derr
}

func readHelper(bm *BlockMan) ([]byte, error) {
	if bm.workingBuffer == nil || (bm.ammountReadInLoadedBlock == 0) {
		bm.workingBuffer = bytes.NewBuffer(bm.loadedBlock)
	}
	line, readingErr := bm.workingBuffer.ReadBytes('\n')
	if readingErr == nil {
		bm.ammountReadInLoadedBlock += len(line)
		return line, nil
	}
	if len(line) > 0 {
		bm.ammountReadInLoadedBlock += len(line)
		return line, readingErr
	}
	return nil, readingErr
}

func close(bm *BlockMan) {
	unix.Msync(bm.loadedBlock, unix.MS_ASYNC|unix.MS_INVALIDATE)
	unix.Munmap(bm.loadedBlock)
}

func (bm *BlockMan) getFileSize() (int64, error) {
	stat, err := bm.File.Stat()
	if err == nil {
		return stat.Size(), nil
	}
	return -1, err
}
