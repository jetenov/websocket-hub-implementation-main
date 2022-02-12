package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

type ServerPool struct {
	Sources []string
	Hub     *hub
	current uint64
}

func (s *ServerPool) nextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.Sources)))
}

func (s *ServerPool) getNextPeer() string {
	next := s.nextIndex()
	l := len(s.Sources) + next

	for i := next; i < l; i++ {
		idx := i % len(s.Sources)

		if i != next {
			atomic.StoreUint64(&s.current, uint64(idx))
		}
		return s.Sources[idx]
	}
	return ""
}

func (s *ServerPool) Run() error {
	ticker := time.NewTicker(getRandInterval() * time.Millisecond)
	for range ticker.C {
		if s.Hub.Empty() {
			continue
		}
		url := s.getNextPeer()
		resp, err := http.Get(url)
		if err != nil {
			return err
		} else if resp.StatusCode != http.StatusOK {
			continue
		}

		message, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		newInterval := getRandInterval() * time.Millisecond
		s.Hub.broadcast <- message
		log.Println(url, newInterval, string(message))
		ticker.Reset(newInterval)
		resp.Body.Close()
	}
	return nil
}

const (
	min = 1000
	max = 3000
)

func getRandInterval() time.Duration {
	return time.Duration(rand.Intn(max-min+1) + min)
}
