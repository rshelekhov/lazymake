package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/sahilm/fuzzy"
)

// CustomFilter implements filtering that always shows headers/separators
// and fuzzy-matches actual Target items
func CustomFilter(term string, targets []string) []list.Rank {
	// If no filter term, show all items
	if term == "" {
		ranks := make([]list.Rank, len(targets))
		for i := range targets {
			ranks[i] = list.Rank{Index: i, MatchedIndexes: []int{}}
		}
		return ranks
	}

	var ranks []list.Rank

	// Process each item
	for i, target := range targets {
		if target == "" {
			// Empty FilterValue = header or separator
			// Always include in filtered results
			ranks = append(ranks, list.Rank{
				Index:          i,
				MatchedIndexes: []int{},
			})
		} else {
			// Fuzzy match actual targets
			matches := fuzzy.Find(term, []string{target})
			if len(matches) > 0 {
				ranks = append(ranks, list.Rank{
					Index:          i,
					MatchedIndexes: matches[0].MatchedIndexes,
				})
			}
		}
	}

	return ranks
}
