package main

import (
	"context"
	"flag"
	_"fmt"
	"io"
	"log"
	_"path/filepath"
	"syscall"

	"github.com/hanwen/go-fuse/fs"
	"github.com/hanwen/go-fuse/fuse"
	//"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	//"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
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
	if destLen - off < destLen {
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

// --------------- Preda check this comm pls:
// Creates a new child. It typically also returns a FileHandle as a
// reference for future reads/writes.
// Default is to return EROFS.
/*func (r *HelloRoot) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	// fmt.Println("nod creat: " + name)
	// fmt.Println(r)
	// newNode := Upload(ctx, &r.Inode, name)
	rootNode := &r.Inode
	fullPath := filepath.Join(rootNode.Path(nil), name)
	newNode := AddFile(ctx, rootNode, name, fullPath)

	return newNode, nil, 0, 0
}*/

// Creates a directory entry and Inode.
// Default is to return EROFS.
/*func (r *HelloRoot) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (node *fs.Inode, errno syscall.Errno) {
	rootNode := &r.Inode
	newNode := AddFolder(ctx, rootNode, name)

	return newNode, 0
}

func (r *HelloRoot) Write(ctx context.Context, fh fs.FileHandle, data []byte, offset int64) (uint32, syscall.Errno) {
	fmt.Println("helloroot - writing")
	return 0, 0
}

func (r *HelloRoot) Flush(ctx context.Context, f fs.FileHandle) syscall.Errno {
	fmt.Println("helloroot - flushed")
	return 0
}

func (r *HelloRoot) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	fmt.Println("helloroot - fsynced")
	return 0
}

func (r *HelloRoot) Allocate(ctx context.Context, f fs.FileHandle, off uint64, size uint64, mode uint32) syscall.Errno {
	fmt.Println("helloroot - allocated")
	return 0
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

func (r *HelloRoot) Mknod(ctx context.Context, name string, mode uint32, dev uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	fmt.Println("HelloRoot - mknod")
	return nil, 0
}

func (r *HelloRoot) Link(ctx context.Context, target fs.InodeEmbedder, name string, out *fuse.EntryOut) (node *fs.Inode, errno syscall.Errno) {
	fmt.Println("HelloRoot - link")
	return nil, 0
}*/

var _ = (fs.NodeGetattrer)((*HelloRoot)(nil))
var _ = (fs.NodeOnAdder)((*HelloRoot)(nil))
/*var _ = (fs.NodeCreater)((*HelloRoot)(nil))
var _ = (fs.NodeMkdirer)((*HelloRoot)(nil))
var _ = (fs.NodeWriter)((*HelloRoot)(nil))
var _ = (fs.NodeFlusher)((*HelloRoot)(nil))
var _ = (fs.NodeFsyncer)((*HelloRoot)(nil))
var _ = (fs.NodeAllocater)((*HelloRoot)(nil))
var _ = (fs.NodeGetlker)((*HelloRoot)(nil))
var _ = (fs.NodeSetlker)((*HelloRoot)(nil))
var _ = (fs.NodeSetlkwer)((*HelloRoot)(nil))
var _ = (fs.NodeMknoder)((*HelloRoot)(nil))
var _ = (fs.NodeLinker)((*HelloRoot)(nil))*/

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
