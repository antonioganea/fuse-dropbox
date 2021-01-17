/*
Generates Tree when first Downloaded
When first starting the tool, builder.go  
*/

package main

import (
	"fmt"

	"github.com/hanwen/go-fuse/fs"
	"github.com/hanwen/go-fuse/fuse"
	"golang.org/x/net/context"
)

var inoIterator uint64 = 2

func AddFile(ctx context.Context, node *fs.Inode, fileName string, fullPath string, modified bool) *fs.Inode {
	drpFileNode := DrpFileNode{}
	drpFileNode.drpPath = fullPath
	drpFileNode.modified = modified
	newfile := node.NewInode(
		ctx, &drpFileNode, fs.StableAttr{Ino: inoIterator})
	node.AddChild(fileName, newfile, false)

	inoIterator++

	return newfile
}

func AddFolder(ctx context.Context, node *fs.Inode, folderName string) *fs.Inode {
	dir := node.NewInode(
		ctx, &DrpFileNode{
			Data: []byte("sample dir data"),
			Attr: fuse.Attr{
				Mode: 0777,
			},
		}, fs.StableAttr{Ino: inoIterator, Mode: fuse.S_IFDIR})
	node.AddChild(folderName, dir, false)
	inoIterator++

	return dir
}

// Constructs the tree from our dropbox :)
func ConstructTreeFromDrpPaths(ctx context.Context, r *HelloRoot, structure []DrpPath) {

	var m map[string](*fs.Inode) = make(map[string](*fs.Inode))

	m[""] = &r.Inode

	fmt.Println("Constructing tree")
	for _, entry := range structure {
		fmt.Println("Processing : " + entry.path)

		var containingFolder = firstPartFromPath(entry.path) // "/dirA" -> ""
		var newNodeName = lastFolderFromPath(entry.path)     // 		-> "dirA"

		fmt.Printf("containing folder : %v, newNodeName : %v \n", containingFolder, newNodeName)

		var parentNode = m[containingFolder]
		var newNode *fs.Inode
		if entry.isFolder {
			newNode = AddFolder(ctx, parentNode, newNodeName)
		} else {
			newNode = AddFile(ctx, parentNode, newNodeName, entry.path, false)
		}

		m[containingFolder+"/"+newNodeName] = newNode

		fmt.Println("Mapped the newly created node in " + containingFolder + "/" + newNodeName)
	}
}

func ConstructTree(ctx context.Context, r *HelloRoot) {
	ConstructTreeFromDrpPaths(ctx, r, getDropboxTreeStructure())
}
