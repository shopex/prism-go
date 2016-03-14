package prism

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
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

func (n *Notify) retry() {
	if n.Client == nil {
		log.Printf("can not retry on a nil cilent")
		return
	}

	if n.conn != nil {
		n.Close()
	}
	// 30秒重连
	c := time.Tick(time.Second * 30)

	var err error
	for {
		<-c
		err = n.dail()
		if err != nil {
			log.Printf("reconnect to websocket fail (%s)\n", err)
		}
		return
	}
}

func (n *Notify) dail() (err error) {
	req, err := n.Client.getRequest("GET", "platform/notify", nil)
	tcpcon, err := net.Dial("tcp", req.URL.Host)
	if err != nil {
		return err
	}
	req.URL.Scheme = "ws"
	n.conn, _, err = websocket.NewClient(tcpcon, req.URL, req.Header, 128, 128)
	return
}

func (n *Notify) Consume(topic string) (ch chan *Delivery, err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("(*Notify).Consume meet a panic (%s)\n", err)
			go n.retry()
		}
	}()
	ch = make(chan *Delivery)
	data := []byte{command_consume}
	if topic != "" {
		l := len(topic)
		data = append(data, uint8(l/256), uint8(l%256))
		data = append(data, []byte(topic)...)
	}
	err = n.conn.WriteMessage(1, data)

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
