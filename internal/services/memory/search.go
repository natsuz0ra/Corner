package memory

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

// openOrCreateBleveIndex opens or creates a bleve index.
func openOrCreateBleveIndex(indexPath string) (bleve.Index, error) {
	// Ensure parent directory of the index path exists.
	parentDir := filepath.Dir(indexPath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return nil, fmt.Errorf("create index parent dir: %w", err)
	}

	// Try to open an existing index.
	idx, err := bleve.Open(indexPath)
	if err == nil {
		return idx, nil
	}

	// No index yet; create a new one.
	mapping := buildBleveMapping()
	idx, err = bleve.New(indexPath, mapping)
	if err != nil {
		return nil, fmt.Errorf("create bleve index: %w", err)
	}
	return idx, nil
}

// buildBleveMapping builds the bleve index mapping.
func buildBleveMapping() *mapping.IndexMappingImpl {
	return bleve.NewIndexMapping()
}

// indexBleveDocument indexes one memory document.
func indexBleveDocument(idx bleve.Index, entry *MemoryEntry) error {
	doc := entry.ToBleveDocument()
	return idx.Index(entry.Slug(), doc)
}

// searchBleveIndex searches the bleve index and returns matching slugs.
// Uses DisjunctionQuery + MatchQuery across fields instead of QueryStringQuery,
// which handles CJK text poorly.
func searchBleveIndex(idx bleve.Index, queryString string, topK int) ([]string, error) {
	// MatchQuery on name, description, and content separately.
	nameMatch := bleve.NewMatchQuery(queryString)
	nameMatch.SetField("name")

	descMatch := bleve.NewMatchQuery(queryString)
	descMatch.SetField("description")

	contentMatch := bleve.NewMatchQuery(queryString)
	contentMatch.SetField("content")

	orQuery := bleve.NewDisjunctionQuery(nameMatch, descMatch, contentMatch)

	searchRequest := bleve.NewSearchRequest(orQuery)
	searchRequest.Size = topK
	searchRequest.Fields = []string{"name", "description", "type"}

	result, err := idx.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("bleve search: %w", err)
	}

	var slugs []string
	for _, hit := range result.Hits {
		slugs = append(slugs, hit.ID)
	}

	return slugs, nil
}

// searchBleveIndexBySession searches session-scoped memories; returns matching slugs.
// ConjunctionQuery combines full-text OR with an exact session_id filter.
func searchBleveIndexBySession(idx bleve.Index, sessionID, queryString string, topK int) ([]string, error) {
	// MatchQuery on name, description, and content separately.
	nameMatch := bleve.NewMatchQuery(queryString)
	nameMatch.SetField("name")

	descMatch := bleve.NewMatchQuery(queryString)
	descMatch.SetField("description")

	contentMatch := bleve.NewMatchQuery(queryString)
	contentMatch.SetField("content")

	// Exact session_id match.
	sessionTerm := bleve.NewTermQuery(sessionID)
	sessionTerm.SetField("session_id")

	// AND: (text OR across fields) AND session filter.
	textOrQuery := bleve.NewDisjunctionQuery(nameMatch, descMatch, contentMatch)
	andQuery := bleve.NewConjunctionQuery(textOrQuery, sessionTerm)

	searchRequest := bleve.NewSearchRequest(andQuery)
	searchRequest.Size = topK
	searchRequest.Fields = []string{"name", "description", "type"}

	result, err := idx.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("bleve search by session: %w", err)
	}

	var slugs []string
	for _, hit := range result.Hits {
		slugs = append(slugs, hit.ID)
	}

	return slugs, nil
}
