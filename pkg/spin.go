package pkg

import (
	"fmt"
	"time"
)

var spinChars = `|/-\`

// ISpinner spinner interface to allow more than one implementations.
type ISpinner interface {
	// Start spin
	Start()
	// Stop spin
	Stop()
}

// spinnerTyp uint8
//go:generate stringer -type=spinnerTyp
type spinnerTyp uint8

const (
	Spinner01 spinnerTyp = iota + 1
)

type spinner01 struct {
	i    int
	quit chan struct{}
}

func newSpinner01() *spinner01 {
	return &spinner01{
		i:    0,
		quit: make(chan struct{}),
	}
}

func (s *spinner01) Start() {
	go func() {
		tick := time.NewTicker(100 * time.Millisecond)
		defer tick.Stop()

		for {
			select {
			case <-tick.C:
				s.tick()
			case <-s.quit:
				return
			}
		}
	}()
}

func (s *spinner01) Stop() {
	s.quit <- struct{}{}
}

func (s *spinner01) tick() {
	fmt.Printf("%c \r", spinChars[s.i])
	s.i = (s.i + 1) % len(spinChars)
}

func NewSpinner(typ spinnerTyp) ISpinner {
	switch typ {
	case Spinner01:
		return newSpinner01()
	default:
		return newSpinner01()
	}
}
