package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/hanwen/go-fuse/fs"
	"github.com/hanwen/go-fuse/fuse"
	"golang.org/x/net/context"
)

const (
	configFileName = "AccessToken"
)

var config dropbox.Config

type DrpPath struct {
	path     string
	isFolder bool
}

func validatePath(p string) (path string, err error) {
	path = p

	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	path = strings.TrimSuffix(path, "/")

	return
}

// Returns a list of paths that represent the DFS traversal
// of the drobpox folder.
func getDropboxTreeStructure() []DrpPath {
	dbx := files.New(config)

	arg := files.NewListFolderArg("")
	arg.Recursive = true

	res, err := dbx.ListFolder(arg)
	var entries []files.IsMetadata
	if err != nil {
		switch e := err.(type) {
		case files.ListFolderAPIError:
			if e.EndpointError.Path.Tag == files.LookupErrorNotFolder {
				var metaRes files.IsMetadata
				metaRes, err = getFileMetadata(dbx, "/")
				entries = []files.IsMetadata{metaRes}
			} else {
				return nil
			}
		default:
			return nil
		}

		if err != nil {
			return nil
		}
	} else {
		entries = res.Entries

		for res.HasMore {
			arg := files.NewListFolderContinueArg(res.Cursor)

			res, err = dbx.ListFolderContinue(arg)
			if err != nil {
				return nil
			}

			entries = append(entries, res.Entries...)
		}
	}

	structure := make([]DrpPath, 0)

	for _, entry := range entries {
		switch f := entry.(type) {
		case *files.FileMetadata:
			fmt.Println(f.PathDisplay)

			var node DrpPath
			node.path = f.PathDisplay
			node.isFolder = false

			structure = append(structure, node)

		case *files.FolderMetadata:
			fmt.Println(f.PathDisplay)

			var node DrpPath
			node.path = f.PathDisplay
			node.isFolder = true

			structure = append(structure, node)
		}
	}
	return structure
}

func Upload(ctx context.Context, newNode *fs.Inode, fullPath string, fileName string, content []byte) *fs.Inode {
	s := new(files.CommitInfo)
	s.Path = "/" + fullPath
	s.Mode = &files.WriteMode{Tagged: dropbox.Tagged{"overwrite"}}
	s.Autorename = false
	s.Mute = false
	s.StrictConflict = false

	fmt.Println("Uploading at path: " + s.Path)

	t := new(virtualFile)
	t.Data = content
	t.offset = 0
	dbx := files.New(config)

	dbx.Upload(s, t)

	return newNode
}

// Sends a get_metadata request for a given path and returns the response.
func getFileMetadata(c files.Client, path string) (files.IsMetadata, error) {
	arg := files.NewGetMetadataArg(path)

	res, err := c.GetMetadata(arg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

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

// Utility functions for ConstructTreeFromDrpPaths:

// Returns the string after the last '/'.
// E.g: /preda/raluca/antonio -> antonio
// 		/preda/raluca -> raluca
func lastFolderFromPath(path string) string {
	slices := strings.Split(path, "/")
	return slices[len(slices)-1]
}

// Returns the string before the last '/'.
// E.g: /preda/raluca/antonio -> /preda/raluca
// 		/preda/raluca -> /preda
func firstPartFromPath(path string) string {
	return path[:strings.LastIndex(path, "/")]
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

// Uploads folder to Dropbox.
func UploadFolder(fullPath string) {
	createFolderArg := files.CreateFolderArg{
		Path:       fullPath,
		Autorename: false,
	}

	dbx := files.New(config)
	dbx.CreateFolderV2(&createFolderArg)
}

func UploadDelete(drpn *DrpFileNode, name string) {
	fullPath := "/" + filepath.Join(drpn.Inode.Path(nil), name)
	fmt.Println(fullPath)
	deleteArg := files.DeleteArg{
		Path: fullPath,
	}
	dbx := files.New(config)
	dbx.DeleteV2(&deleteArg)
}
