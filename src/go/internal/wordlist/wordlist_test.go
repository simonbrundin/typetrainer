package wordlist

import (
	"testing"
)

func TestNew(t *testing.T) {
	wl := New()

	if wl == nil {
		t.Fatal("New() returned nil")
	}
	if wl.words == nil {
		t.Error("words should not be nil")
	}
	if wl.bigramIndex == nil {
		t.Error("bigramIndex should not be nil")
	}
	if len(wl.words) == 0 {
		t.Error("words should not be empty")
	}
}

func TestWordList_Len(t *testing.T) {
	wl := New()

	got := wl.Len()
	if got == 0 {
		t.Error("Len() should not return 0")
	}
}

func TestWordList_RandomWord(t *testing.T) {
	wl := New()

	// Get multiple random words - should get different results (statistically)
	different := false
	first := wl.RandomWord()

	for i := 0; i < 100; i++ {
		w := wl.RandomWord()
		if w.Text != first.Text {
			different = true
			break
		}
	}

	if !different {
		t.Log("Warning: RandomWord() returned same word 100 times (unlikely but possible)")
	}
}

func TestWordList_RandomWords(t *testing.T) {
	wl := New()

	words := wl.RandomWords(10)

	if len(words) != 10 {
		t.Errorf("RandomWords(10) returned %d words, want 10", len(words))
	}

	// All words should be non-nil
	for i, w := range words {
		if w == nil {
			t.Errorf("RandomWords(10)[%d] is nil", i)
		}
	}
}

func TestWordList_WordsWithBigram(t *testing.T) {
	wl := New()

	// Test with a common bigram that should exist in Swedish
	words := wl.WordsWithBigram("st")

	// "st" is common in Swedish (system, sjuk, etc.)
	if len(words) == 0 {
		t.Error("WordsWithBigram(\"st\") returned empty, expected some words")
	}

	// Verify all returned words contain the bigram
	for _, w := range words {
		found := false
		runes := []rune(w.Text)
		for i := 0; i < len(runes)-1; i++ {
			bg := string(runes[i]) + string(runes[i+1])
			if bg == "st" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Word %q does not contain bigram 'st'", w.Text)
		}
	}
}

func TestWordList_WordsWithBigram_NotFound(t *testing.T) {
	wl := New()

	words := wl.WordsWithBigram("xz")

	if words != nil && len(words) > 0 {
		t.Errorf("WordsWithBigram(\"xz\") returned %d words, want nil or empty", len(words))
	}
}

func TestWordList_AdaptiveSelection_NoWeakBigrams(t *testing.T) {
	wl := New()

	// When no weak bigrams, should return random words
	words := wl.AdaptiveSelection(nil, 10)

	if len(words) != 10 {
		t.Errorf("AdaptiveSelection(nil, 10) returned %d words, want 10", len(words))
	}
}

func TestWordList_AdaptiveSelection_WithWeakBigrams(t *testing.T) {
	wl := New()

	// Get words containing specific bigrams
	words := wl.AdaptiveSelection([]string{"st", "an"}, 20)

	if len(words) == 0 {
		t.Error("AdaptiveSelection with bigrams should return some words")
	}

	// Most words should contain at least one of the weak bigrams
	weakSet := map[string]bool{"st": true, "an": true}
	containingWeak := 0

	for _, w := range words {
		runes := []rune(w.Text)
		for i := 0; i < len(runes)-1; i++ {
			bg := string(runes[i]) + string(runes[i+1])
			if weakSet[bg] {
				containingWeak++
				break
			}
		}
	}

	// At least some words should contain weak bigrams
	if containingWeak == 0 {
		t.Error("AdaptiveSelection should return words containing weak bigrams")
	}
}

func TestWordList_AdaptiveSelection_Count(t *testing.T) {
	wl := New()

	tests := []int{5, 10, 20, 50}

	for _, n := range tests {
		words := wl.AdaptiveSelection(nil, n)
		if len(words) != n {
			t.Errorf("AdaptiveSelection(nil, %d) returned %d words, want %d", n, len(words), n)
		}
	}
}

func TestWordList_AdaptiveSelection_NoDuplicates(t *testing.T) {
	wl := New()

	words := wl.AdaptiveSelection(nil, 20)

	// Check for duplicates
	seen := make(map[*Word]bool)
	for _, w := range words {
		if seen[w] {
			t.Error("AdaptiveSelection returned duplicate words")
		}
		seen[w] = true
	}
}

func TestGlobal(t *testing.T) {
	w1 := Global()
	w2 := Global()

	if w1 != w2 {
		t.Error("Global() should return the same instance")
	}
}
