package wordlist

import (
	_ "embed"
	"math/rand"
	"sort"
	"sync"

	"typetrainer/internal/bigram"
)

//go:embed words.txt
var embeddedWords string

// Word represents a word with metadata
type Word struct {
	Text      string
	Frequency int
}

// WordList holds all words with indexing
type WordList struct {
	words       []Word
	bigramIndex map[string][]int // bigram -> word indices
}

// New creates a new word list from embedded data
func New() *WordList {
	wl := &WordList{
		bigramIndex: make(map[string][]int),
	}

	lines := parseLines(embeddedWords)
	wl.words = make([]Word, 0, len(lines))

	for _, line := range lines {
		if parts := splitLine(line); len(parts) == 2 {
			wl.words = append(wl.words, Word{
				Text:      parts[0],
				Frequency: parseInt(parts[1]),
			})
		}
	}

	// Sort by frequency (most common first)
	sort.Slice(wl.words, func(i, j int) bool {
		return wl.words[i].Frequency < wl.words[j].Frequency
	})

	// Build bigram index AFTER sorting so indices are correct
	for i, word := range wl.words {
		for _, bg := range bigram.BigramsInWord(word.Text) {
			wl.bigramIndex[bg] = append(wl.bigramIndex[bg], i)
		}
	}

	return wl
}

func parseLines(s string) []string {
	var lines []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			line := s[start:i]
			if len(line) > 0 {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	// Last line (no trailing newline)
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitLine(line string) []string {
	for i := 0; i < len(line); i++ {
		if line[i] == ',' {
			return []string{line[:i], line[i+1:]}
		}
	}
	return nil
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

// RandomWord returns a random word
func (wl *WordList) RandomWord() *Word {
	return &wl.words[rand.Intn(len(wl.words))]
}

// RandomWords returns n random words
func (wl *WordList) RandomWords(n int) []*Word {
	result := make([]*Word, n)
	for i := range result {
		result[i] = &wl.words[rand.Intn(len(wl.words))]
	}
	return result
}

// WordsWithBigram returns words containing a specific bigram
func (wl *WordList) WordsWithBigram(bigram string) []*Word {
	indices := wl.bigramIndex[bigram]
	if len(indices) == 0 {
		return nil
	}

	result := make([]*Word, len(indices))
	for i, idx := range indices {
		result[i] = &wl.words[idx]
	}
	return result
}

// AdaptiveSelection selects words prioritizing weak bigrams
func (wl *WordList) AdaptiveSelection(weakBigrams []string, count int) []*Word {
	if len(weakBigrams) == 0 {
		return wl.RandomWords(count)
	}

	// Collect all words containing any weak bigram
	seen := make(map[int]bool)
	var candidates []*Word

	// Use set for thread safety
	weakSet := make(map[string]bool)
	for _, bg := range weakBigrams {
		weakSet[bg] = true
	}

	for _, bg := range weakBigrams {
		indices := wl.bigramIndex[bg]
		for _, idx := range indices {
			if !seen[idx] {
				seen[idx] = true
				candidates = append(candidates, &wl.words[idx])
			}
		}
	}

	if len(candidates) == 0 {
		return wl.RandomWords(count)
	}

	// Shuffle candidates first
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	// Score by number of weak bigrams
	type scored struct {
		word  *Word
		score int
	}
	var scoredWords []scored
	for _, w := range candidates {
		score := 0
		for _, bg := range bigram.BigramsInWord(w.Text) {
			if weakSet[bg] {
				score++
			}
		}
		scoredWords = append(scoredWords, scored{w, score})
	}

	// Sort by score descending
	sort.Slice(scoredWords, func(i, j int) bool {
		return scoredWords[i].score > scoredWords[j].score
	})

	// Pick words with variety - don't repeat words in the same session
	used := make(map[*Word]bool)
	result := make([]*Word, 0, count)

	for _, sw := range scoredWords {
		if len(result) >= count {
			break
		}
		if !used[sw.word] {
			used[sw.word] = true
			result = append(result, sw.word)
		}
	}

	// If we need more words, shuffle and add from all words
	if len(result) < count {
		allWords := wl.RandomWords(count - len(result))
		for _, w := range allWords {
			if len(result) >= count {
				break
			}
			if !used[w] {
				used[w] = true
				result = append(result, w)
			}
		}
	}

	return result
}

// Len returns number of words
func (wl *WordList) Len() int {
	return len(wl.words)
}

var once sync.Once
var global *WordList

// Global returns the global word list instance
func Global() *WordList {
	once.Do(func() {
		global = New()
	})
	return global
}
