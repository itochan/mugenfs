package fs

import (
	"fmt"
	"log"

	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type MugenFs struct {
	pathfs.FileSystem
}

func (me *MugenFs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	if name == "" {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	}
	r, err := getFileInfo(name)
	if err != nil {
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0644, Size: uint64(len(name)),
		}, fuse.OK
	}

	if r != nil {
		// layout := "yyyy-MM-DDTHH:mm:ssZ"
		atime, _ := time.Parse(time.RFC3339Nano, r.ViewedByMeTime)
		atimeUnix := atime.Unix()
		mtime, _ := time.Parse(time.RFC3339Nano, r.ModifiedTime)
		mtimeUnix := mtime.Unix()
		ctime, _ := time.Parse(time.RFC3339Nano, r.CreatedTime)
		ctimeUnix := ctime.Unix()
		return &fuse.Attr{
			Atime: uint64(atimeUnix), Mtime: uint64(mtimeUnix), Ctime: uint64(ctimeUnix),
			Mode: fuse.S_IFREG | 0644, Size: uint64(len(name)),
		}, fuse.OK
	}
	return nil, fuse.ENOENT
}

func (me *MugenFs) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	r, err := getList()

	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}

	if len(r.Files) == 0 {
		fmt.Println("No files found.")
		return nil, fuse.ENOENT
	}

	entries := make([]fuse.DirEntry, len(r.Files))
	for i, f := range r.Files {
		entries[i] = fuse.DirEntry{Name: f.Name, Mode: fuse.S_IFREG}
		fmt.Printf("%s (%s)\n", f.Name, f.Id)
	}
	return entries, fuse.OK
}

func (me *MugenFs) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	if name != "file.txt" {
		return nil, fuse.ENOENT
	}
	if flags&fuse.O_ANYWRITE != 0 {
		return nil, fuse.EPERM
	}
	return nodefs.NewDataFile([]byte(name)), fuse.OK
}
