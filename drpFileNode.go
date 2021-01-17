/*
	This file contains the definition for a DrpFileNode.
	That is, an extended fs.Indode.

	Essentialy, because we're programming a custom
	filesystem, we have to register custom functionality
	for the common filesystem usage events such as
	Open, Setattr, Getattr, Read, Write, Flush, etc...
*/

package main

import (
	"fmt"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/hanwen/go-fuse/fs"
	"github.com/hanwen/go-fuse/fuse"
	"golang.org/x/net/context"
)

type DrpFileNode struct {
	// Must embed an Inode for the struct to work as a node.
	fs.Inode

	// drpPath is the path of this file/directory
	drpPath string

	mu       sync.Mutex
	Data     []byte
	Attr     fuse.Attr
	modified bool
}

// This is called by the kernel when it needs file meta attributes.
func (bn *DrpFileNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	//bn.mu.Lock()
	//defer bn.mu.Unlock()
	//fmt.Println("DrpFileNode - getattr")

	if bn.Inode.IsDir() {
		out.Attr = bn.Attr
		out.Attr.Size = uint64(len(bn.Data))
		return fs.OK
	}

	dbx := files.New(config)

	downloadArg := files.NewDownloadArg(bn.drpPath)

	meta, _, err := dbx.Download(downloadArg)
	if err != nil {
		return 404
	}

	out.Size = meta.Size
	out.Mode = 0777

	return fs.OK
}

// This is called when the kernel tries to read from the file.
func (drpn *DrpFileNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	// drpn.mu.Lock()
	// defer drpn.mu.Unlock()
	fmt.Println("DrpFileNode - Read")

	destLen := int64(len(dest))

	dbx := files.New(config)

	downloadArg := files.NewDownloadArg(drpn.drpPath)

	meta, content, err := dbx.Download(downloadArg)
	if err != nil {
		return nil, 404
	}
	if off == int64(meta.Size) {
		return fuse.ReadResultData(make([]byte, 0)), 0
	}

	b1 := make([]byte, meta.Size)
	n1, err := content.Read(b1)

	fmt.Println(string(b1[:n1]))

	var readEnd int64
	if int64(meta.Size)-off < destLen {
		readEnd = int64(meta.Size)
	} else {
		readEnd = off + destLen
	}

	return fuse.ReadResultData(b1[off:readEnd]), 0
}

// This is called when the kernel tries to open the file.
func (f *DrpFileNode) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	fmt.Println("DrpFileNode - open")
	return new(fs.FileHandle), fuse.FOPEN_KEEP_CACHE, fs.OK
}

// This is called when the kernel tries to write to the file.
func (drpn *DrpFileNode) Write(ctx context.Context, fh fs.FileHandle, data []byte, off int64) (uint32, syscall.Errno) {
	fmt.Println("DrpFileNode - writing")
	drpn.mu.Lock()
	defer drpn.mu.Unlock()

	end := int64(len(data)) + off
	if int64(len(drpn.Data)) < end {
		n := make([]byte, end)
		copy(n, drpn.Data)
		drpn.Data = n
	}

	copy(drpn.Data[off:off+int64(len(data))], data)
	drpn.modified = true

	return uint32(len(data)), 0
}

func (drpn *DrpFileNode) Flush(ctx context.Context, f fs.FileHandle) syscall.Errno {
	fmt.Println("DrpFileNode - flushed")
	if drpn.modified == true {
		path, parent := drpn.Inode.Parent()

		fileName := lastFolderFromPath(path)
		fullPath := filepath.Join(parent.Path(nil), fileName)
		Upload(fullPath, drpn.Data)
		drpn.modified = false
	}
	return 0
}

func (drpn *DrpFileNode) Fsync(ctx context.Context, f fs.FileHandle, flags uint32) syscall.Errno {
	fmt.Println("DrpFileNode - fsynced")
	return 0
}

func (drpn *DrpFileNode) Allocate(ctx context.Context, f fs.FileHandle, off uint64, size uint64, mode uint32) syscall.Errno {
	fmt.Println("DrpFileNode - allocated")
	return 0
}

func (drpn *DrpFileNode) Getlk(ctx context.Context, f fs.FileHandle, owner uint64, lk *fuse.FileLock, flags uint32, out *fuse.FileLock) syscall.Errno {
	fmt.Println("DrpFileNode - getlk")
	return 0
}

func (drpn *DrpFileNode) Setlk(ctx context.Context, f fs.FileHandle, owner uint64, lk *fuse.FileLock, flags uint32) syscall.Errno {
	fmt.Println("DrpFileNode - setlk")
	return 0
}

func (drpn *DrpFileNode) Setlkw(ctx context.Context, f fs.FileHandle, owner uint64, lk *fuse.FileLock, flags uint32) syscall.Errno {
	fmt.Println("DrpFileNode - setlkw")
	return 0
}

// This is called by the kernel when it sets file meta attributes.
func (drpn *DrpFileNode) Setattr(ctx context.Context, f fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	fmt.Println("DrpFileNode - setattr")
	drpn.mu.Lock()
	defer drpn.mu.Unlock()

	if sz, ok := in.GetSize(); ok {
		drpn.Data = drpn.Data[:sz]
	}
	out.Attr = drpn.Attr
	out.Size = uint64(len(drpn.Data))
	return 0
}

// This is called upon creation.
func (drpn *DrpFileNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	fmt.Println("nod creat: " + name)
	rootNode := &drpn.Inode
	fullPath := "/" + filepath.Join(rootNode.Path(nil), name)
	newNode := AddFile(ctx, rootNode, name, fullPath, true)

	return newNode, nil, 0, 0
}

func (drpn *DrpFileNode) Mknod(ctx context.Context, name string, mode uint32, dev uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	fmt.Println("DrpFileNode - mknod")
	return nil, 0
}

func (drpn *DrpFileNode) Link(ctx context.Context, target fs.InodeEmbedder, name string, out *fuse.EntryOut) (node *fs.Inode, errno syscall.Errno) {
	fmt.Println("DrpFileNode - link")
	return nil, 0
}

// This is called when a directory is created.
func (drpn *DrpFileNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (node *fs.Inode, errno syscall.Errno) {
	fmt.Println("DrpFileNode - Mkdir")
	rootNode := &drpn.Inode
	newNode := AddFolder(ctx, rootNode, name)

	fullPath := "/" + filepath.Join(rootNode.Path(nil), name)
	UploadFolder(fullPath)

	return newNode, 0
}

// This is called when a file is deleted.
func (drpn *DrpFileNode) Unlink(ctx context.Context, name string) syscall.Errno {
	fmt.Println("DrpFileNode - Unlink")
	RemoteDelete(drpn, name)
	return 0
}

// This is called when a folder is deleted.
func (drpn *DrpFileNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	fmt.Println("DrpFileNode - Rmdir")
	RemoteDelete(drpn, name)
	return 0
}

// These are part of the Go syntax.
// What they do is that they make sure
// that Go understands the fact that
// the DrpFileNode structure implements
// the following INTERFACES.
var _ = (fs.NodeOpener)((*DrpFileNode)(nil))
var _ = (fs.NodeReader)((*DrpFileNode)(nil))
var _ = (fs.NodeWriter)((*DrpFileNode)(nil))
var _ = (fs.NodeFlusher)((*DrpFileNode)(nil))
var _ = (fs.NodeFsyncer)((*DrpFileNode)(nil))
var _ = (fs.NodeAllocater)((*DrpFileNode)(nil))
var _ = (fs.NodeGetlker)((*DrpFileNode)(nil))
var _ = (fs.NodeSetlker)((*DrpFileNode)(nil))
var _ = (fs.NodeSetlkwer)((*DrpFileNode)(nil))
var _ = (fs.NodeSetlkwer)((*DrpFileNode)(nil))
var _ = (fs.NodeSetattrer)((*DrpFileNode)(nil))
var _ = (fs.NodeGetattrer)((*DrpFileNode)(nil))
var _ = (fs.NodeCreater)((*DrpFileNode)(nil))
var _ = (fs.NodeMknoder)((*DrpFileNode)(nil))
var _ = (fs.NodeLinker)((*DrpFileNode)(nil))
var _ = (fs.NodeMkdirer)((*DrpFileNode)(nil))
var _ = (fs.NodeUnlinker)((*DrpFileNode)(nil))
var _ = (fs.NodeRmdirer)((*DrpFileNode)(nil))
