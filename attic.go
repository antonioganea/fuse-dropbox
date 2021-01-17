package main

import (
	"fmt"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
)

// WORKING UNUSED/DEPRECATED FUNCTIONS
// mostly used for testing while developing

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
