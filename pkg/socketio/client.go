package socketio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Client struct {
	Host         string
	Sid          string
	PingInterval int
	PingTimeout  int
	maxPayload   int
}

func (c *Client) Emit() error {

	return nil
}

func (c *Client) Connect() error {
	if c.Sid == "" {
		return errors.New("client sid is empty")
	}
	if c.Host == "" {
		return errors.New("client host is empty")
	}

	// body is "40"
	body := "40"
	// check the status of the client
	// make a get request to the socket.io url
	url := fmt.Sprintf("http://%s/socket.io/?EIO=4&transport=polling&sid=%s", c.Host, c.Sid)
	res, err := http.Post(url, "text/plain", strings.NewReader(body))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return errors.New("client connection failed when checking SID status")
	}
	// make websocket handshake
	// ws://localhost:3000/socket.io/?EIO=4&transport=websocket&sid=PDkonChbbdzViTCfAAAR
	// upgrade connection
	url = fmt.Sprintf("ws://%s/socket.io/?EIO=4&transport=websocket&sid=%s", c.Host, c.Sid)
	upgradeRequest, err := http.NewRequest("GET", url, strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	upgradeRequest.Header.Add("Connection", "Upgrade")
	upgradeRequest.Header.Add("Upgrade", "websocket")
	upgradeRequest.Header.Add("Sec-WebSocket-Version", "13")

	client := &http.Client{}
	res, err = client.Do(upgradeRequest)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	return nil
}

func attachClientInfo(c *Client, host string) error {
	// get the sid
	socketIOUrl := "http://" + host + "/socket.io/?EIO=4&transport=polling"

	// make a get request to the socket.io url
	res, err := http.Get(socketIOUrl)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// remove the first character from the response body
	// this is the number 0
	// get the body as string
	bodyRes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// remove the first character
	bodyRes = bodyRes[1:]

	log.Printf("Socket io response body: %s", bodyRes)

	// parse the response body as json
	var clientPollingResponse ClientPollingResponse
	err = json.Unmarshal(bodyRes, &clientPollingResponse)
	if err != nil {
		return err
	}
	log.Printf("Socket io response body: %s", clientPollingResponse.Sid)

	// setup the client
	c.Sid = clientPollingResponse.Sid
	c.PingInterval = clientPollingResponse.PingInterval
	c.PingTimeout = clientPollingResponse.PingTimeout
	c.maxPayload = clientPollingResponse.MaxPayload

	return nil
}

func NewClient(host string) (*Client, error) {
	client := Client{
		Host: host,
	}
	// host should not contain the protocol. Check this
	if strings.Contains("://", host) {
		// return error
		return &Client{}, errors.New("host should not contain the protocol")
	}

	err := attachClientInfo(&client, host)
	if err != nil {
		return &Client{}, err
	}
	log.Printf("Socket io client: %s", client.Sid)
	err = client.Connect()
	if err != nil {
		return &Client{}, err
	}
	log.Printf("Client connection OK: %s", client.Sid)
	return &client, nil
}
