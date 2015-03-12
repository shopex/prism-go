package prism

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net"
	"strconv"
)

func (me *Client) Notify() (n *Notify, err error) {
	n = &Notify{Client: me}
	err = n.dail()
	return
}

const (
	command_publish byte = 1
	command_consume byte = 2
	command_ack     byte = 3
)

type Notify struct {
	Client *Client
	conn   *websocket.Conn
}

type Delivery struct {
	Key         string      `json:"client_id"`
	App         string      `json:"app"`
	RoutingKey  string      `json:"key"`
	ContentType string      `json:"type"`
	Body        interface{} `json:"body"`
	Time        int32       `json:"time"`
	Tag         int64       `json:"tag"`
	conn        *websocket.Conn
}

func (d *Delivery) Ack() error {
	buf := bytes.NewBuffer([]byte{command_ack})
	buf.WriteString(strconv.FormatInt(d.Tag, 10))
	return d.conn.WriteMessage(1, buf.Bytes())
}

func (n *Notify) dail() (err error) {
	req, err := n.Client.getRequest("GET", "platform/notify", nil)
	tcpcon, _ := net.Dial("tcp", req.URL.Host)
	n.conn, _, err = websocket.NewClient(tcpcon, req.URL, req.Header, 128, 128)
	return
}

func (n *Notify) Consume() (ch chan *Delivery, err error) {
	ch = make(chan *Delivery)
	err = n.conn.WriteMessage(1, []byte{command_consume})

	go func() {
		for {
			_, data, err := n.conn.ReadMessage()
			d := &Delivery{}
			err = json.Unmarshal(data, d)
			d.conn = n.conn
			if err == nil {
				ch <- d
			}
		}
	}()
	return
}

func (n *Notify) encode(v interface{}) (bin []byte) {
	switch v.(type) {
	case []byte:
		bin = v.([]byte)
	case string:
		bin = []byte(v.(string))
	default:
		bin, _ = json.Marshal(v)
	}
	return
}


func (n *Notify) Pub(routingKey, contentType string, body interface{}) (err error) {
	buf := bytes.NewBuffer([]byte{command_publish})

	binary.Write(buf, binary.BigEndian, uint16(len(routingKey)))
	buf.WriteString(routingKey)

	body_bin := n.encode(body)
	binary.Write(buf, binary.BigEndian, uint32(len(body_bin)))
	buf.Write(body_bin)

	binary.Write(buf, binary.BigEndian, uint16(len(contentType)))
	buf.WriteString(contentType)

	err = n.conn.WriteMessage(1, buf.Bytes())
	return
}

func (n *Notify) Close() error {
	return n.conn.Close()
}
