package asio

import (
	"bufio"
	"fmt"
	"os"
)

type Session struct {
	DriverName string
	SampleRate float64
	IOHandler  func(in, out [][]int32)
	WaitFunc   func()
}

func (s Session) Run() error {

	if s.IOHandler == nil {
		return fmt.Errorf("IOHandler must be provided")
	}
	if s.DriverName == "" {
		s.DriverName = "ASIO4ALL v2"
	}
	if s.SampleRate == 0 {
		s.SampleRate = 44100
	}
	if s.WaitFunc == nil {
		s.WaitFunc = func() {
			fmt.Println("press enter to continue...")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}
	}

	d := Device{}

	if err := d.Load(s.DriverName); err != nil {
		return err
	}
	defer d.Unload()

	if err := d.SetSampleRate(s.SampleRate); err != nil {
		return err
	}
	if err := d.Open(); err != nil {
		return err
	}

	if err := d.Start(s.IOHandler); err != nil {
		return err
	}

	s.WaitFunc()

	if err := d.Stop(); err != nil {
		return err
	}

	if err := d.Close(); err != nil {
		return err
	}

	return nil
}
