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

type HelloRoot struct {
	DrpFileNode
}

type virtualFile struct {
	io.Reader
	offset int
	Data   []byte
}

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

func (r *HelloRoot) OnAdd(ctx context.Context) {
	ConstructTree(ctx, r)
}

func (r *HelloRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0777
	return 0
}

var _ = (fs.NodeGetattrer)((*HelloRoot)(nil))
var _ = (fs.NodeOnAdder)((*HelloRoot)(nil))

func main() {
	debug := flag.Bool("debug", false, "print debug data")
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  hello MOUNTPOINT")
	}
	opts := &fs.Options{}
	opts.Debug = *debug

	initDbx()

	// unmount in case you have errors because of ghosts
	server, err := fs.Mount(flag.Arg(0), &HelloRoot{}, opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}

	server.Wait()
}
