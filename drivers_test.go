package asio

import (
	"fmt"
	"testing"
	"time"
)

func TestListDrivers(t *testing.T) {
	drivers, err := ListDrivers()
	if err != nil {
		t.Error(err)
		return
	}

	for _, drv := range drivers {
		fmt.Printf("%s: %s\n", drv.CLSID, drv.Name)
	}

	// Use the first driver for testing.
	var driver *ASIODriver
	for _, drv := range drivers {
		driver = drv
		break
	}
	if driver == nil {
		t.Error("No driver found")
		return
	}

	{
		fmt.Printf("CoInitialize(0)\n")
		CoInitialize(0)
		defer fmt.Printf("CoUninitialize()\n")
		defer CoUninitialize()

		fmt.Printf("driver.Open()\n")
		err = driver.Open()
		defer fmt.Printf("driver.Close()\n")
		defer driver.Close()
		if err != nil {
			t.Error(err)
			return
		}

		drv := driver.ASIO

		fmt.Printf("%s opened.\n", driver.Name)

		fmt.Printf("getDriverName():      '%s'\n", drv.GetDriverName())
		fmt.Printf("getDriverVersion():   %d\n", drv.GetDriverVersion())
		fmt.Printf("getErrorMessage():    '%s'\n", drv.GetErrorMessage())

		// ASIO startup procedure:

		// getChannels
		in, out, err := drv.GetChannels()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("getChannels():        %d, %d\n", in, out)

		// getBufferSize
		minSize, maxSize, preferredSize, granularity, err := drv.GetBufferSize()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("getBufferSize():      %d, %d, %d, %d\n", minSize, maxSize, preferredSize, granularity)

		// canSampleRate
		ok := drv.CanSampleRate(44100.)
		fmt.Printf("canSampleRate(44100): %v\n", ok)

		// getSampleRate
		srate, err := drv.GetSampleRate()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("getSampleRate():      %v\n", srate)

		// SetSampleRate
		err = drv.SetSampleRate(44100.)
		fmt.Printf("setSampleRate(44100): %v\n", err)

		// outputReady
		fmt.Printf("outputReady():        %v\n", drv.OutputReady())

		// open control panel:
		//drv.ControlPanel()

		// getChannelInfo (for N)
		bufferDescriptors := make([]BufferInfo, 0, in+out)
		for i := 0; i < in; i++ {
			bufferDescriptors = append(bufferDescriptors, BufferInfo{
				Channel: i,
				IsInput: true,
			})
			cinfo, err := drv.GetChannelInfo(i, true)
			if err != nil {
				t.Error(err)
				continue
			}
			fmt.Printf(" IN%-2d: active=%v, group=%d, type=%d, name=%s\n", i+1, cinfo.IsActive, cinfo.ChannelGroup, cinfo.SampleType, cinfo.Name)
		}
		for i := 0; i < out; i++ {
			bufferDescriptors = append(bufferDescriptors, BufferInfo{
				Channel: i,
				IsInput: false,
			})
			cinfo, err := drv.GetChannelInfo(i, false)
			if err != nil {
				t.Error(err)
				continue
			}
			fmt.Printf("OUT%-2d: active=%v, group=%d, type=%d, name=%s\n", i+1, cinfo.IsActive, cinfo.ChannelGroup, cinfo.SampleType, cinfo.Name)
		}

		// createBuffers (set callbacks)
		err = drv.CreateBuffers(bufferDescriptors, preferredSize, Callbacks{
			BufferSwitch: func(doubleBufferIndex int32, directProcess bool) {
				fmt.Printf("BufferSwitch(%d, %v)\n", doubleBufferIndex, directProcess)
			},
		})
		if err != nil {
			t.Error(err)
			return
		}
		defer fmt.Printf("disposeBuffers()\n")
		defer drv.DisposeBuffers()
		fmt.Printf("createBuffers()\n")

		// getLatencies
		latin, latout, err := drv.GetLatencies()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("getLatencies():       %d, %d\n", latin, latout)

		// start
		err = drv.Start()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("start()\n")

		// wait
		time.Sleep(100 * time.Millisecond)

		// stop
		err = drv.Stop()
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("stop()\n")

		// disposeBuffers
		// obj->Release()
	}
}
