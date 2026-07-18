package lexer

const FuzzyThreshold = 0.80

func levenshteinDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)

	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			deletion := prev[j] + 1
			insertion := curr[j-1] + 1
			substitution := prev[j-1] + cost
			curr[j] = min3(deletion, insertion, substitution)
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

func normalizedSimilarity(input, candidate string) float64 {
	dist := levenshteinDistance(input, candidate)

	maxLen := len([]rune(input))
	if cl := len([]rune(candidate)); cl > maxLen {
		maxLen = cl
	}
	if maxLen == 0 {
		return 1.0
	}

	return 1.0 - (float64(dist) / float64(maxLen))
}

type FuzzyMatchResult struct {
	Matched     bool
	Type        TokenType
	MatchedWord string
	Score       float64
}

func LookupFuzzy(word string) FuzzyMatchResult {
	best := FuzzyMatchResult{}

	for _, cand := range fuzzyCandidates {
		score := normalizedSimilarity(word, cand.Word)
		if score > best.Score {
			best = FuzzyMatchResult{
				Matched:     true,
				Type:        cand.Type,
				MatchedWord: cand.Word,
				Score:       score,
			}
		}
	}

	if best.Score < FuzzyThreshold {
		return FuzzyMatchResult{}
	}
	return best
}
