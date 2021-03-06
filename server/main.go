package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/shindakun/protobuf/proto"
	"google.golang.org/grpc"
)

var files []string

type Server struct {
	proto.UnimplementedGetFileServiceServer
}

func loadFileList() {

	// Open file
	f, err := os.Open("files.txt")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	// Read file line by line
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) GetFile(ctx context.Context, in *proto.Message) (*proto.Request, error) {
	if in.Body == "next" {
	newFile:
		var file string
		file, files = files[0], files[1:]

		if _, err := os.Stat(file); err == nil {
			goto newFile
			// path/to/whatever exists
			//fmt.Println("File exists")

		} else if errors.Is(err, os.ErrNotExist) {

			// path/to/whatever does *not* exist
			return &proto.Request{
				Filelocation: file,
			}, nil
		}

		// return &proto.Request{
		// 	Filelocation: file,
		// }, nil
	}
	return &proto.Request{Filelocation: ""}, errors.New("error")
}

func (s *Server) UploadFile(ctx context.Context, in *proto.FileResponse) (*proto.Message, error) {
	log.Printf("Recieving file: %s", in.Path)

	splitPath := strings.Split(in.Path, "/")
	directory := strings.Join(splitPath[:len(splitPath)-1], "/")

	if _, err := os.Stat(directory); err == nil {

		// path/to/whatever exists
		//fmt.Println("File exists")

	} else if errors.Is(err, os.ErrNotExist) {

		// path/to/whatever does *not* exist
		err := os.MkdirAll(directory, 0777)
		if err != nil {
			log.Println(err)
		}
	}

	inFile, err := os.Create(in.Path)
	if err != nil {
		log.Println(err)
	}

	defer inFile.Close()

	_, err = inFile.Write(in.Data)
	if err != nil {

		// log.Fatal(err)
		return &proto.Message{
			Body: "error",
		}, errors.New("error writing file")
	}

	return &proto.Message{
		Body: "File uploaded successfully",
	}, nil
}

func main() {
	fmt.Println("protobuf")

	loadFileList()

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := Server{}
	grpcServer := grpc.NewServer()
	proto.RegisterGetFileServiceServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
