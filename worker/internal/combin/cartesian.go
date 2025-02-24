package combin

import (
	"errors"
	"sync"
)

// Global error variables
var (
	errEmptyAlphabet        = errors.New("alphabet must not be empty")
	errInvalidMaxLength     = errors.New("maxLength must be positive")
	errInvalidStartIndex    = errors.New("startIndex must be non-negative")
	errStartIndexOutOfRange = errors.New("startIndex exceeds the total number of combinations")
)

// AlphabetIterator iterates over all combinations of strings from the given alphabet
// with lengths ranging from 1 to maxLength.
type AlphabetIterator struct {
	alphabet   string       // Alphabet from which combinations are generated
	maxLength  int          // Maximum length of combinations
	current    []rune       // Current combination
	length     int          // Current length of the combination
	done       bool         // Flag indicating if iteration is complete
	startIndex int          // Starting index for generation
	rw         sync.RWMutex // Mutex for thread-safe access
}

// NewAlphabetIterator creates a new iterator for generating combinations of strings
// from the given alphabet with lengths ranging from 1 to maxLength.
// The `startIndex` parameter specifies the index of the first combination to generate.
func NewAlphabetIterator(alphabet string, maxLength int, startIndex int) (*AlphabetIterator, error) {
	if len(alphabet) == 0 {
		return nil, errEmptyAlphabet
	}
	if maxLength <= 0 {
		return nil, errInvalidMaxLength
	}
	if startIndex < 0 {
		return nil, errInvalidStartIndex
	}

	it := &AlphabetIterator{
		alphabet:   alphabet,
		maxLength:  maxLength,
		current:    make([]rune, maxLength),
		length:     1,
		done:       false,
		startIndex: startIndex,
	}

	// Fast-forward the iterator to the starting index
	for i := 0; i < startIndex; i++ {
		if !it.Next() {
			return nil, errStartIndexOutOfRange
		}
	}

	return it, nil
}

// Next moves to the next combination. It returns true if there is a next combination,
// and false if the iteration is complete.
func (it *AlphabetIterator) Next() bool {
	it.rw.Lock()
	defer it.rw.Unlock()

	if it.done {
		return false
	}

	// If the current combination is not initialized, start with the first character
	if it.current[0] == 0 {
		it.current[0] = rune(it.alphabet[0])
		return true
	}

	// Iterate through the characters in the current combination
	for i := it.length - 1; i >= 0; i-- {
		// Find the next character in the alphabet
		nextIndex := indexOf(it.alphabet, byte(it.current[i]))
		if nextIndex+1 < len(it.alphabet) {
			it.current[i] = rune(it.alphabet[nextIndex+1])
			return true
		} else {
			// If the end of the alphabet is reached, reset the character and move to the previous one
			it.current[i] = rune(it.alphabet[0])
		}
	}

	// If all combinations of the current length are exhausted, increase the length
	if it.length < it.maxLength {
		it.length++
		for i := 0; i < it.length; i++ {
			it.current[i] = rune(it.alphabet[0])
		}
		return true
	}

	// If the maximum length is reached, mark the iteration as complete
	it.done = true
	return false
}

// Current returns the current combination as a string.
func (it *AlphabetIterator) Current() string {
	it.rw.RLock()
	defer it.rw.RUnlock()

	return string(it.current[:it.length])
}

// Helper function to find the index of a character in a string
func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
