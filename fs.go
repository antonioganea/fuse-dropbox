package main

import (
	"context"
	"flag"
	"log"
	"syscall"
	"io"

	"github.com/hanwen/go-fuse/fs"
	"github.com/hanwen/go-fuse/fuse"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

type HelloRoot struct {
	fs.Inode
}

//Alexandra

type forUpdate struct {
	x io.Reader
	offset int
}

func (x *forUpdate) Read(buf []byte)(n int, err error) {
	if(x.offset < 4) {
		buf[0] = 's'
		buf[1] = 't'
		buf[2] = 'r'
		buf[3] = 0;
		x.offset = 4
		return 4, nil
	}
	return 0, io.EOF
}


func (r *HelloRoot) OnAdd(ctx context.Context) {
	ConstructTree(ctx, r)
}

func (r *HelloRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755
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

	s := new(files.CommitInfo)
	s.Path = "/so.txt"
	s.Mode = &files.WriteMode{Tagged: dropbox.Tagged{"add"}}
	s.Autorename = false
	s.Mute = false
	s.StrictConflict = false

	t := new(forUpdate)
	t.offset = 0
	dbx := files.New(config)
	dbx.Upload(s, t)

	server.Wait()
}
