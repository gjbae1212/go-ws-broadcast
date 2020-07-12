package websocket

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type (
	ErrorHandler func(error)

	Breaker interface {
		Register(conn *websocket.Conn) (*Client, error)
		UnRegister(client *Client) error
		BroadCast(msg Message) error
		MaxReadLimit() int64
	}

	// Breaker could manage all of something on the websocket.
	breaker struct {
		// ClientMap for all of user connections.
		clientMap map[*Client]bool

		// Message for broadcast
		broadcast chan Message

		// Register connection
		register chan *Client

		// UnRegister connection
		unregister chan *Client

		// Error handler
		errorHandler ErrorHandler

		// Max read limit
		maxReadLimit int64

		// stop
		quit chan bool
	}
)

// NewBreaker returns a breaker for websocket.
func NewBreaker(opts ...Option) (Breaker, error) {
	bk := &breaker{
		clientMap:  make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		quit:       make(chan bool),
	}

	// default message length and error handler
	o := []Option{
		WithErrorHandlerOption(func(err error) {
			fmt.Printf("%+v \n", err)
		}),
		WithMaxReadLimit(512),
	}
	o = append(o, opts...)
	for _, opt := range o {
		opt.apply(bk)
	}
	// default message pool length
	if bk.broadcast == nil {
		WithMaxMessagePoolLength(100).apply(bk)
	}
	bk.start()
	return Breaker(bk), nil
}

// Register is added websocket connection to Breaker.
func (bk *breaker) Register(conn *websocket.Conn) (*Client, error) {
	if conn == nil {
		return nil, fmt.Errorf("[err] Register empty params")
	}
	client, err := NewClient(bk, conn)
	if err != nil {
		return nil, fmt.Errorf("[err] Register %w", err)
	}
	bk.register <- client
	return client, nil
}

// UnRegister is deleted websocket connection from Breaker.
func (bk *breaker) UnRegister(client *Client) error {
	if client == nil {
		return fmt.Errorf("[err] UnRegister empty params")
	}
	bk.unregister <- client
	return nil
}

// Broadcast is a message to all of clients which is connected.
func (bk *breaker) BroadCast(msg Message) error {
	if msg == nil {
		return fmt.Errorf("[err] BroadCast empty params")
	}
	bk.broadcast <- msg
	return nil
}

// MaxReadLimit returns max read limit.
func (bk *breaker) MaxReadLimit() int64 {
	return bk.maxReadLimit
}

// Start a Breaker Loop
func (bk *breaker) start() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				werr := fmt.Errorf("[err] breaker start panic %w", r.(error))
				bk.errorHandler(werr)
				go bk.start()
			}
		}()

		for {
			select {
			case client := <-bk.register:
				bk.clientMap[client] = true
			case client := <-bk.unregister:
				if _, ok := bk.clientMap[client]; ok {
					delete(bk.clientMap, client)
					close(client.send)
					if err := client.conn.Close(); err != nil {
						werr := fmt.Errorf("[err] websocket close %w", err)
						bk.errorHandler(werr)
					}
				}
			case msg := <-bk.broadcast:
				for client, _ := range bk.clientMap {
					client.send <- msg
				}
			case <-bk.quit:
				return
			}
		}
	}()
}

func (bk *breaker) stop() {
	bk.quit <- true
}
