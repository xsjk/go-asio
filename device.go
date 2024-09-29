package asio

import (
	"fmt"
	"unsafe"
)

// asioMessage selectors
const (
	kAsioSelectorSupported    = 1 + iota // selector in <value>, returns 1L if supported, 0 otherwise
	kAsioEngineVersion                   // returns engine (host) asio implementation version, 2 or higher
	kAsioResetRequest                    // request driver reset. if accepted, this will close the driver (ASIO_Exit() ) and re-open it again (ASIO_Init() etc). some drivers need to reconfigure for instance when the sample rate changes, or some basic changes have been made in ASIO_ControlPanel(). returns 1L; note the request is merely passed to the application, there is no way to determine if it gets accepted at this time (but it usually will be).
	kAsioBufferSizeChange                // not yet supported, will currently always return 0L. for now, use kAsioResetRequest instead. once implemented, the new buffer size is expected in <value>, and on success returns 1L
	kAsioResyncRequest                   // the driver went out of sync, such that the timestamp is no longer valid. this is a request to re-start the engine and slave devices (sequencer). returns 1 for ok, 0 if not supported.
	kAsioLatenciesChanged                // the drivers latencies have changed. The engine will refetch the latencies.
	kAsioSupportsTimeInfo                // if host returns true here, it will expect the callback bufferSwitchTimeInfo to be called instead of bufferSwitch
	kAsioSupportsTimeCode                //
	kAsioMMCCommand                      // unused - value: number of commands, message points to mmc commands
	kAsioSupportsInputMonitor            // kAsioSupportsXXX return 1 if host supports this
	kAsioSupportsInputGain               // unused and undefined
	kAsioSupportsInputMeter              // unused and undefined
	kAsioSupportsOutputGain              // unused and undefined
	kAsioSupportsOutputMeter             // unused and undefined
	kAsioOverload                        // driver detected an overload
	kAsioNumMessageSelectors
)

type Device struct {
	driver     *ASIODriver
	io_handler func(
		inputChannelData [][]int32,
		outputChannelData [][]int32,
	)
	currentSampleRate float64
}

func (dev *Device) getDriver() (*IASIO, error) {
	if drv := dev.driver.ASIO; drv == nil {
		return nil, fmt.Errorf("driver not loaded")
	} else {
		return drv, nil
	}
}

func (dev *Device) Load(name string) error {

	CoInitialize(0)

	drivers, err := ListDrivers()
	if err != nil {
		return err
	}

	dev.driver = drivers[name]
	if dev.driver == nil {
		return fmt.Errorf("driver not found: %s", name)
	}

	return dev.driver.Open()
}

func (dev *Device) Unload() {

	if dev.driver != nil {
		dev.driver.Close()
	}

	CoUninitialize()
}

func (dev *Device) CanSampleRate(rate float64) error {
	drv, err := dev.getDriver()
	if err != nil {
		return err
	}
	return drv.CanSampleRate(rate)
}

func (dev *Device) GetSampleRate() (float64, error) {
	drv, err := dev.getDriver()
	if err != nil {
		return 0, err
	}
	rate, err := drv.GetSampleRate()
	dev.currentSampleRate = rate
	return rate, err
}

func (dev *Device) SetSampleRate(rate float64) error {
	drv, err := dev.getDriver()
	if err != nil {
		return err
	}
	err = drv.SetSampleRate(rate)
	if err != nil {
		return err
	}
	dev.currentSampleRate = rate
	return nil
}

func (dev *Device) Open() error {
	drv, err := dev.getDriver()
	if err != nil {
		return err
	}

	n_in, n_out, err := drv.GetChannels()
	if err != nil {
		return err
	}
	fmt.Printf("getChannels():        %d, %d\n", n_in, n_out)

	rawInBuffers := make([][]int32, n_in)
	rawOutBuffers := make([][]int32, n_out)

	// getBufferSize
	minSize, maxSize, preferredSize, granularity, err := drv.GetBufferSize()
	if err != nil {
		return err
	}
	fmt.Printf("getBufferSize():      %d, %d, %d, %d\n", minSize, maxSize, preferredSize, granularity)

	// canSampleRate

	// getChannelInfo (for N)
	bufferDescriptors := make([]BufferInfo, 0, n_in+n_out)
	for i := range n_in {
		bufferDescriptors = append(bufferDescriptors, BufferInfo{
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
	for i := range n_out {
		bufferDescriptors = append(bufferDescriptors, BufferInfo{
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

	// use minSize as buffer size for the lowest latency
	bufferSize := preferredSize

	// createBuffers (set callbacks)
	err = drv.CreateBuffers(bufferDescriptors, bufferSize, Callbacks{
		BufferSwitch: func(doubleBufferIndex int32, directProcess bool) {

			for i := range n_in {
				rawInBuffers[i] = (*[(1 << 48) - 1]int32)(unsafe.Pointer(
					bufferDescriptors[i].Buffers[doubleBufferIndex]))[:bufferSize:bufferSize]
			}
			for i := range n_out {
				rawOutBuffers[i] = (*[(1 << 48) - 1]int32)(unsafe.Pointer(
					bufferDescriptors[i+n_in].Buffers[doubleBufferIndex]))[:bufferSize:bufferSize]
			}

			if dev.io_handler != nil {
				dev.io_handler(rawInBuffers, rawOutBuffers)
			}

		},
		SampleRateDidChange: func(rate float64) {
			fmt.Printf("SampleRateDidChange(%f)\n", rate)
		},
		AsioMessage: func(selector, value int32, message uintptr, opt *float64) int32 {
			fmt.Printf("AsioMessage(%d, %d)\n", selector, value)
			switch selector {
			case kAsioSelectorSupported:
				switch value {
				case kAsioResetRequest:
				case kAsioEngineVersion:
				case kAsioResyncRequest:
				case kAsioLatenciesChanged:
				case kAsioSupportsInputMonitor:
				case kAsioOverload:
					return 1
				default:
					return 0
				}
			case kAsioBufferSizeChange:
				fmt.Printf("kAsioBufferSizeChange\n")
				dev.Reset()
				return 1
			case kAsioResetRequest:
				fmt.Printf("kAsioResetRequest\n")
				dev.Reset()
				return 1
			case kAsioResyncRequest:
				fmt.Printf("kAsioResyncRequest\n")
				dev.Reset()
				return 1
			case kAsioLatenciesChanged:
				fmt.Printf("kAsioLatenciesChanged\n")
				return 1
			case kAsioEngineVersion:
				return 2
			case kAsioSupportsTimeInfo:
			case kAsioSupportsTimeCode:
				return 0
			case kAsioOverload:
				fmt.Printf("kAsioOverload\n")
				return 1
			}
			return 0
		},
		BufferSwitchTimeInfo: func(params *ASIOTime, doubleBufferIndex int32, directProcess bool) *ASIOTime {
			fmt.Printf("BufferSwitchTimeInfo(%v, %d, %v)\n", params, doubleBufferIndex, directProcess)
			return nil
		}})

	return err
}

func (dev *Device) Close() error {
	if drv, err := dev.getDriver(); err != nil {
		return err
	} else {
		return drv.DisposeBuffers()
	}
}

func (dev *Device) Reset() error {
	if drv, err := dev.getDriver(); err != nil {
		return err
	} else {
		if err = drv.Stop(); err != nil {
			return err
		}
		if err = drv.DisposeBuffers(); err != nil {
			return err
		}
		if err = dev.Open(); err != nil {
			return err
		}
		if err = drv.Start(); err != nil {
			return err
		}
		return nil
	}
}

func (dev *Device) Start(handler func([][]int32, [][]int32)) error {
	if drv, err := dev.getDriver(); err != nil {
		return err
	} else {
		if handler != nil {
			dev.io_handler = handler
		}
		return drv.Start()
	}
}
func (dev *Device) Stop() error {
	if drv, err := dev.getDriver(); err != nil {
		return err
	} else {
		return drv.Stop()
	}
}
