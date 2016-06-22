package main

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"strings"

	"flag"
	"fmt"
	"github.com/secmask/go-redisproto"
	"sync/atomic"
	"time"
)

var (
	tempQueueLength = flag.Int("s", 50000, "Limit length before client get kick out")

	message   = []byte("message")
	subscribe = []byte("subscribe")

	pubCommand int64
	pubRate    int64
)

type Client struct {
	conn              net.Conn
	parser            *redisproto.Parser
	redisWriter       *redisproto.Writer
	broadcast         *BroadcastChannels
	currentSubChannel *Producer
	eventChan         chan interface{}
	doneChan          chan struct{}
	subChanName       string
}

func init() {
	go func() {
		c := time.NewTicker(time.Second)
		for range c.C {
			pubRate = atomic.LoadInt64(&pubCommand)
			atomic.StoreInt64(&pubCommand, 0)
		}
	}()
}

func NewClient(conn net.Conn, bm *BroadcastChannels) *Client {
	buffWriter := bufio.NewWriter(conn)
	return &Client{
		conn:        conn,
		parser:      redisproto.NewParser(conn),
		redisWriter: redisproto.NewWriter(buffWriter),
		broadcast:   bm,
		eventChan:   make(chan interface{}, *tempQueueLength),
		doneChan:    make(chan struct{}),
	}
}

func (c *Client) Close() (err error) {
	if c.currentSubChannel != nil {
		c.currentSubChannel.RemoveReceiver(c)
		c.currentSubChannel = nil
	}
	err = c.conn.Close()
	return
}

func (c *Client) Overflow() {
	log.Printf("Overflow session [%s] : %d\n", c.subChanName, len(c.eventChan))
	close(c.doneChan)
	go c.Close()

}

func (c *Client) handleBroadcastData() {
out:
	for {
		select {
		case _ = <-c.doneChan:
			break out
		case data := <-c.eventChan:
			_, fErr := c.redisWriter.Write(data.([]byte))
			if fErr = c.redisWriter.Flush(); fErr != nil {
				break out
			}
		}
	}
	log.Printf("End push to session %s for client %s\n", c.subChanName, c.conn.RemoteAddr())
	//c.Close()
}

func (c *Client) handleSubscribed(command *redisproto.Command) (err error) {
	c.subChanName = string(command.Get(1))
	if c.subChanName == "" {
		c.redisWriter.WriteError("Channel name cannot empty")
		err = c.redisWriter.Flush()
		return
	}
	c.currentSubChannel = c.broadcast.GetOrCreate(c.subChanName)
	go c.handleBroadcastData()
	c.currentSubChannel.AddReceiver(c)
	var count int64 = 1
	c.redisWriter.WriteObjects([]interface{}{subscribe, command.Get(1), count})
	err = c.redisWriter.Flush()
	return
}

func (c *Client) handlePublish(command *redisproto.Command) (err error) {
	sc := string(command.Get(1))
	data := command.Get(2)
	if sc == "" {
		c.redisWriter.WriteError("Channel name cannot empty")
		err = c.redisWriter.Flush()
		return
	}
	if len(data) == 0 {
		c.redisWriter.WriteError("Empty data")
		err = c.redisWriter.Flush()
		return
	}

	temBuff := bytes.NewBuffer(make([]byte, 0, 4096))
	rWriter := redisproto.NewWriter(temBuff)
	rWriter.WriteObjects([]interface{}{message, command.Get(1), data})
	b := c.broadcast.GetOrCreate(sc)
	b.Send(temBuff.Bytes())
	err = c.redisWriter.WriteInt(1)
	err = c.redisWriter.Flush()
	atomic.AddInt64(&pubCommand, 1)
	return
}

func (c *Client) handleInfo() error {
	cs := c.broadcast.Channels()
	c.redisWriter.WriteBulkString(fmt.Sprintf("channels: %v\nPublishRate: %d\n", cs, pubRate))
	return c.redisWriter.Flush()
}

func (c *Client) Run() {
	for {
		command, err := c.parser.ReadCommand()
		if err != nil {
			_, ok := err.(*redisproto.ProtocolError)
			if ok {
				err = c.redisWriter.WriteError(err.Error())
				err = c.redisWriter.Flush()
			} else {
				err = c.Close()
				break
			}
		}
		cmd := strings.ToUpper(string(command.Get(0)))
		switch cmd {
		case "SUBSCRIBE":
			err = c.handleSubscribed(command)
			if err != nil {
				log.Println("sub error", err)
				return
			}
			break
		case "PUBLISH":
			err = c.handlePublish(command)
			if err != nil {
				log.Println("pub error", err)
				return
			}
		case "QUIT":
			c.Close()
			return
		case "INFO":
			err = c.handleInfo()
			if err != nil {

			}
		case "PING":
			err = c.redisWriter.WriteSimpleString("PONG")
			err = c.redisWriter.Flush()
		default:
			err = c.redisWriter.WriteError(fmt.Sprintf("Command not support [%s]", cmd))
			err = c.redisWriter.Flush()
		}
	}
}
