package fs

import (
	"encoding/json"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/itochan/mugenfs/driveApi"
	drive "google.golang.org/api/drive/v3"
)

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

	driveApi.Init()
}

func getList() (*drive.FileList, error) {
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

	return r, err
}
