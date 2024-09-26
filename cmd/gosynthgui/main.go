package main

// This binary simply tests the module. though i suppose you could use it as a standalone cli tool if so desired.

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/TH3-F001/gosynth/internal/interfaces"
	"github.com/TH3-F001/gosynth/pkg/notes"
	"github.com/TH3-F001/gosynth/pkg/oscillators"
	"github.com/TH3-F001/gosynth/pkg/synthengine"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	winWidth  = 800
	winHeight = 600
)

var (
	mutex    sync.Mutex
	waveform []int16
)

// #region GUI
func sdlInit() (*sdl.Window, *sdl.Renderer, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, nil, err
	}

	// Get the current display bounds
	displayBounds, err := sdl.GetDisplayBounds(0)
	if err != nil {
		return nil, nil, err
	}

	screenWidth := displayBounds.W
	var xPos int32 = screenWidth / 2
	var yPos int32 = 0

	window, err := sdl.CreateWindow("GoSynth Visualizer", xPos, yPos,
		winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, nil, err
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, nil, err
	}

	return window, renderer, nil
}

func drawWaveform(renderer *sdl.Renderer, waveform []int16, xOffset int) {
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()
	renderer.SetDrawColor(0, 255, 196, 255)

	numSamples := len(waveform)
	if numSamples == 0 {
		return
	}
	xScale := float32(winWidth) / float32(numSamples)
	for i := 0; i < numSamples-1; i++ {
		idx1 := (i + xOffset) % numSamples
		idx2 := (i + 1 + xOffset) % numSamples
		x1 := int32(float32(i) * xScale)
		x2 := int32(float32(i+1) * xScale)
		y1 := int32(winHeight/2 - (int(waveform[idx1]) * winHeight / 32768))
		y2 := int32(winHeight/2 - (int(waveform[idx2]) * winHeight / 32768))
		renderer.DrawLine(x1, y1, x2, y2)
	}
	renderer.Present()
}

func sdlLoop(synth *synthengine.SynthEngine, renderer *sdl.Renderer) {
	running := true
	const bufferSize = winWidth * 2 // Adjust buffer size as needed
	waveformBuffer := make([]int16, 0, bufferSize)
	for i := 0; i < bufferSize; i++ {
		waveformBuffer = append(waveformBuffer, 0)
	}
	var xOffset int
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					switch e.Keysym.Sym {
					case sdl.K_ESCAPE:
						running = false
					}
				}
			}
		}

		// Get new samples and update the buffer
		waveform := synth.GetWaveFormSamples()

		// Handle case where waveform is larger than bufferSize
		if len(waveform) >= bufferSize {
			// Keep only the last bufferSize samples
			waveformBuffer = waveform[len(waveform)-bufferSize:]
		} else {
			// Ensure waveformBuffer doesn't exceed bufferSize
			totalLength := len(waveformBuffer) + len(waveform)
			if totalLength > bufferSize {
				excess := totalLength - bufferSize
				// Ensure excess does not exceed len(waveformBuffer)
				if excess > len(waveformBuffer) {
					waveformBuffer = waveformBuffer[:0]
				} else {
					waveformBuffer = waveformBuffer[excess:]
				}
			}
			waveformBuffer = append(waveformBuffer, waveform...)
		}

		// Increment xOffset to scroll the waveform
		xOffset = (xOffset + 1) % len(waveformBuffer)

		drawWaveform(renderer, waveformBuffer, xOffset)
		sdl.Delay(16) // Approximate 60 FPS
	}
}

// #endregion
func main() {
	// establish I/O for testing
	window, renderer, err := sdlInit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize SDL: %s\n", err)
		return
	}
	defer window.Destroy()
	defer renderer.Destroy()
	defer sdl.Quit()

	// Create a new synth engine with 48000 Sample rate,  2 channels, buffer of 512 and a volume of 0.3
	engine, err := synthengine.NewSynthEngine(48000, 2, 512, 0.08)
	if err != nil {
		log.Fatal(err)
	}

	defer engine.Terminate()
	var wave interfaces.AudioSource
	// sine := oscillators.NewSineWave(notes.A4, 0.5)
	square := oscillators.NewSquareWave(notes.A2, 0.3)
	wave = square
	engine.SetUserFunction(wave.Sample)
	engine.Start()
	sdlLoop(engine, renderer)

	// engine.PlaySound(time.Second * 1)
	engine.Terminate()
	// os.Exit(0)
}
