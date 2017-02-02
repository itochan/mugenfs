package main

import (
	"flag"
	"log"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"github.com/itochan/mugenfs/driveApi"
	"github.com/itochan/mugenfs/fuse"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  mugenfs MOUNTPOINT")
	}

	driveApi.Init()
	fuse.Init()
	nfs := pathfs.NewPathNodeFs(&fuse.MugenFs{FileSystem: pathfs.NewDefaultFileSystem()}, nil)
	server, _, err := nodefs.MountRoot(flag.Arg(0), nfs.Root(), nil)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	server.Serve()
}
