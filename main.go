package main

import (
	"log"
	"os"
	"time"

	"github.com/t3rm1n4l/go-mega"
)

var public *mega.Mega
var private *mega.Mega

func main() {
	private = mega.New()
	public = mega.New()

	err := private.Login("kidandcat+1@gmail.com", "akatsuki")
	if err != nil {
		panic(err)
	}
	err = public.Login("kidandcat+2@gmail.com", "akatsuki")
	if err != nil {
		panic(err)
	}

	loop()
}

func loop() {
	for {
		log.Print("Tick")
		root := private.FS.GetRoot()
		nodes, err := private.FS.GetChildren(root)
		if err != nil {
			panic(err)
		}
		log.Printf("Found %d nodes in private", len(nodes))
		for _, node := range nodes {
			privHash := node.GetName()
			foundnodes, err := public.FS.PathLookup(public.FS.GetRoot(), []string{privHash})
			if len(foundnodes) == 0 || err != nil {
				log.Printf("Node %s not found in public, err: %s, copying...", node.GetName(), err)
				copy(node)
			} else {
				log.Printf("Node %s found in public, skipping...", node.GetName())
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func copy(node *mega.Node) {
	defer func() {
		err := os.Remove("private.tmp")
		if err != nil {
			panic(err)
		}
	}()
	progress := make(chan int)
	go func() {
		total := 0
		for {
			if i, ok := <-progress; ok {
				total += i
				log.Printf("%s: Downloaded %.1f/%.1fMB", node.GetName(), float64(total)/1000000, float64(node.GetSize())/1000000)
			} else {
				log.Printf("Download of %s complete", node.GetName())
				break
			}
		}
	}()
	err := private.DownloadFile(node, "private.tmp", &progress)
	if err != nil {
		panic(err)
	}
	uploadProgress := make(chan int)
	log.Printf("Uploading %s", node.GetName())
	go func() {
		total := 0
		for {
			if i, ok := <-uploadProgress; ok {
				total += i
				log.Printf("%s: Uploaded %.1f/%.1fMB", node.GetName(), float64(total)/1000000, float64(node.GetSize())/1000000)
			} else {
				log.Printf("Upload of %s complete", node.GetName())
				break
			}
		}
	}()
	pubnode, err := public.UploadFile("private.tmp", public.FS.GetRoot(), node.GetName(), &uploadProgress)
	if err != nil {
		panic(err)
	}
	log.Printf("Linking %s", pubnode.GetName())
	link, err := public.Link(pubnode, true)
	if err != nil {
		panic(err)
	}
	log.Printf("%s made public in %s", pubnode.GetName(), link)
}
