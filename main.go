package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/ThunbergOlle/mc-voip/internal/microphone"
	"github.com/ThunbergOlle/mc-voip/pkg/errorHandler"
	"github.com/ThunbergOlle/mc-voip/pkg/socketio"
	"github.com/gordonklaus/portaudio"
)

func main() {
	fmt.Println("Hello, World!")

	_, err := socketio.NewClient("localhost:3000")
	errorHandler.Panic(err)

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
