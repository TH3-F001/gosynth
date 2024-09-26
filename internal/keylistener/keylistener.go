package keylistener

import (
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

type KeyState int

const (
	Off KeyState = iota
	Released
	Pressed
	Held
)

type KeyCallback func()

type KeyListener struct {
	callbacks map[tcell.Key]map[KeyState][]KeyCallback
	mutex     sync.Mutex
	screen    tcell.Screen
	keyStates map[tcell.Key]KeyState
	lastEvent map[tcell.Key]time.Time
	done      chan struct{}
}

func NewKeyListener() (*KeyListener, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}

	kl := &KeyListener{
		callbacks: make(map[tcell.Key]map[KeyState][]KeyCallback),
		screen:    screen,
		keyStates: make(map[tcell.Key]KeyState),
		lastEvent: make(map[tcell.Key]time.Time),
		done:      make(chan struct{}),
	}
	go kl.listen()
	return kl, nil
}

func (kl *KeyListener) Subscribe(key tcell.Key, state KeyState, callback KeyCallback) {
	kl.mutex.Lock()
	defer kl.mutex.Unlock()
	if _, exists := kl.callbacks[key]; !exists {
		kl.callbacks[key] = make(map[KeyState][]KeyCallback)
	}
	kl.callbacks[key][state] = append(kl.callbacks[key][state], callback)
}

func (kl *KeyListener) listen() {
	defer kl.screen.Fini()
	for {
		select {
		case <-kl.done:
			return
		default:
			event := kl.screen.PollEvent()
			switch ev := event.(type) {
			case *tcell.EventKey:
				kl.handleKeyEvent(ev)
			}
		}
	}
}

func (kl *KeyListener) handleKeyEvent(ev *tcell.EventKey) {
	kl.mutex.Lock()
	defer kl.mutex.Unlock()
	key := ev.Key()
	state := kl.keyStates[key]
	var newState KeyState

	now := time.Now()
	lastTime := kl.lastEvent[key]
	kl.lastEvent[key] = now
	if state == Off || state == Released {
		newState = Pressed
	} else if state == Pressed || state == Held {
		if now.Sub(lastTime) < 200*time.Millisecond {
			newState = Held
		} else {
			newState = Pressed
		}
	}
	kl.keyStates[key] = newState

	if callbacks, ok := kl.callbacks[key]; ok {
		if cbs, ok := callbacks[newState]; ok {
			for _, cb := range cbs {
				go cb()
			}
		}
	}
	go kl.checkReleased(key, newState)
}

func (kl *KeyListener) checkReleased(key tcell.Key, state KeyState) {
	time.Sleep(200 * time.Millisecond)
	kl.mutex.Lock()
	defer kl.mutex.Unlock()
	if kl.keyStates[key] == state && time.Since(kl.lastEvent[key]) >= 200*time.Millisecond {
		kl.keyStates[key] = Released
		if callbacks, ok := kl.callbacks[key]; ok {
			if cbs, ok := callbacks[Released]; ok {
				for _, cb := range cbs {
					go cb()
				}
			}
		}
		kl.keyStates[key] = Off
	}
}

func (kl *KeyListener) Stop() {
	close(kl.done)
}
