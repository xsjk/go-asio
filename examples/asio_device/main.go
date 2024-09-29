package main

import (
	"bufio"
	"os"

	"github.com/xsjk/go-asio"
)

func pause() {
	println("press enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func main() {
	device := asio.Device{}

	device.Load("ASIO4ALL v2")
	device.Open()
	device.Start(func(in, out [][]int32) {
		for i := range out[0] {
			out[0][i] = in[0][i]
		}
	})
	pause()
	device.Stop()
	device.Close()
	device.Unload()
}
