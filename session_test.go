package asio

import (
	"testing"
	"time"
)

func TestSession(t *testing.T) {
	err := Session{
		DriverName: "ASIO4ALL v2",
		SampleRate: 44100,
		IOHandler: func(in, out [][]int32) {
			for i := range out[0] {
				out[0][i] = in[0][i]
			}
		},
		WaitFunc: func() {
			time.Sleep(time.Second)
		},
	}.Run()
	if err != nil {
		t.Error(err)
	}
}
