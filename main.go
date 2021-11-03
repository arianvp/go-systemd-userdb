package main

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"

	"net"
)

type UserDatabaseService interface {
	GetUserRecord(GetUserRecordRequest) (GetUserRecordReply, error)
}
type UserDatabaseClient struct{}

type UserDatabaseDecoder struct {
	reader *bufio.Reader
}

type UserDatabaseEncoder struct {
	encoder *json.Encoder
	writer  io.Writer
}

func (e *UserDatabaseEncoder) EncodeGetUserRecord(req GetUserRecordRequest) error {
	err := e.encoder.Encode(req)
	if err != nil {
		return err
	}
	var b [1]byte
	_, err = e.writer.Write(b[:])
	return err
}

func NewUserDatabaseEncoder(w io.Writer) UserDatabaseEncoder {
	return UserDatabaseEncoder{
		encoder: json.NewEncoder(w),
		writer:  w,
	}
}

type UserRecord struct {
	UserName string `json:"userName"`
	// TODO other fields
}

type GetUserRecordRequestParams struct {
	UserName string `json:"userName,omitempty"`
	Uid      *int64 `json:"uid,omitempty"`
	Service  string `json:"service"`
}

type GetUserRecordRequest struct {
	Method     string                     `json:"method"`
	Parameters GetUserRecordRequestParams `json:"parameters"`
	More       bool                       `json:"more"`
}

type GetUserRecordReplyParams struct {
	Record UserRecord `json:"record"`
}

type GetUserRecordReply struct {
	Parameters GetUserRecordReplyParams `json:"parameters"`
	Continues  bool                     `json:"continues,omitempty"`
	Error      string                   `json:"error,omitempty"`
}

func (d *UserDatabaseDecoder) DecodeGetUserRecordReply() (reply *GetUserRecordReply, err error) {
	bytes, err := d.reader.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	reply = new(GetUserRecordReply)
	err = json.Unmarshal(bytes[:len(bytes)-1], reply)
	return
}

func NewUserDatabaseDecoder(reader io.Reader) UserDatabaseDecoder {
	return UserDatabaseDecoder{
		reader: bufio.NewReader(reader),
	}
}

func multiplex() error {
	dir := "/run/systemd/userdb/"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	replies := make(chan *GetUserRecordReply)
	errors := make(chan error)
	for _, f := range files {
		service := f.Name()
		if service == "io.systemd.Multiplexer" {
			continue
		}
		log.Printf("contacting %s", service)
		go func() {
			path := dir + service
			conn, err := net.Dial("unix", path)
			if err != nil {
				// TODO?
				log.Fatal(err)
			}
			defer conn.Close()
			encoder := NewUserDatabaseEncoder(conn)
			decoder := NewUserDatabaseDecoder(conn)

			encoder.EncodeGetUserRecord(GetUserRecordRequest{
				Method:     "io.systemd.UserDatabase.GetUserRecord",
				Parameters: GetUserRecordRequestParams{Service: service},
				More:       true,
			})

			continues := true
			for continues {
				reply, err := decoder.DecodeGetUserRecordReply()
				if err != nil {
					log.Print(err)
					errors <- err
					break
				}
				replies <- reply
				continues = reply.Continues
			}
		}()
	}
	for reply := range replies {
		log.Printf("%+v", reply)
	}
	return nil

}

func main() {

	service := "io.systemd.Multiplexer"
	conn, err := net.Dial("unix", "/run/systemd/userdb/"+service)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	encoder := NewUserDatabaseEncoder(conn)
	decoder := NewUserDatabaseDecoder(conn)

	encoder.EncodeGetUserRecord(GetUserRecordRequest{
		Method:     "io.systemd.UserDatabase.GetUserRecord",
		Parameters: GetUserRecordRequestParams{Service: service},
		More:       true,
	})

	continues := true
	for continues {
		reply, err := decoder.DecodeGetUserRecordReply()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%+v", reply)
		continues = reply.Continues
	}
}
