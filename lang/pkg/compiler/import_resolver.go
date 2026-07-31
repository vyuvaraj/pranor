package compiler

import (
	"path/filepath"
	"strings"
	"sync"
)

// ImportedSymbol represents a schema type symbol imported from another file.
type ImportedSymbol struct {
	SymbolName string `json:"symbol_name"`
	SourceFile string `json:"source_file"`
	TypeKind   string `json:"type_kind"` // "struct", "enum", "type_alias"
}

// MultiFileImportResolver resolves cross-file `import` statements and builds a linked symbol table.
type MultiFileImportResolver struct {
	mu           sync.RWMutex
	symbolTable  map[string]ImportedSymbol // symbolName -> ImportedSymbol
	loadedFiles  map[string]bool
}

// NewMultiFileImportResolver creates a MultiFileImportResolver instance.
func NewMultiFileImportResolver() *MultiFileImportResolver {
	return &MultiFileImportResolver{
		symbolTable: make(map[string]ImportedSymbol),
		loadedFiles: make(map[string]bool),
	}
}

// ParseAndRegisterFile parses import directives and exports symbols into global symbol table.
func (mfir *MultiFileImportResolver) ParseAndRegisterFile(filePath string, fileContent string) ([]string, error) {
	mfir.mu.Lock()
	defer mfir.mu.Unlock()

	cleanPath := filepath.Clean(filePath)
	mfir.loadedFiles[cleanPath] = true

	var importedPaths []string
	lines := strings.Split(fileContent, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import ") {
			// Extract import path: e.g. import "./models.pnr"
			rawPath := strings.TrimPrefix(trimmed, "import ")
			rawPath = strings.Trim(rawPath, "\"' ;")
			importedPaths = append(importedPaths, rawPath)
		} else if strings.HasPrefix(trimmed, "struct ") || strings.HasPrefix(trimmed, "type ") {
			// Register exported symbol name
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				symName := strings.TrimSuffix(parts[1], "{")
				mfir.symbolTable[symName] = ImportedSymbol{
					SymbolName: symName,
					SourceFile: cleanPath,
					TypeKind:   parts[0],
				}
			}
		}
	}

	return importedPaths, nil
}

// ResolveSymbol checks if a symbol exists across imported multi-file modules.
func (mfir *MultiFileImportResolver) ResolveSymbol(symbolName string) (ImportedSymbol, bool) {
	mfir.mu.RLock()
	defer mfir.mu.RUnlock()
	sym, ok := mfir.symbolTable[symbolName]
	return sym, ok
}

// CheckUnresolvedSymbols validates if a list of reference types exist in symbol table.
func (mfir *MultiFileImportResolver) CheckUnresolvedSymbols(referencedSymbols []string) []string {
	mfir.mu.RLock()
	defer mfir.mu.RUnlock()

	var unresolved []string
	for _, sym := range referencedSymbols {
		if _, exists := mfir.symbolTable[sym]; !exists {
			unresolved = append(unresolved, sym)
		}
	}
	return unresolved
}
