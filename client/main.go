package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/shindakun/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func uploadToServer(c proto.GetFileServiceClient, dirandname string, data []byte) {
	s := strings.Split(dirandname, "/")
	dir := s[0]
	filename := s[1]

	log.Println("Uploading to server:", dirandname)

	response, err := c.UploadFile(context.Background(), &proto.FileResponse{
		Directory: dir,
		FileName:  filename,
		Data:      data,
	})
	if err != nil {
		log.Fatalf("Error when calling UploadFile: %s", err)
	}
	log.Printf("Response from server: %s", response.Body)
}

func downloadFile(conn proto.GetFileServiceClient, f string) {
	log.Println("Downloading file:", f)

retry:
	c, err := ftp.Dial("archives.thebbs.org:21", ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Println(err, "Retrying...")
		time.Sleep(5 * time.Minute)
		goto retry
	}

	err = c.Login("anonymous", "anonymous@anonymous.com")
	if err != nil {
		log.Println(err)
	}

	dir := strings.Split(f, "/")[0]
	if _, err := os.Stat(dir); err == nil {

		// path/to/whatever exists
		//fmt.Println("File exists")

	} else if errors.Is(err, os.ErrNotExist) {

		// path/to/whatever does *not* exist
		err := os.Mkdir(dir, 0777)
		if err != nil {
			log.Println(err)
		}
	}

	r, err := c.Retr(f)
	if err != nil {

		// Create dummy file so we can carry on
		outFile, err := os.Create(f)
		if err != nil {
			log.Println(err)
		}
		defer outFile.Close()

		// Bail out if file does not exist or other error
		if err := c.Quit(); err != nil {
			log.Println(err)
		}
		return
	}
	defer r.Close()

	outFile, err := os.Create(f)
	if err != nil {
		log.Println(err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, r)
	if err != nil {
		log.Fatal(err)
	}

	buf, err := io.ReadAll(r)
	if err != nil {
		log.Println(err)
	}
	uploadToServer(conn, f, buf)

	if err := c.Quit(); err != nil {
		log.Println(err)
	}
}

func main() {
	fmt.Println("protobuf")

	if len(os.Args) < 2 {
		log.Fatal("IP address of server is required")
	}

	var conn *grpc.ClientConn
	conn, err := grpc.Dial(os.Args[1]+":9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	defer conn.Close()

	c := proto.NewGetFileServiceClient(conn)

	loop := true

	for loop {
		response, err := c.GetFile(context.Background(), &proto.Message{Body: "next"})
		if err != nil {
			loop = false
			log.Printf("Error when calling GetFile: %s", err)
		}
		log.Printf("Response from server: %s", response.Filelocation)

		downloadFile(c, response.Filelocation)
	}
}
