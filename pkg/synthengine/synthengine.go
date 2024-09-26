package synthengine

import (
	"fmt"
	"sync"
	"time"

	"github.com/hajimehoshi/oto"
)

// SynthEngine ... A 16 bit Digital Synthesizer engine driven by oto
type SynthEngine struct {
	sampleRate     int
	numChannels    int
	bitDepth       int
	bufferSize     int
	context        *oto.Context
	player         *oto.Player
	userSampleFunc func(float64) float64
	globalTime     float64
	mutex          sync.Mutex
	volume         float64

	waveform []int16
}

func NewSynthEngine(sampleRate, channels, bufferSize int, vol float64) (*SynthEngine, error) {
	context, err := oto.NewContext(sampleRate, channels, 2, bufferSize)
	if err != nil {
		return nil, err
	}

	return &SynthEngine{
		sampleRate:  sampleRate,
		numChannels: channels,
		context:     context,
		player:      context.NewPlayer(),
		globalTime:  0,
		volume:      vol,
	}, nil
}

func (engine *SynthEngine) SetUserFunction(f func(t float64) float64) {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	engine.userSampleFunc = f
}

func (engine *SynthEngine) Start() {
	go engine.mainLoop()
}

func (engine *SynthEngine) PlaySound(duration time.Duration) error {
	engine.mutex.Lock()
	sampleFunc := engine.userSampleFunc
	engine.mutex.Unlock()

	if sampleFunc == nil {
		return fmt.Errorf("from SynthEngine.PlaySound: sample function is not set")
	}

	maxInt16Val := 32767.0
	sampleDuration := 1.0 / float64(engine.sampleRate)
	totalSamples := engine.sampleRate * int(duration.Seconds())
	// multiplied by two because we're taking 16-bit samples
	buffer := make([]byte, totalSamples*2*engine.numChannels)
	engine.waveform = engine.waveform[:0]

	for i := 0; i < totalSamples; i++ {
		sampleVal := sampleFunc(engine.globalTime) * engine.volume

		if sampleVal > 1 {
			sampleVal = 1
		} else if sampleVal < -1 {
			sampleVal = -1
		}

		intSample := int16(sampleVal * maxInt16Val)

		engine.mutex.Lock()
		engine.waveform = append(engine.waveform, intSample)
		engine.mutex.Unlock()

		for channel := 0; channel < engine.numChannels; channel++ {
			index := (i*engine.numChannels + channel) * 2
			buffer[index] = byte(intSample)
			buffer[index+1] = byte(intSample >> 8)
		}
		engine.globalTime += sampleDuration
	}
	_, err := engine.player.Write(buffer)
	if err != nil {
		return err
	}

	return nil
}

func (engine *SynthEngine) mainLoop() {
	maxInt16Val := 32767.0
	sampleDuration := 1.0 / float64(engine.sampleRate)
	bufferSize := engine.sampleRate / 10 // Adjust buffer size as needed

	for {
		engine.mutex.Lock()
		sampleFunc := engine.userSampleFunc
		engine.mutex.Unlock()

		if sampleFunc == nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		buffer := make([]byte, bufferSize*2*engine.numChannels)

		engine.mutex.Lock()
		engine.waveform = engine.waveform[:0]
		engine.mutex.Unlock()

		for i := 0; i < bufferSize; i++ {
			sampleVal := sampleFunc(engine.globalTime) * engine.volume

			if sampleVal > 1.0 {
				sampleVal = 1.0
			} else if sampleVal < -1.0 {
				sampleVal = -1.0
			}

			intSample := int16(sampleVal * maxInt16Val)

			engine.mutex.Lock()
			engine.waveform = append(engine.waveform, intSample)
			engine.mutex.Unlock()

			for channel := 0; channel < engine.numChannels; channel++ {
				index := (i*engine.numChannels + channel) * 2
				buffer[index] = byte(intSample)
				buffer[index+1] = byte(intSample >> 8)
			}

			engine.globalTime += sampleDuration
		}
		_, err := engine.player.Write(buffer)
		if err != nil {
			fmt.Printf("Error writing buffer: %v\n", err)
			return
		}
	}
}

func (engine *SynthEngine) Terminate() {
	engine.context.Close()
}

func (engine *SynthEngine) GetWaveFormSamples() []int16 {
	engine.mutex.Lock()
	defer engine.mutex.Unlock()
	return engine.waveform
}
