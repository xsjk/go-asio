package asio

import (
	"fmt"
	"testing"
	"time"
)

func TestDevice(t *testing.T) {
	drivers, err := ListDrivers()
	if err != nil {
		t.Error(err)
		return
	}

	for _, drv := range drivers {
		fmt.Printf("%s: %s\n", drv.CLSID, drv.Name)
	}

	device := Device{}

	// load
	if err = device.Load("ASIO4ALL v2"); err != nil {
		t.Error(err)
		return
	}
	defer device.Unload()

	// open
	if err = device.Open(); err != nil {
		t.Error(err)
		return
	}
	defer device.Close()

	if err = device.Start(func(inputChannelData, outputChannelData [][]int32) {
		out_array := outputChannelData[0]
		in_array := inputChannelData[0]
		for i := range out_array {
			out_array[i] = in_array[i]
		}
	}); err != nil {
		t.Error(err)
		return
	}

	time.Sleep(100 * time.Millisecond)

	if err = device.Stop(); err != nil {
		t.Error(err)
		return
	}

}
