package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

type server struct {
	rooms    map[string]*room
	commands chan command
}

func newServer() *server {
	return &server{
		rooms:    make(map[string]*room),
		commands: make(chan command),
	}
}

func (s *server) run() {
	for cmd := range s.commands {
		switch cmd.id {
		case CMD_ADI:
			s.adi(cmd.client, cmd.args)

		case CMD_JOIN:
			s.join(cmd.client, cmd.args)
		case CMD_MSG:
			s.msg(cmd.client, cmd.args)

		case CMD_ROOMS:
			s.listRooms(cmd.client)
	
		case CMD_QUIT:
			s.quit(cmd.client)
		}
	}
}

func (s *server) adi(c *client, args []string) {
	if len(args) < 2 {
		c.msg("adi is required. usage: /adi NAME")
		return
	}

	c.adi = args[1]
	c.msg(fmt.Sprintf("all righty, i will call you: %s", c.adi))
}

func (s *server) join(c *client, args []string) {
	roomName := args[1]
	r, ok := s.rooms[roomName]
	if !ok {
		r = &room{
			name:    roomName,
			members: make(map[net.Addr]*client),
		}
		s.rooms[roomName] = r
	}

	r.members[c.conn.RemoteAddr()] = c

	s.quitCurrentRoom(c)

	c.room = r
	r.broadcast(c, fmt.Sprintf("%s has joined the room", c.adi))
	c.msg(fmt.Sprintf("welcome to %s", r.name))
}

func (s *server) msg(c *client, args []string) {
	if c.room == nil {
		c.err(errors.New("you must join room first"))
		return
	}

	c.room.broadcast(c, c.adi+": "+strings.Join(args[1:len(args)], " "))
}
func (s *server) listRooms(c *client) {
	var rooms []string
	for name := range s.rooms {
		rooms = append(rooms, name)
	}

	c.msg(fmt.Sprintf("available rooms are: %s", strings.Join(rooms, ", ")))
}

func (s *server) quit(c *client) {
	log.Printf("client has disconnected: %s", c.conn.RemoteAddr().String())

	s.quitCurrentRoom(c)
	c.msg("bye bye")
	c.conn.Close()
}
func (s *server) quitCurrentRoom(c *client) {
	if c.room != nil {
		delete(c.room.members, c.conn.RemoteAddr())
		c.room.broadcast(c, fmt.Sprintf("%s has left the room"))
	}
}

func (s *server) newClient(conn net.Conn) *client {
	log.Printf("new client has been connected: %s", conn.RemoteAddr().String())

	return &client{
		conn:     conn,
		adi:      "anonymous",
		commands: s.commands,
	}
}
