# TypeTrainer - Swedish Typing Practice

## Overview

A CLI-based typing trainer focused on Swedish text that adapts to the user's weaknesses by tracking and prioritizing practice on difficult bigram (two-letter) combinations.

## Core Concept

Most typing errors and slowness occur on specific **bigrams** (letter pairs) like "sj", "er", "st", "an" rather than individual letters. By tracking error rates and timing for each bigram, the trainer can generate personalized practice sessions that focus on problem areas.

## Architecture

```
typetrainer/
├── src/go/
│   ├── main.go                    # CLI entry point (cobra)
│   └── internal/
│       ├── bigram/bigram.go       # Bigram tracking & analysis
│       ├── wordlist/wordlist.go   # Word list with bigram metadata
│       ├── stats/stats.go         # Statistics & history
│       └── session/session.go      # TUI session logic (bubbletea)
├── words/words.txt                # Swedish word list
└── SPEC.md
```

## Data Model

### BigramStats
```go
type Stats struct {
    Count   uint32 // Times encountered
    TotalMs uint64 // Total time spent
    Errors  uint32 // Number of errors
}
```

### Session
```go
type Session struct {
    Words          []*Word
    CurrentWordIdx int
    CurrentWord    WordResult
    WordResults    []WordResult
    StartedAt      time.Time
    TotalErrors    int
    Bigrams        *bigram.Database
}
```

### GlobalStats
```go
type Stats struct {
    Bigrams      *bigram.Database
    Sessions     []SessionSummary
    PersonalBest map[string]float64
}
```

## Features

### 1. Adaptive Training Mode
- Analyzes user's weak bigrams (high error rate OR high avg time)
- Generates word lists prioritizing these bigrams
- Weights selection by difficulty score
- Shows current word highlight + next words preview

### 2. Practice Modes

**`typetrainer adaptive`** (default)
- Prioritizes weak bigrams
- Shows WPM + accuracy live
- Session ends after N words

**`typetrainer random`**
- Completely random word selection
- Good for warming up

**`typetrainer sprint [seconds]`**
- Timed challenge (default 60s)
- Shows live WPM counter and countdown
- End-of-session summary

**`typetrainer practice <bigram>`**
- Focus on one specific bigram
- Shows all words containing that bigram
- Tracks improvement over time

**`typetrainer stats`**
- Shows overall statistics
- Lists worst 5 bigrams
- Shows WPM trend

**`typetrainer reset`**
- Reset all statistics (with confirmation)

### 3. Real-time TUI
```
╔══════════════════════════════════════════╗
║          TYPE TRAINER - Adaptiv träning  ║
╠══════════════════════════════════════════╣
║  Skriv orden som visas.                 ║
║  Tryck Ctrl+C för att avsluta.          ║
╚══════════════════════════════════════════╝


  Ord: vikt[igast   • försiktigt  • verkligen


  WPM: 45.2  │  Accuracy: 94.1%  │  Errors: 3


  [████████████░░░░░░░░░░░░░░░░░░] 8/20
```

- Green/red highlight for correct/incorrect keystrokes
- Lipgloss styling with colors
- Bubbletea framework for TUI

### 4. Word List

Swedish word list (~1000 words) with:
- Most common Swedish words by frequency
- Bigram index for fast filtering
- Embedded in binary

## Configuration

**Location:** `~/.config/typetrainer/config.json`

```json
{
  "words_per_session": 20,
  "show_next_words": 3,
  "sprint_duration": 60,
  "bigram": {
    "min_samples": 10,
    "weak_threshold": 0.15,
    "time_weight": 0.3,
    "error_weight": 0.7
  }
}
```

**Location:** `~/.config/typetrainer/stats.json`

Persistent bigram statistics and session history.

## Algorithm: Adaptive Word Selection

```go
func (wl *WordList) AdaptiveSelection(weakBigrams []string, count int) []*Word {
    // 1. Get all words containing weak bigrams
    candidates := wl.wordsWithAnyBigram(weakBigrams)

    // 2. Score by number of weak bigrams
    for _, w := range candidates {
        score := countBigrams(w, weakBigrams)
    }

    // 3. Weighted random selection favoring harder words
    // But still including some easier words to not frustrate
    weightedRandomPick(scoredWords, count, hardBias: 0.7)
}
```

## Technical Stack

- **Language:** Go 1.21+
- **CLI:** `cobra` (v1.8)
- **TUI:** `bubbletea` + `lipgloss`
- **Serialization:** `encoding/json`
- **Directory config:** `os.UserHomeDir` + `filepath`

## Installation

```bash
cd src/go
go build -o typetrainer .
cp typetrainer ~/.local/bin/
```

## Usage Examples

```bash
# Start adaptive training (default)
typetrainer
typetrainer adaptive

# 30-second sprint
typetrainer sprint 30

# Practice the "sj" bigram
typetrainer practice sj

# View your statistics
typetrainer stats

# Random words for warmup
typetrainer random

# Reset all stats
typetrainer reset
```

## Polish Details

- Ctrl+C to quit gracefully
- Show encouragement messages at session end
- Track "personal best" WPM
- Color-coded difficulty indicators in stats
- Statistics persist between sessions

## Future Enhancements (Out of Scope for V1)

- Custom text/paste mode
- Multi-user profiles
- Haptic/visual keyboard diagram
- Sentence practice mode
- Export statistics to CSV
- Integration with online leaderboards
