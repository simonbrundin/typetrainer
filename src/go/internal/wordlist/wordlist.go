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
	index     int // internt index i ordlistan
}

// familyPrefixLen är längden på det gemensamma prefix som krävs för att två
// ord ska räknas som samma ordfamilj (t.ex. "prata", "pratar", "pratade").
const familyPrefixLen = 4

// maxPerFamily är maximalt antal ord per inflectionsfamilj i en session.
const maxPerFamily = 1

// maxPerEnding är maximalt antal ord som slutar med samma bokstavpar i en
// session, så att inte hela sessionen består av t.ex. ord på "-at".
const maxPerEnding = 3

// WordList holds all words with indexing
type WordList struct {
	words       []Word
	bigramIndex map[string][]int // bigram -> word indices
	family      []int            // ordindex -> familj-id (inflectionsgruppering)
}

// New creates a new word list from embedded data
func New() *WordList {
	wl := &WordList{
		bigramIndex: make(map[string][]int),
	}

	lines := parseLines(embeddedWords)
	wl.words = make([]Word, 0, len(lines))

	seen := make(map[string]bool)
	for _, line := range lines {
		if parts := splitLine(line); len(parts) == 2 {
			// Hoppa över dubbletter - samma ord ska bara finnas en gång
			if seen[parts[0]] {
				continue
			}
			seen[parts[0]] = true
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

	for i := range wl.words {
		wl.words[i].index = i
	}

	// Build bigram index AFTER sorting so indices are correct
	for i, word := range wl.words {
		for _, bg := range bigram.BigramsInWord(word.Text) {
			wl.bigramIndex[bg] = append(wl.bigramIndex[bg], i)
		}
	}

	wl.family = computeFamilies(wl.words)

	return wl
}

// computeFamilies grupperar böjningar av samma ord (prata/pratar/pratade).
// Ord hamnar i samma familj om de delar ett prefix på minst familyPrefixLen
// tecken, eller om det ena ordet är själva prefixet till det andra.
func computeFamilies(words []Word) []int {
	order := make([]int, len(words))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool {
		return words[order[a]].Text < words[order[b]].Text
	})

	family := make([]int, len(words))
	current := 0
	for pos, idx := range order {
		if pos > 0 {
			prev := order[pos-1]
			if sameFamily(words[prev].Text, words[idx].Text) {
				current = family[prev]
			} else {
				current++
			}
		}
		family[idx] = current
	}
	return family
}

// sameFamily avgör om två ord hör till samma infaktionsfamilj
func sameFamily(a, b string) bool {
	n := commonPrefixLen(a, b)
	if n >= familyPrefixLen {
		return true
	}
	// Det ena ordet är ett prefix till det andra ("hus" / "huset")
	shortest := len([]rune(a))
	if l := len([]rune(b)); l < shortest {
		shortest = l
	}
	return n == shortest && shortest >= 3
}

// commonPrefixLen returnerar längden på det gemensamma prefixet mellan två ord
func commonPrefixLen(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	n := len(ra)
	if len(rb) < n {
		n = len(rb)
	}
	for i := 0; i < n; i++ {
		if ra[i] != rb[i] {
			return i
		}
	}
	return n
}

// endingOf returnerar det sista bokstavparet i ett ord
func endingOf(text string) string {
	runes := []rune(text)
	if len(runes) < 2 {
		return text
	}
	return string(runes[len(runes)-2:])
}

// familyOf returnerar familj-id för ett ordindex
func (wl *WordList) familyOf(idx int) int {
	if idx < 0 || idx >= len(wl.family) {
		return -1
	}
	return wl.family[idx]
}

func parseLines(s string) []string {
	var lines []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			line := s[start:i]
			// Skip empty lines and comment lines
			if len(line) > 0 && line[0] != '#' {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	// Last line (no trailing newline)
	if start < len(s) {
		line := s[start:]
		if len(line) > 0 && line[0] != '#' {
			lines = append(lines, line)
		}
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

// RandomWords returns n distinct random words
func (wl *WordList) RandomWords(n int) []*Word {
	if n <= 0 || len(wl.words) == 0 {
		return nil
	}
	if n > len(wl.words) {
		n = len(wl.words)
	}

	// Plocka utan återläggning så att samma ord inte upprepas i en session
	indices := rand.Perm(len(wl.words))[:n]
	result := make([]*Word, n)
	for i, idx := range indices {
		result[i] = &wl.words[idx]
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

// PracticeWords returns words containing a bigram, but with at most maxPerFamily
// ord per inflectionsfamilj (t.ex. inte prata, pratar och pratade samtidigt).
func (wl *WordList) PracticeWords(bigram string, maxPerFamily int) []*Word {
	all := wl.WordsWithBigram(bigram)
	if maxPerFamily <= 0 {
		return all
	}

	familyCount := make(map[int]int)
	var result []*Word
	for _, w := range all {
		fam := wl.familyOf(w.index)
		if familyCount[fam] >= maxPerFamily {
			continue
		}
		familyCount[fam]++
		result = append(result, w)
	}

	// Om urvalet blev för litet, fyll på med de övriga orden i samma familj
	if len(result) < len(all) {
		familyCount = make(map[int]int)
		for _, w := range result {
			familyCount[wl.familyOf(w.index)]++
		}
		for _, w := range all {
			if len(result) >= len(all) {
				break
			}
			fam := wl.familyOf(w.index)
			if familyCount[fam] >= maxPerFamily {
				continue
			}
			familyCount[fam]++
			result = append(result, w)
		}
	}

	return result
}

// AdaptiveSelection selects words prioritizing weak bigrams.
//
// Ord väljs genom att gå varvet runt över de svaga bigramen (i ordning från
// svårast till lättast) så att sessionen sprider ut träningen över flera
// problemområden i stället för att köra samma mönster igen och igen.
// Varje ordfamilj (t.ex. prata/pratar/pratade) bidrar med max maxPerFamily ord.
func (wl *WordList) AdaptiveSelection(weakBigrams []string, count int) []*Word {
	if count <= 0 {
		return nil
	}
	if len(weakBigrams) == 0 {
		return wl.RandomWords(count)
	}

	weakSet := make(map[string]bool, len(weakBigrams))
	for _, bg := range weakBigrams {
		weakSet[bg] = true
	}

	// Poäng: antal svaga bigram i ordet (fler = träningsvärdesare)
	score := make(map[int]int)
	for i, w := range wl.words {
		s := 0
		for _, bg := range bigram.BigramsInWord(w.Text) {
			if weakSet[bg] {
				s++
			}
		}
		if s > 0 {
			score[i] = s
		}
	}
	if len(score) == 0 {
		return wl.RandomWords(count)
	}

	// Kandidatlistor per svagt bigram, sorterade efter poäng (fallback slump)
	type candidates struct {
		words []*Word
		pos   int
	}
	lists := make([]candidates, 0, len(weakBigrams))

	for _, bg := range weakBigrams {
		indices := wl.bigramIndex[bg]
		if len(indices) == 0 {
			continue
		}
		ws := make([]*Word, 0, len(indices))
		for _, idx := range indices {
			ws = append(ws, &wl.words[idx])
		}
		rand.Shuffle(len(ws), func(i, j int) { ws[i], ws[j] = ws[j], ws[i] })
		sort.SliceStable(ws, func(i, j int) bool {
			return score[ws[i].index] > score[ws[j].index]
		})
		lists = append(lists, candidates{words: ws})
	}

	if len(lists) == 0 {
		return wl.RandomWords(count)
	}

	used := make(map[int]bool)
	familyCount := make(map[int]int)
	endingCount := make(map[string]int)
	result := make([]*Word, 0, count)

	take := func(w *Word) bool {
		fam := wl.familyOf(w.index)
		if familyCount[fam] >= maxPerFamily {
			return false
		}
		if endingCount[endingOf(w.Text)] >= maxPerEnding {
			return false
		}
		familyCount[fam]++
		endingCount[endingOf(w.Text)]++
		used[w.index] = true
		result = append(result, w)
		return true
	}

	// Varvet runt: ett ord per bigram per omgång
	for len(result) < count {
		progress := false
		for li := range lists {
			if len(result) >= count {
				break
			}
			l := &lists[li]
			for l.pos < len(l.words) {
				w := l.words[l.pos]
				l.pos++
				if used[w.index] {
					continue
				}
				if take(w) {
					progress = true
				}
				break
			}
		}
		if !progress {
			break
		}
	}

	// Fyll på med övriga ord (slumpade) om vi fortfarande behöver fler.
	// Familjgränsen hålls även här.
	if len(result) < count {
		pool := rand.Perm(len(wl.words))
		for _, idx := range pool {
			if len(result) >= count {
				break
			}
			if used[idx] {
				continue
			}
			fam := wl.familyOf(idx)
			if familyCount[fam] >= maxPerFamily {
				continue
			}
			if endingCount[endingOf(wl.words[idx].Text)] >= maxPerEnding {
				continue
			}
			familyCount[fam]++
			endingCount[endingOf(wl.words[idx].Text)]++
			used[idx] = true
			result = append(result, &wl.words[idx])
		}
	}

	// Sista utväg: om vi fortfarande är under målet, ta de återstående orden
	// även om det bryter mot familj-/ändelsereglerna
	if len(result) < count {
		pool := rand.Perm(len(wl.words))
		for _, idx := range pool {
			if len(result) >= count {
				break
			}
			if used[idx] {
				continue
			}
			used[idx] = true
			result = append(result, &wl.words[idx])
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
