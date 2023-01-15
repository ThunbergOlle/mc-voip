package microphone

import (
	"fmt"

	"github.com/ThunbergOlle/mc-voip/pkg/errorHandler"
	"github.com/gordonklaus/portaudio"
)

func Stream(output *[]int32, ready chan bool) {

	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int32, 8192)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), &in)
	errorHandler.Panic(err)
	defer stream.Close()

	errorHandler.Panic(stream.Start())
	defer stream.Stop()
	ready <- true
	for {
		err := stream.Read()
		if err != nil {
			fmt.Print("Error handling microphone stream: " + err.Error())
			panic(err)
		}
		*output = in
	}

}
