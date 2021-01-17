/*
	This file contains the core code, and also the main function.
	Execution starts here.
*/

package main

import (
	"context"
	"flag"
	_ "fmt"
	"io"
	"log"
	_ "path/filepath"
	"syscall"

	"github.com/hanwen/go-fuse/fs"
	"github.com/hanwen/go-fuse/fuse"
)

// We need a root inode structure that is mounted when the
// program starts.
type HelloRoot struct {
	DrpFileNode
}

// Virtualized file - it holds data but it's not a file
// on disk. It is stored in RAM.
type virtualFile struct {
	io.Reader
	offset int
	Data   []byte
}

// The virtualFile structure implements the io.Reader interface
// Therefore, we need to write an implementation for the Read
// function.
func (x *virtualFile) Read(buf []byte) (n int, err error) {
	destLen := len(x.Data)
	off := x.offset

	var readEnd int
	if destLen-off < destLen {
		readEnd = destLen
	} else {
		readEnd = off + destLen
	}

	x.offset = readEnd

	if off == readEnd {
		//cazul 1: mai avem date de output, in care dam return > 0
		return 0, io.EOF
	} else {
		//cazul 2: nu mai avem date, dai return 0, io.EOF
		copy(buf, x.Data[off:readEnd])
		return readEnd - off, nil
	}
}

// FUSE framework related function.
// It is ran when the root node is added to the tree
// But that means this function is executed when the
// root is mounted. - That moment is the start
// of the program.
func (r *HelloRoot) OnAdd(ctx context.Context) {
	ConstructTree(ctx, r)
}

// This returns 0777, giving all the permissions to the
// root of the filesystem. ( just in case )
func (r *HelloRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0777
	return 0
}

// In order to run the program, you need an empty folder.
// The command you should run in your terminal is :
// go run *.go <mount_directory>
// Where mount_directory is a path to an empty folder,
// where the virtual filesystem will be mounted.
// Warning : Your current working directory
// should be here ( in the same folder as the
// Go scripts ).
func main() {
	debug := flag.Bool("debug", false, "print debug data")
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  hello MOUNTPOINT")
	}
	opts := &fs.Options{}
	opts.Debug = *debug

	// This makes sure you have the dropbox account set up.
	initDbx()

	// This mounts the virtual filesystem.
	server, err := fs.Mount(flag.Arg(0), &HelloRoot{}, opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}

	// Infinite loop that serves requests
	server.Wait()
}

// Warning : After you kill the process,
// Unmount in case you have errors because of ghosts
// terminal command : fusermount -u <mount_directory>

// Go syntax. Links structures to interfaces.
var _ = (fs.NodeGetattrer)((*HelloRoot)(nil))
var _ = (fs.NodeOnAdder)((*HelloRoot)(nil))
