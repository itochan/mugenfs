package fuse

import (
	"fmt"
	"log"

	drive "google.golang.org/api/drive/v3"

	"github.com/boltdb/bolt"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"os"

	"encoding/json"

	"github.com/itochan/mugenfs/driveApi"
)

type MugenFs struct {
	pathfs.FileSystem
}

const DirBucketName = "Dir"

var database *bolt.DB

func Init() {
	os.Mkdir(os.Getenv("HOME")+"/.mugenfs", 0700)
	db, err := bolt.Open(os.Getenv("HOME")+"/.mugenfs/metadata.db", 0600, nil)
	database = db
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(DirBucketName)); err != nil {
			return err
		}
		return nil
	})
}

func (me *MugenFs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	switch name {
	case "":
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	default:
		return &fuse.Attr{
			Mode: fuse.S_IFREG | 0644, Size: uint64(len(name)),
		}, fuse.OK
	}
	return nil, fuse.ENOENT
}

func (me *MugenFs) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	req := make(chan []byte)
	go func() {
		err := database.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(DirBucketName))
			v := b.Get([]byte("/"))
			req <- v
			return nil
		})
		if err != nil {
			log.Fatalln(err.Error())
		}
	}()
	value := <-req

	var r *drive.FileList
	var err error
	if len(value) == 0 {
		r, err = driveApi.List("'root' in parents")
		err := database.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(DirBucketName))
			v, err := json.Marshal(r)
			if err != nil {
				return err
			}
			err = b.Put([]byte("/"), []byte(v))
			return err
		})
		if err != nil {
			log.Panicln(err.Error())
		}
	} else {
		err = json.Unmarshal(value, &r)
	}

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
