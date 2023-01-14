package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/ThunbergOlle/mc-voip/internal/microphone"
	"github.com/ThunbergOlle/mc-voip/pkg/errorHandler"
	"github.com/gordonklaus/portaudio"
	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("Hello, World!")
	u := url.URL{Scheme: "ws", Host: "localhost:3000", Path: "/socket.io/?EIO=4&transport=websocket"}

	connection, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("handshake failed with status %d", resp.StatusCode)
		log.Fatal("dial:", err)
	}
	//When the program closes close the connection
	defer connection.Close()
	// send socketio connect message
	err = connection.WriteMessage(websocket.TextMessage, []byte("40"))
	if err != nil {
		log.Fatal("write:", err)
	}

	log.Printf("NewClient success\n")

	killSig := make(chan os.Signal, 1)
	signal.Notify(killSig, os.Interrupt, os.Kill)

	// start the microphone as a goroutine
	out := make([]int32, 128)
	ready := make(chan bool)

	go microphone.Stream(&out, ready)

	// wait for the microphone to be ready
	<-ready

	portaudio.Initialize()

	stream, err := portaudio.OpenDefaultStream(0, 1, 44100, len(out), &out)
	if err != nil {
		fmt.Println("Error opening speaker stream: " + err.Error())
		panic(err)
	}
	errorHandler.Panic(stream.Start())
	for {
		err := stream.Write()

		if err != nil {
			fmt.Println("Error handling speaker stream: " + err.Error())
			panic(err)
		}

		select {
		case <-killSig:
			fmt.Println("Kill signal received")
			return
		default:
		}
	}

}
