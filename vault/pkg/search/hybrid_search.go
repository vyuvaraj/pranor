package import (
	"sort"
	"strings"
)

// SearchResultItem represents a ranked object match in hybrid search.
type SearchResultItem struct {
	ObjectID    string                 `json:"object_id"`
	CombinedScore float64              `json:"combined_score"`
	KeywordScore  float64              `json:"keyword_score"`
	VectorScore   float64              `json:"vector_score"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// HybridSearchEngine merges BM25 keyword matching with vector similarity using Reciprocal Rank Fusion (RRF).
type HybridSearchEngine struct{}

// NewHybridSearchEngine creates a HybridSearchEngine instance.
func NewHybridSearchEngine() *HybridSearchEngine {
	return &HybridSearchEngine{}
}

// PerformHybridSearch combines keyword and vector match results using RRF ranking.
func (hse *HybridSearchEngine) PerformHybridSearch(keywordMatches, vectorMatches map[string]float64, kConst float64) []SearchResultItem {
	if kConst <= 0 {
		kConst = 60.0 // RRF standard constant
	}

	// Calculate RRF scores: RRF = 1 / (k + rank_keyword) + 1 / (k + rank_vector)
	type rankItem struct {
		ID    string
		Score float64
	}

	kwRanks := sortByScore(keywordMatches)
	vecRanks := sortByScore(vectorMatches)

	kwRankMap := make(map[string]int)
	for i, item := range kwRanks {
		kwRankMap[item.ID] = i + 1
	}

	vecRankMap := make(map[string]int)
	for i, item := range vecRanks {
		vecRankMap[item.ID] = i + 1
	}

	allIDs := make(map[string]bool)
	for id := range keywordMatches {
		allIDs[id] = true
	}
	for id := range vectorMatches {
		allIDs[id] = true
	}

	var results []SearchResultItem
	for id := range allIDs {
		rrfScore := 0.0
		if rank, ok := kwRankMap[id]; ok {
			rrfScore += 1.0 / (kConst + float64(rank))
		}
		if rank, ok := vecRankMap[id]; ok {
			rrfScore += 1.0 / (kConst + float64(rank))
		}

		results = append(results, SearchResultItem{
			ObjectID:      id,
			CombinedScore: rrfScore,
			KeywordScore:  keywordMatches[id],
			VectorScore:   vectorMatches[id],
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CombinedScore > results[j].CombinedScore
	})

	return results
}

type itemPair struct {
	ID    string
	Score float64
}

func sortByScore(m map[string]float64) []itemPair {
	pairs := make([]itemPair, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, itemPair{ID: k, Score: v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Score > pairs[j].Score
	})
	return pairs
}

// SimpleKeywordMatch computes basic frequency term overlap score.
func SimpleKeywordMatch(text, query string) float64 {
	words := strings.Fields(strings.ToLower(text))
	terms := strings.Fields(strings.ToLower(query))

	if len(words) == 0 || len(terms) == 0 {
		return 0.0
	}

	matches := 0
	for _, t := range terms {
		for _, w := range words {
			if strings.