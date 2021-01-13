package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/oauth2"

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

type DrpFileNode struct {
	// Must embed an Inode for the struct to work as a node.
	fs.Inode

	// drpPath is the path of this file/directory
	drpPath string
}

// type NodeReader interface {
// 	Read(ctx context.Context, f FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno)
// }

var _ = (fs.NodeGetattrer)((*DrpFileNode)(nil))

func (bn *DrpFileNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	//bn.mu.Lock()
	//defer bn.mu.Unlock()

	dbx := files.New(config)

	// TODO: make sure file name is correct
	downloadArg := files.NewDownloadArg(bn.drpPath)

	meta, _, err := dbx.Download(downloadArg)
	if err != nil {
		return 404
	}

	//bn.getattr(out)
	out.Size = meta.Size
	//out.SetTimes(nil, &bn.mtime, nil)

	return 0
}


//folosim interfata NodeReader
var _ = (fs.NodeReader)((*DrpFileNode)(nil))

//Read se aplica pe un DrpFileNode 
//array de bytes numit dest, oof int64 nr pe 64 biti -> ptr offset 
//ultima paranteza return 

func (drpn *DrpFileNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	// drpn.mu.Lock()
	// defer drpn.mu.Unlock()

	destLen := int64(len(dest))

	dbx := files.New(config)

	// TODO: make sure file name is correct
	downloadArg := files.NewDownloadArg(drpn.drpPath)

	meta, content, err := dbx.Download(downloadArg)
	if err != nil {
		return nil, 404
	}
	if off == int64(meta.Size) {
		return fuse.ReadResultData(make([]byte, 0)), 0
	}

	// Here we'd need a better file reading mechanic ( so we know for sure we've read all )
	b1 := make([]byte, meta.Size)
	n1, err := content.Read(b1)
	// if int64(n1) < destLen {
	// 	destLen = int64(n1)
	// }

	fmt.Println(string(b1[:n1]))

	// TRACTOR !!!!!!!!
	// Chestia asta nu e facuta tocmai bn, dar merge <3
	var readSize int64
	if int64(meta.Size) < destLen {
		readSize = int64(meta.Size)
	} else {
		readSize = off + destLen
	}

	//constructor sa faca un stream, e un SLICE!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//Alexandra mai citeste documentatia la go!!!!!
	//0 e cod de eroare
	return fuse.ReadResultData(b1[off:readSize]), 0
}

var _ = (fs.NodeOpener)((*DrpFileNode)(nil))

//f = func
//Open se aplica pe un pointer de tip DrpFileNode nuimt f (f e pointer)
//context ca la daw
//uint32 - integer unsigned pe 32 de biti
//un flag se refera la chestiile pe care le combini - read, write, append (BinaryFlags sau BitFlags), 
//ca sa nu tina 8 boolene care sa contina 8 biti, se tineua flag-uro pe biti => super COOL!!!!!! (se folosesc la chestii low-level)
//ultima paranteza e return type-ul (in Go e misto ca poti sa ai return multiplu)
//a doua paranteza reprezinta parametri

func (f *DrpFileNode) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, 0, 0
}

func validatePath(p string) (path string, err error) {
	path = p

	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	path = strings.TrimSuffix(path, "/")

	return
}

func readToken(filePath string) (string, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeToken(filePath string, token string) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Doesn't exist; lets create it
		err = os.MkdirAll(filepath.Dir(filePath), 0700)
		if err != nil {
			return
		}
	}
	b := []byte(token)
	if err := ioutil.WriteFile(filePath, b, 0600); err != nil {
		return
	}
}

// returns a list of paths
// that represent the DFS traversal of  the drobpox folder
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

func initDbx() (err error) {

	memorizedToken, err := readToken(configFileName)

	if err != nil {
		conf := oauth2.Config{ // maybe a & reference here
			ClientID:     appKey,
			ClientSecret: appSecret,
			Endpoint:     dropbox.OAuthEndpoint(""),
		}

		fmt.Printf("1. Go to %v\n", conf.AuthCodeURL("state"))
		fmt.Printf("2. Click \"Allow\" (you might have to log in first).\n")
		fmt.Printf("3. Copy the authorization code.\n")
		fmt.Printf("Enter the authorization code here: ")

		var code string
		if _, err = fmt.Scan(&code); err != nil {
			return
		}
		var token *oauth2.Token
		ctx := context.Background()
		token, err = conf.Exchange(ctx, code)
		if err != nil {
			return
		}
		memorizedToken = token.AccessToken
	}

	writeToken(configFileName, memorizedToken)

	config = dropbox.Config{
		Token:    memorizedToken,
		LogLevel: dropbox.LogOff, // if needed, set the desired logging level. Default is off
	}

	fmt.Print("Dropbox Config'd!\n")

	return
}

func copyOperation() error {
	// Here we do some basic operation : copying file.txt into /dirA/newfile.txt
	dbx := files.New(config)

	relocArg := files.NewRelocationArg("/file.txt", "/dirA/copiedFile.txt")

	if _, err := dbx.CopyV2(relocArg); err != nil {
		return err
	}
	return nil
}

func downloadOp() {
	dbx := files.New(config)

	downloadArg := files.NewDownloadArg("/file.txt")

	meta, content, err := dbx.Download(downloadArg)
	if err != nil {
		return
	}

	// Here we'd need a better file reading mechanic ( so we know for sure we've read all )
	b1 := make([]byte, meta.Size)
	n1, err := content.Read(b1)
	/////////////////////////

	fmt.Println(string(b1[:n1]))
}

// Sends a get_metadata request for a given path and returns the response
func getFileMetadata(c files.Client, path string) (files.IsMetadata, error) {
	arg := files.NewGetMetadataArg(path)

	res, err := c.GetMetadata(arg)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// /preda/raluca/antonio -> antonio
// /preda/raluca -> raluca
func lastFolderFromPath(path string) string {
	slices := strings.Split(path, "/")
	return slices[len(slices)-1]
}

// /preda/raluca/antonio -> /preda/raluca
// /preda/raluca -> /preda
func firstPartFromPath(path string) string {
	return path[:strings.LastIndex(path, "/")]
}

func listDirTopLevel() error {
	dbx := files.New(config)

	arg := files.NewListFolderArg("")
	arg.Recursive = false

	res, err := dbx.ListFolder(arg)
	var entries []files.IsMetadata
	if err != nil {
		switch e := err.(type) {
		case files.ListFolderAPIError:
			// Don't treat a "not_folder" error as fatal; recover by sending a
			// get_metadata request for the same path and using that response instead.
			if e.EndpointError.Path.Tag == files.LookupErrorNotFolder {
				var metaRes files.IsMetadata
				metaRes, err = getFileMetadata(dbx, "/")
				entries = []files.IsMetadata{metaRes}
			} else {
				return err
			}
		default:
			return err
		}

		// Return if there's an error other than "not_folder" or if the follow-up
		// metadata request fails.
		if err != nil {
			return err
		}
	} else {
		entries = res.Entries

		for res.HasMore {
			arg := files.NewListFolderContinueArg(res.Cursor)

			res, err = dbx.ListFolderContinue(arg)
			if err != nil {
				return err
			}

			entries = append(entries, res.Entries...)
		}
	}

	for _, entry := range entries {
		switch f := entry.(type) {
		case *files.FileMetadata:
			//printFileMetadata(w, f)
			fmt.Println(f.PathDisplay)
		case *files.FolderMetadata:
			//printFolderMetadata(w, f)
			fmt.Println(f.PathDisplay)
		}
	}
	return err
}

var inoIterator uint64 = 2

func AddFile(ctx context.Context, node *fs.Inode, fileName string, fullPath string) *fs.Inode {
	// newfile := node.NewInode(ctx, operations, stable)
	drpFileNode := DrpFileNode{}
	drpFileNode.drpPath = fullPath
	newfile := node.NewInode(
		ctx, &drpFileNode, fs.StableAttr{Ino: inoIterator})
	node.AddChild(fileName, newfile, false)

	inoIterator++

	return newfile
}

func AddFolder(ctx context.Context, node *fs.Inode, folderName string) *fs.Inode {
	dir := node.NewInode(
		ctx, &fs.MemRegularFile{
			Data: []byte("sample dir data"),
			Attr: fuse.Attr{
				Mode: 0644,
			},
		}, fs.StableAttr{Ino: inoIterator, Mode: fuse.S_IFDIR})
	node.AddChild(folderName, dir, false)
	inoIterator++

	return dir
}

func ConstructTreeFromDrpPaths(ctx context.Context, r *HelloRoot, structure []DrpPath) {
	// aici se va construi arborele

	var m map[string](*fs.Inode) = make(map[string](*fs.Inode))

	m[""] = &r.Inode

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
			newNode = AddFile(ctx, parentNode, newNodeName, entry.path)
		}

		m[containingFolder+"/"+newNodeName] = newNode

		fmt.Println("Mapped the newly created node in " + containingFolder + "/" + newNodeName)
	}
}

func ConstructTree(ctx context.Context, r *HelloRoot) {
	ConstructTreeFromDrpPaths(ctx, r, getDropboxTreeStructure())
}


//Alexandra

//interfata
/*
var _ = (fs.NodeWriter)((*DrpFileNode)(nil))

func(drpn *DrpFileNode) Write(ctx context.Context, f FileHandle, data []byte, off int64) (written uint32, errno syscall.Errno) {
	//open callback - event - linia 
	//sourceLen := int64(len(dest))

	return errno
}

func uploadOp() {
	dbx := files.New(config)

	downloadArg := files.NewDownloadArg("/file.txt")

	meta, content, err := dbx.Download(downloadArg)
	if err != nil {
		return
	}

	// Here we'd need a better file reading mechanic ( so we know for sure we've read all )
	b1 := make([]byte, meta.Size)
	n1, err := content.Read(b1)
	/////////////////////////

	fmt.Println(string(b1[:n1]))
}
*/

func drb_main() {
	//initDbx()
	//copyOperation()
	//downloadOp()
	//listDirTopLevel()
}
