package websocket

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 15 * time.Second

	pongWait = 60 * time.Second

	pingWait = 50 * time.Second
)

var (
	Upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
)

// Client is websocket client.
type Client struct {
	// breaker
	breaker Breaker

	// a connection of user
	conn *websocket.Conn

	// a buffer under send.
	send chan Message
}

// NewClient creates client for websocket.
func NewClient(bk Breaker, conn *websocket.Conn) (*Client, error) {
	client := &Client{
		breaker: bk,
		conn:    conn,
		send:    make(chan Message, 200), // messages is stacked under 200.
	}
	go client.loopOfRead()
	go client.loopOfWrite()
	return client, nil
}

func (client *Client) loopOfRead() {
	defer func() {
		if err := client.breaker.UnRegister(client); err != nil {
			werr := fmt.Errorf("[err] loopOfRead %w", err)
			client.breaker.(*breaker).errorHandler(werr)
		}
	}()

	// it is setting max count which closes websocket connection when exceeds its count.
	client.conn.SetReadLimit(client.breaker.MaxReadLimit())
	// it is setting deadline of a websocket.
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	// it is newly setting deadline of a websocket when a ping message received.
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// If length of message a client is received will exceed over limit, it is raised error.
		_, message, err := client.conn.ReadMessage()
		if err != nil { // Usually be closed a connection, a error is raised
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
				werr := fmt.Errorf("[err] loopOfRead read close %w", err)
				client.breaker.(*breaker).errorHandler(werr)
			}
			return
		}
		// Message received will broadcast all of users.
		if err := client.breaker.BroadCast(&internalMessage{payload: message}); err != nil {
			werr := fmt.Errorf("[err] loopOfRead broadcast %w", err)
			client.breaker.(*breaker).errorHandler(werr)
		}
	}
}

func (client *Client) loopOfWrite() {
	ticker := time.NewTicker(pingWait)
	defer func() {
		if err := client.breaker.UnRegister(client); err != nil {
			werr := fmt.Errorf("[err] loopOfWrite panic %w", err)
			client.breaker.(*breaker).errorHandler(werr)
		}
		ticker.Stop()
	}()

	for {
		select {
		case msg, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			// If channel of the send is closed
			if !ok {
				// Do not check a error, because already connection will possibly be closed.
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// A message send to a user.
			if err := client.conn.WriteMessage(websocket.TextMessage, msg.GetMessage()); err != nil { // Usually be closed a connection, a error is raised
				if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseGoingAway) {
					werr := fmt.Errorf("[err] loopOfWrite write close %w", err)
					client.breaker.(*breaker).errorHandler(werr)
				}
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			// a ping message will periodically send to a client.
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				werr := fmt.Errorf("[err] loopOfWrite ping %w", err)
				client.breaker.(*breaker).errorHandler(werr)
				return
			}
		}
	}
}
