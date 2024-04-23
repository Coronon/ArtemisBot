package sockjs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type SockJSMessage struct {
	Command string                  `json:"command"`
	Headers map[string]string       `json:"headers"`
	Body    *map[string]interface{} `json:"body"`
}

var sockJSRegex = regexp.MustCompile(`^a?\["(?P<command>.+?)\\n(?P<headers>(?:[a-zA-Z0-9-_/#,. ]+:[a-zA-Z0-9-_/#,. ]+\\n)*)\\n(?P<body>.*)\\u0000"]$`)

// Parse a raw SockJS message
func ParseSockJSMessage(data string) (*SockJSMessage, error) {
	if data == "o" {
		return nil, nil
	}

	if data == `a["\n"]` {
		return nil, nil
	}

	msg := SockJSMessage{}

	// Parse command
	matches := sockJSRegex.FindStringSubmatch(data)
	if len(matches) == 0 {
		return nil, fmt.Errorf("failed to parse SockJS message: %s", data)
	}

	msg.Command = matches[sockJSRegex.SubexpIndex("command")]

	// Parse headers
	msg.Headers = make(map[string]string)
	headers := matches[sockJSRegex.SubexpIndex("headers")]
	for _, header := range strings.Split(headers, `\n`) {
		if header == "" {
			continue
		}

		parts := strings.Split(header, ":")
		msg.Headers[parts[0]] = parts[1]
	}

	// Parse body
	body := matches[sockJSRegex.SubexpIndex("body")]
	if body != "" {
		msg.Body = &map[string]interface{}{}
		body = strings.ReplaceAll(body, `\"`, `"`)
		err := json.Unmarshal([]byte(body), msg.Body)
		if err != nil {
			return nil, err
		}
	}

	return &msg, nil
}

type SockJSClient struct {
	mtx        sync.Mutex
	ws         *websocket.Conn
	msgCounter atomic.Int64
	sessionID  string

	msgChan  chan *SockJSMessage
	errChan  chan error
	doneChan chan struct{}
	isClosed bool
}

func NewSockJSClient(wsURL string, headers http.Header) (*SockJSClient, error) {
	client := &SockJSClient{
		mtx:        sync.Mutex{},
		msgCounter: atomic.Int64{},

		msgChan:  make(chan *SockJSMessage),
		errChan:  make(chan error),
		doneChan: make(chan struct{}),
		isClosed: false,
	}

	ws, _, err := websocket.DefaultDialer.Dial(
		wsURL,
		headers,
	)
	if err != nil {
		return nil, err
	}

	client.ws = ws

	client.sessionID, err = client.handleConnect()
	if err != nil {
		return nil, err
	}

	go client.handleProtocol()

	return client, nil
}

// Closes the websocket connection.
func (c *SockJSClient) Close() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.isClosed {
		return
	}

	close(c.doneChan)
	c.ws.Close()
	close(c.msgChan)
	close(c.errChan)
}

// Returns a channel that receives messages from the server.
//
// Protocol messages are handled internally and are not sent to the channel.
func (c *SockJSClient) Messages() <-chan *SockJSMessage {
	return c.msgChan
}

// Returns a channel that receives errors from the connection.
func (c *SockJSClient) Errors() <-chan error {
	return c.errChan
}

// Returns a channel that is closed when the connection is closed.
func (c *SockJSClient) Done() <-chan struct{} {
	return c.doneChan
}

// Send a ping message to the server
func (c *SockJSClient) SendPing() error {
	return c.ws.WriteMessage(websocket.TextMessage, []byte(`["\n"]`))
}

// Subscribe to a SockJS destination
func (c *SockJSClient) Subscribe(destination string) error {
	message := fmt.Sprintf(
		`["SUBSCRIBE\nid:%s\ndestination:%s\n\n\u0000"]`,
		fmt.Sprintf("%s-%d", c.sessionID, c.msgCounter.Add(1)),
		destination,
	)

	return c.ws.WriteMessage(websocket.TextMessage, []byte(message))
}

// Setup protocol authentication, heartbeat and return the session ID
func (c *SockJSClient) handleConnect() (string, error) {
	c.ws.ReadMessage() // initial o

	// Send CONNECT message
	c.ws.WriteMessage(websocket.TextMessage, []byte(`["CONNECT\naccept-version:1.2\nheart-beat:10000,10000\n\n\u0000"]`))
	messageType, data, err := c.ws.ReadMessage()
	if err != nil {
		return "", err
	}
	if messageType != websocket.TextMessage {
		return "", fmt.Errorf("unexpected message type %d", messageType)
	}

	// Parse session ID
	msg, err := ParseSockJSMessage(string(data))
	if err != nil {
		return "", err
	}
	sessionID, ok := msg.Headers["session"]
	if !ok {
		return "", fmt.Errorf("session ID not found in CONNECT message")
	}

	// Parse heartbeat
	heartbeat, ok := msg.Headers["heart-beat"]
	if !ok {
		return "", fmt.Errorf("heartbeat not found in CONNECT message")
	}
	heartbeatParts := strings.Split(heartbeat, ",")
	if len(heartbeatParts) != 2 {
		return "", fmt.Errorf("invalid heartbeat format")
	}
	heartbeatSendInterval, err := strconv.Atoi(heartbeatParts[0])
	if err != nil {
		return "", err
	}

	// Start heartbeat
	go func() {
		interval := time.NewTicker(time.Duration(heartbeatSendInterval) * time.Millisecond)
		defer interval.Stop()

		for {
			select {
			case <-c.Done():
				return
			case <-interval.C:
				c.SendPing()
			}
		}
	}()

	return sessionID, nil
}

func (c *SockJSClient) handleProtocol() {
	for {
		messageType, data, err := c.ws.ReadMessage()
		c.mtx.Lock()
		if err != nil {
			if c.isClosed {
				c.mtx.Unlock()
				return
			}

			c.errChan <- err
			c.mtx.Unlock()
			continue
		}

		if messageType != websocket.TextMessage {
			c.mtx.Unlock()
			continue
		}

		// Parse message
		msg, err := ParseSockJSMessage(string(data))
		if err != nil {
			c.errChan <- err
			c.mtx.Unlock()
			continue
		}

		if msg != nil {
			c.msgChan <- msg
		}
		c.mtx.Unlock()
	}
}
