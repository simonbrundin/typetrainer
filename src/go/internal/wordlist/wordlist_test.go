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

func TestWordList_NoDuplicateWords(t *testing.T) {
	wl := New()

	seen := make(map[string]bool)
	for _, w := range wl.words {
		if seen[w.Text] {
			t.Errorf("word list contains duplicate word %q", w.Text)
		}
		seen[w.Text] = true
	}
}

func TestWordList_RandomWords_Distinct(t *testing.T) {
	wl := New()

	words := wl.RandomWords(30)

	seen := make(map[string]bool)
	for _, w := range words {
		if seen[w.Text] {
			t.Errorf("RandomWords returned duplicate word %q", w.Text)
		}
		seen[w.Text] = true
	}
}

func TestComputeFamilies_Inflections(t *testing.T) {
	words := []Word{
		{Text: "prata"}, {Text: "pratar"}, {Text: "pratade"}, {Text: "pratat"},
		{Text: "hus"}, {Text: "huset"},
	}

	families := computeFamilies(words)

	same := func(a, b int) bool { return families[a] == families[b] }
	if !same(0, 1) || !same(0, 2) || !same(0, 3) {
		t.Error("inflections of \"prata\" should belong to the same family")
	}
	if same(0, 4) {
		t.Error("\"prata\" and \"hus\" should not belong to the same family")
	}
	if !same(4, 5) {
		t.Error("\"hus\" and \"huset\" should belong to the same family")
	}
}

func TestWordList_AdaptiveSelection_LimitsInflectionFamily(t *testing.T) {
	wl := New()

	// "ra", "ta" och "at" är svaga bigram i många ord som prata/pratar/pratade
	words := wl.AdaptiveSelection([]string{"ra", "ta", "at"}, 20)

	if len(words) != 20 {
		t.Fatalf("AdaptiveSelection returned %d words, want 20", len(words))
	}

	familyCount := make(map[int]int)
	for _, w := range words {
		fam := wl.familyOf(w.index)
		familyCount[fam]++
		if familyCount[fam] > maxPerFamily {
			t.Errorf("family of %q appears %d times, want at most %d", w.Text, familyCount[fam], maxPerFamily)
		}
	}
}

func TestWordList_AdaptiveSelection_NoInflectionCluster(t *testing.T) {
	wl := New()

	words := wl.AdaptiveSelection([]string{"ra", "ta", "at", "pr"}, 20)

	// Räkna hur många ord som tillhör "prata"-familjen
	prataWords := 0
	for _, w := range words {
		for _, fam := range wl.words {
			if fam.Text == "prata" && wl.familyOf(fam.index) == wl.familyOf(w.index) {
				prataWords++
				break
			}
		}
	}

	if prataWords > maxPerFamily {
		t.Errorf("session contains %d words from the \"prata\" family, want at most %d", prataWords, maxPerFamily)
	}
}

func TestWordList_PracticeWords_FamilyCap(t *testing.T) {
	wl := New()

	words := wl.PracticeWords("pr", 1)

	familyCount := make(map[int]int)
	for _, w := range words {
		familyCount[wl.familyOf(w.index)]++
	}
	for w, n := range familyCount {
		if n > 1 {
			t.Errorf("family appears %d times in practice words", n)
		}
		_ = w
	}

	// Alla ord ska fortfarande innehålla bigramet
	for _, w := range words {
		found := false
		runes := []rune(w.Text)
		for i := 0; i < len(runes)-1; i++ {
			if string(runes[i])+string(runes[i+1]) == "pr" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("word %q does not contain bigram 'pr'", w.Text)
		}
	}
}

func TestWordList_AdaptiveSelection_LimitsWordEndings(t *testing.T) {
	wl := New()

	words := wl.AdaptiveSelection([]string{"ra", "ta", "at", "pr"}, 20)

	endingCount := make(map[string]int)
	for _, w := range words {
		endingCount[endingOf(w.Text)]++
	}
	for ending, n := range endingCount {
		if n > maxPerEnding {
			t.Errorf("ending %q appears %d times, want at most %d", ending, n, maxPerEnding)
		}
	}
}

func TestEndingOf(t *testing.T) {
	tests := []struct {
		word string
		want string
	}{
		{"prata", "ta"},
		{"pratat", "at"},
		{"hej", "ej"},
		{"a", "a"},
	}

	for _, tt := range tests {
		if got := endingOf(tt.word); got != tt.want {
			t.Errorf("endingOf(%q) = %q, want %q", tt.word, got, tt.want)
		}
	}
}
