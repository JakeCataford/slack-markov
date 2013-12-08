package main

// This Markov chain code is taken from the "Generating arbitrary text"
// codewalk: http://golang.org/doc/codewalk/markov/
//
// Minor modifications have been made to make it easier to integrate
// with a webserver and to save/load state

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano()) // Seed the random number generator.
}

// Prefix is a Markov chain prefix of one or more words.
type Prefix []string

// String returns the Prefix as a string (for use as a map key).
func (p Prefix) String() string {
	return strings.Join(p, " ")
}

// Shift removes the first word from the Prefix and appends the given word.
func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

// Chain contains a map ("chain") of prefixes to a list of suffixes.
// A prefix is a string of prefixLen words joined with spaces.
// A suffix is a single word. A prefix can have multiple suffixes.
type Chain struct {
	chain     map[string][]string
	prefixLen int
	mu        sync.Mutex
}

// NewChain returns a new Chain with prefixes of prefixLen words.
func NewChain(prefixLen int) *Chain {
	return &Chain{
		chain:     make(map[string][]string),
		prefixLen: prefixLen,
	}
}

// Write parses the bytes into prefixes and suffixes that are stored in Chain.
func (c *Chain) Write(in string) (int, error) {
	sr := strings.NewReader(in)
	p := make(Prefix, c.prefixLen)
	for {
		var s string
		if _, err := fmt.Fscan(sr, &s); err != nil {
			break
		}
		key := p.String()
		log.Printf("Adding '%s' for key '%s'", s, key)
		c.mu.Lock()
		c.chain[key] = append(c.chain[key], s)
		c.mu.Unlock()
		p.Shift(s)
	}
	return len(in), nil
}

// Generate returns a string of at most n words generated from Chain.
func (c *Chain) Generate(n int) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	p := make(Prefix, c.prefixLen)
	var words []string
	for i := 0; i < n; i++ {
		choices := c.chain[p.String()]
		if len(choices) == 0 {
			break
		}
		next := choices[rand.Intn(len(choices))]
		words = append(words, next)
		p.Shift(next)
	}
	return strings.Join(words, " ")
}
