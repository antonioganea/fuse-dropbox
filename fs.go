package main

import (
	"context"
	"flag"
	"log"
	"fmt"
	"syscall"
	"io"
	"path/filepath"

	"github.com/hanwen/go-fuse/fs"
	"github.com/hanwen/go-fuse/fuse"

	//"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	//"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

type HelloRoot struct {
	fs.Inode
}

//Alexandra

type virtualFile struct {
	io.Reader
	offset int
	Data []byte
}

func (x *virtualFile) Read(buf []byte) (n int, err error) {
	destLen := len(x.Data)
	if(x.offset < destLen) {
		copy(buf, x.Data)
		x.offset = destLen
		return destLen, nil
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

func (r *HelloRoot) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	//fmt.Println("nod creat: " + name)
	//fmt.Println(r)
	//newNode := Upload(ctx, &r.Inode, name)
	rootNode :=  &r.Inode
	fullPath := filepath.Join(rootNode.Path(nil), name)
	newNode := AddFile(ctx, rootNode, name, fullPath)

	return newNode, nil, 0, 0
}

func (r *HelloRoot) Write(ctx context.Context, fh fs.FileHandle, data []byte, offset int64) (uint32, syscall.Errno) {
	fmt.Println("helloroot - writing")
	return 0, 0
}

func (r *HelloRoot) Flush(ctx context.Context, f fs.FileHandle) syscall.Errno {
	fmt.Println("helloroot - flushed")
	return 0;
}

func (r *HelloRoot) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	fmt.Println("helloroot - fsynced")
	return 0;
}

func (r *HelloRoot) Allocate(ctx context.Context, f fs.FileHandle, off uint64, size uint64, mode uint32) syscall.Errno {
	fmt.Println("helloroot - allocated")
	return 0;
}

func (r *HelloRoot) Getlk(ctx context.Context, f fs.FileHandle, owner uint64, lk *fuse.FileLock, flags uint32, out *fuse.FileLock) syscall.Errno {
	fmt.Println("helloroot - getlk")
	return 0
}

func (r *HelloRoot) Setlk(ctx context.Context, f fs.FileHandle, owner uint64, lk *fuse.FileLock, flags uint32) syscall.Errno {
	fmt.Println("helloroot - setlk")
	return 0
}

func (r *HelloRoot) Setlkw(ctx context.Context, f fs.FileHandle, owner uint64, lk *fuse.FileLock, flags uint32) syscall.Errno {
	fmt.Println("helloroot - setlkw")
	return 0
}

var _ = (fs.NodeGetattrer)((*HelloRoot)(nil))
var _ = (fs.NodeOnAdder)((*HelloRoot)(nil))
var _ = (fs.NodeCreater)((*HelloRoot)(nil))
var _ = (fs.NodeWriter)((*HelloRoot)(nil))
var _ = (fs.NodeFlusher)((*HelloRoot)(nil))
var _ = (fs.NodeFsyncer)((*HelloRoot)(nil))
var _ = (fs.NodeAllocater)((*HelloRoot)(nil))
var _ = (fs.NodeGetlker)((*HelloRoot)(nil))
var _ = (fs.NodeSetlker)((*HelloRoot)(nil))
var _ = (fs.NodeSetlkwer)((*HelloRoot)(nil))


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
