package fs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/itochan/mugenfs/driveApi"
	drive "google.golang.org/api/drive/v3"
)

const dirBucketName = "Dir"
const fileBucketName = "File"

var database *bolt.DB

func Init() {
	os.Mkdir(os.Getenv("HOME")+"/.mugenfs", 0700)
	db, err := bolt.Open(os.Getenv("HOME")+"/.mugenfs/metadata.db", 0600, nil)
	database = db
	if err != nil {
		log.Fatal(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte(dirBucketName)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(fileBucketName)); err != nil {
			return err
		}
		return nil
	})

	driveApi.Init()
}

func getList() (*drive.FileList, error) {
	req := make(chan []byte)
	go func() {
		err := database.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(dirBucketName))
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
			b := tx.Bucket([]byte(dirBucketName))
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

		for _, f := range r.Files {
			database.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(fileBucketName))
				v, err := json.Marshal(f)
				if err != nil {
					return err
				}
				err = b.Put([]byte("/"+f.Name), []byte(v))
				return err
			})
			fmt.Printf("%s (%s)\n", f.Name, f.Id)
		}
	} else {
		err = json.Unmarshal(value, &r)
	}

	return r, err
}

func getFileInfo(fname string) (*drive.File, error) {
	req := make(chan []byte)
	go func() {
		err := database.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(fileBucketName))
			v := b.Get([]byte("/" + fname))
			req <- v
			return nil
		})
		if err != nil {
			log.Fatalln(err.Error())
		}
	}()
	value := <-req

	var r *drive.File
	if len(value) == 0 {
		return nil, nil
	}
	err := json.Unmarshal(value, &r)
	return r, err
}
