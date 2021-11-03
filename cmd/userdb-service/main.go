package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"syscall"

	"github.com/arianvp/go-systemd-userdb/userdb"
	"github.com/coreos/go-systemd/v22/activation"
)

type UserDatabase interface {
	GetUserRecord(*userdb.GetUserRecordRequest) func() userdb.GetUserRecordReply
}

type ExampleDatabase struct {
	userRecords []userdb.UserRecord
}

func (d *ExampleDatabase) GetUserRecord(req *userdb.GetUserRecordRequest) func() userdb.GetUserRecordReply {

	current := 0
	return func() userdb.GetUserRecordReply {
		if req.Parameters.Uid != nil {
			return userdb.GetUserRecordReply{
				Parameters: userdb.GetUserRecordReplyParams{},
				Continues:  false,
				Error:      "",
			}
		} else if req.Parameters.UserName != nil {
			return userdb.GetUserRecordReply{
				Parameters: userdb.GetUserRecordReplyParams{},
				Continues:  false,
				Error:      "",
			}
		} else if req.More {
			record := d.userRecords[current]
			continues := current < len(d.userRecords)-1
			current++
			return userdb.GetUserRecordReply{
				Parameters: userdb.GetUserRecordReplyParams{
					Record: record,
				},
				Continues: continues,
			}
		} else {
			return userdb.GetUserRecordReply{
				Error: "unknown",
			}
		}
	}
}

type UserDatabaseSession struct {
	conn    net.Conn
	encoder *json.Encoder
	reader  *bufio.Reader
}

// TODO; should I return errors or report them inline?
func (s *UserDatabaseSession) Handle(database UserDatabase) error {
	data, err := s.reader.ReadBytes(0)
	if err != nil {
		return err
	}

	type parameters struct {
		Method string `json:"method"`
	}
	type request struct {
		Parameters parameters `json:"parameters"`
	}

	var req request
	if err := json.Unmarshal(data, &req); err != nil {
		return err
	}
	switch req.Parameters.Method {
	case "io.systemd.UserDatabase.GetUserRecord":
		var req userdb.GetUserRecordRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return err
		}
		continues := true
		b := [1]byte{0}
		getUserRecord := database.GetUserRecord(&req)
		for continues {
			reply := getUserRecord()
			if err := s.encoder.Encode(reply); err != nil {
				return err
			}
			if _, err := s.conn.Write(b[:]); err != nil {
				return err
			}
			continues = reply.Continues
		}

	case "io.systemd.UserDatabase.GetGroupRecord":
	case "io.systemd.UserDatabase.GetMemberships":
	default:
		return fmt.Errorf("unimplemented method %s", req.Parameters.Method)
	}
	return nil
}

func main() {
	listeners, err := activation.Listeners()
	if err != nil {
		log.Fatal(err)
	}
	var listener net.Listener
	if len(listeners) != 1 {
		oldmask := syscall.Umask(0)
		listener, err = net.Listen("unix", "/run/systemd/userdb/me.arianvp.userdb")
		if err != nil {
			log.Fatal(err)
		}
		syscall.Umask(oldmask)
	} else {
		listener = listeners[1]
	}
	for {
		conn, err := listener.Accept()
		session := UserDatabaseSession{
			conn:    conn,
			encoder: json.NewEncoder(conn),
			reader:  bufio.NewReader(conn),
		}
		go session.Handle(&ExampleDatabase{
			userRecords: []userdb.UserRecord{
				{
					UserName: "arian",
				},
				{
					UserName: "lennart",
				},
			},
		})
		if err != nil {
			log.Fatal(err)
		}
	}

}
