package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/xsjk/go-asio"
)

func pause() {
	fmt.Println("press enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func main() {
	drivers, err := asio.ListDrivers()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, drv := range drivers {
		fmt.Printf("%s: %s\n", drv.CLSID, drv.Name)
	}

	{
		fmt.Printf("CoInitialize(0)\n")
		asio.CoInitialize(0)

		defer fmt.Printf("CoUninitialize()\n")
		defer asio.CoUninitialize()

		// wait for Windows to load the driver
		time.Sleep(200 * time.Millisecond)

		driver := drivers["ASIO4ALL v2"]

		fmt.Printf("driver.Open()\n")
		if driver.Open() != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer fmt.Printf("driver.Close()\n")
		defer driver.Close()

		drv := driver.ASIO

		fmt.Printf("ASIO4ALL v2 opened.\n")

		fmt.Printf("getDriverName():      '%s'\n", drv.GetDriverName())
		fmt.Printf("getDriverVersion():   %d\n", drv.GetDriverVersion())
		fmt.Printf("getErrorMessage():    '%s'\n", drv.GetErrorMessage())

		// // controlPanel
		// drv.ControlPanel()

		/// ASIO startup procedure:

		// getChannels
		n_in, n_out, err := drv.GetChannels()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("getChannels():        %d, %d\n", n_in, n_out)

		// getBufferSize
		minSize, maxSize, preferredSize, granularity, err := drv.GetBufferSize()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("getBufferSize():      %d, %d, %d, %d\n", minSize, maxSize, preferredSize, granularity)

		// getSampleRate
		srate, err := drv.GetSampleRate()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("getSampleRate():      %v\n", srate)

		// // canSampleRate
		// curSrate := 1.0
		// for srate := 100; srate <= 480000; srate += 100 {
		// 	err := drv.CanSampleRate(float64(srate))
		// 	if err == nil {
		// 		curSrate = float64(srate)
		// 		fmt.Printf("canSampleRate(%d)\n", srate)
		// 	}
		// }

		curSrate := 44100.0

		// SetSampleRate
		err = drv.SetSampleRate(curSrate)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("setSampleRate(%d)\n", int32(curSrate))

		// outputReady
		fmt.Printf("outputReady():        %v\n", drv.OutputReady())

		// getChannelInfo (for N)
		bufferDescriptors := make([]asio.BufferInfo, 0, n_in+n_out)
		for i := 0; i < n_in; i++ {
			bufferDescriptors = append(bufferDescriptors, asio.BufferInfo{
				Channel: i,
				IsInput: true,
			})
			cinfo, err := drv.GetChannelInfo(i, true)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf(" IN%-2d: active=%v, group=%d, type=%d, name=%s\n",
				i+1, cinfo.IsActive, cinfo.ChannelGroup, cinfo.SampleType, cinfo.Name)
		}
		for i := 0; i < n_out; i++ {
			bufferDescriptors = append(bufferDescriptors, asio.BufferInfo{
				Channel: i,
				IsInput: false,
			})
			cinfo, err := drv.GetChannelInfo(i, false)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("OUT%-2d: active=%v, group=%d, type=%d, name=%s\n",
				i+1, cinfo.IsActive, cinfo.ChannelGroup, cinfo.SampleType, cinfo.Name)
		}

		// createBuffers (set callbacks)
		bufferSize := preferredSize
		err = drv.CreateBuffers(bufferDescriptors, bufferSize, asio.Callbacks{
			BufferSwitch: func(doubleBufferIndex int32, directProcess bool) {
				in := (*[(1 << 48) - 1]int32)(unsafe.Pointer(bufferDescriptors[0].Buffers[doubleBufferIndex]))[:bufferSize:bufferSize]
				out := (*[(1 << 48) - 1]int32)(unsafe.Pointer(bufferDescriptors[n_in].Buffers[doubleBufferIndex]))[:bufferSize:bufferSize]

				for i := range bufferSize {
					out[i] = in[i]
				}

			}})

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		defer fmt.Printf("disposeBuffers()\n")
		defer drv.DisposeBuffers()
		fmt.Printf("createBuffers()\n")

		// getLatencies
		latin, latout, err := drv.GetLatencies()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("getLatencies():       %d, %d\n", latin, latout)

		// start
		err = drv.Start()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("start()\n")

		// wait
		pause()

		// stop
		err = drv.Stop()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("stop()\n")

	}
}
