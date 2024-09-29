package main

import (
	"github.com/xsjk/go-asio"
)

func main() {
	asio.Session{
		IOHandler: func(in, out [][]int32) {
			for i := range out[0] {
				out[0][i] = in[0][i]
			}
		},
	}.Run()

}
