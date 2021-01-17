package main

import (
	"fmt"
	"path/filepath"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	"github.com/hanwen/go-fuse/fs"
	"golang.org/x/net/context"
)

type DrpPath struct {
	path     string
	isFolder bool
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
