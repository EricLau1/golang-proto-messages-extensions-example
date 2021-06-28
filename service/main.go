package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-proto-tags/pb"
	"log"
	"net"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func main() {

	listener, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}

	messengerService := NewMessagerService()

	grpcServer := grpc.NewServer()

	pb.RegisterMessengerServer(grpcServer, messengerService)

	log.Println("messenger service listening on :8888")

	grpcServer.Serve(listener)

}

type service struct {
	messages map[string]*pb.Message
}

func NewMessagerService() pb.MessengerServer {
	return &service{messages: make(map[string]*pb.Message)}
}

func (s *service) CreateMessage(ctx context.Context, req *pb.Message) (*pb.Message, error) {

	id := fmt.Sprint(time.Now().UnixNano())
	req.Id = id
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)
	req.Author = strings.TrimSpace(req.Author)

	ReadProtoExtensions(req)

	if req.Author == "" {
		req.Author = "anonymous"
	}

	req.Created = time.Now().Unix()
	req.Updated = time.Now().Unix()

	s.messages[id] = req

	return req, nil
}

func (s *service) GetMessages(req *pb.GetMessagesRequest, stream pb.Messenger_GetMessagesServer) error {

	for _, message := range s.messages {
		stream.Send(message)
	}

	return nil
}

type FieldProps struct {
	OriginalName string  `json:"original_name"`
	Protected    *bool   `json:"protected"`
	Editable     *bool   `json:"editable"`
	CustomName   *string `json:"custom_name"`
}

func ReadProtoExtensions(msg proto.Message) {

	m := proto.MessageReflect(msg)

	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {

		opts := fd.Options().(*descriptorpb.FieldOptions)

		fieldProps := &FieldProps{OriginalName: fd.JSONName()}

		protected, err := proto.GetExtension(opts, pb.E_Protected)
		if err == nil && protected != nil {
			isProtected, ok := protected.(*bool)
			if ok {
				fieldProps.Protected = isProtected
			}
		}

		editable, err := proto.GetExtension(opts, pb.E_Editable)
		if err == nil && editable != nil {
			isEditable, ok := editable.(*bool)
			if ok {
				fieldProps.Editable = isEditable
			}
		}

		customName, err := proto.GetExtension(opts, pb.E_CustomFieldName)
		if err == nil && customName != nil {
			name, ok := customName.(*string)
			if ok {
				fieldProps.CustomName = name
			}
		}

		b, _ := json.MarshalIndent(fieldProps, "", "\t")
		fmt.Println(string(b))

		return true
	})
}
