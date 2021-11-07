package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/arianvp/go-systemd-userdb/userdb"
	"github.com/coreos/go-systemd/v22/activation"
)

type ExampleDatabase struct {
	userRecords []userdb.UserRecord
}

func (db *ExampleDatabase) GetGroupRecord(req userdb.GetGroupRecordRequest) func() userdb.GetGroupRecordReply {
	return nil
}

func (db *ExampleDatabase) GetMemberships(req userdb.GetMembershipsRequest) func() userdb.GetMembershipsReply {
	return nil
}

func (d *ExampleDatabase) GetUserRecord(req userdb.GetUserRecordRequest) func() userdb.GetUserRecordReply {
	current := 0
	return func() (reply userdb.GetUserRecordReply) {
		if req.Parameters.Uid != nil {
			for _, r := range d.userRecords {
				if r.Uid == req.Parameters.Uid {
					reply.Parameters.Record = &r
					break
				}
			}
		}
		if req.Parameters.UserName != nil {
			for _, r := range d.userRecords {
				if r.UserName == *req.Parameters.UserName {
					if reply.Parameters.Record == nil || reply.Parameters.Record.Uid == r.Uid {
						reply.Parameters.Record = &r
						break
					} else {
						reply.Error = userdb.ConflictingRecordFound
						return
					}
				}
			}
		}
		if (req.Parameters.Uid != nil || req.Parameters.UserName != nil) && reply.Parameters.Record == nil {
			reply.Error = userdb.NoRecodFound
			return
		}
		// list all
		if req.Parameters.Uid == nil && req.Parameters.UserName == nil && req.More {
			reply.Parameters.Record = &d.userRecords[current]
			continues := current < len(d.userRecords)-1
			current++
			reply.Continues = continues
		}
		return
	}
}

type UserDatabaseSession struct {
	conn    net.Conn
	encoder *json.Encoder
	reader  *bufio.Reader
}

// TODO; should I return errors or report them inline?
func (s *UserDatabaseSession) Handle(db userdb.UserDatabase) error {
	defer s.conn.Close()
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
		getUserRecord := db.GetUserRecord(req)
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
		var req userdb.GetGroupRecordRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return err
		}
		continues := true
		b := [1]byte{0}
		getGroupRecord := db.GetGroupRecord(req)
		for continues {
			reply := getGroupRecord()
			if err := s.encoder.Encode(reply); err != nil {
				return err
			}
			if _, err := s.conn.Write(b[:]); err != nil {
				return err
			}
			continues = reply.Continues
		}
	case "io.systemd.UserDatabase.GetMemberships":
	default:
		return fmt.Errorf("unimplemented method %s", req.Parameters.Method)
	}
	return nil
}

type server struct {
	l net.Listener
}

// Serve always returns a non-nill error and closes l
func (srv *server) Serve(ctx context.Context) error {
	defer srv.l.Close()
	for {
		c, err := srv.l.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return err
			}
		}
		go func() {
			<-ctx.Done()
			c.Close()
		}()
		session := UserDatabaseSession{
			conn:    c,
			encoder: json.NewEncoder(c),
			reader:  bufio.NewReader(c),
		}
		uid := uint32(999999)
		go session.Handle(&ExampleDatabase{
			userRecords: []userdb.UserRecord{
				{
					UserFields: userdb.UserFields{
						UserName: "arian",
						Uid:      &uid,
						Gid:      &uid,
					},
				},
			},
		})
	}
}

// returns a context that gets cancalled on any of the passed signals
func signalContext(sig ...os.Signal) context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sig...)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-c
		cancel()
	}()
	return ctx
}

func main() {
	ctx := signalContext(os.Interrupt, syscall.SIGTERM)
	listeners, err := activation.Listeners()
	if err != nil {
		log.Fatal(err)
	}
	listenConfig := net.ListenConfig{}
	var l net.Listener
	if len(listeners) != 1 {
		oldmask := syscall.Umask(0)
		l, err = listenConfig.Listen(ctx, "unix", "./lol-sock")
		if err != nil {
			log.Fatal(err)
		}
		syscall.Umask(oldmask)
	} else {
		l = listeners[1]
	}
	if err != nil {
		log.Fatal(err)
	}
	srv := server{l}
	go srv.Serve(ctx)
	<-ctx.Done()
	log.Print(ctx.Err())
}
