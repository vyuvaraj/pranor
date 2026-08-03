package lsp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// WorkspaceSymbolInformation represents a symbol found workspace-wide (LSP.3)
type WorkspaceSymbolInformation struct {
	Name          string   `json:"name"`
	Kind          int      `json:"kind"`
	Location      Location `json:"location"`
	ContainerName string   `json:"containerName,omitempty"`
}

// CallHierarchyItem represents a function/route node in call hierarchy (LSP.4)
type CallHierarchyItem struct {
	Name           string   `json:"name"`
	Kind           int      `json:"kind"`
	Detail         string   `json:"detail,omitempty"`
	URI            string   `json:"uri"`
	Range          Range    `json:"range"`
	SelectionRange Range    `json:"selectionRange"`
}

type CallHierarchyIncomingCall struct {
	From       CallHierarchyItem `json:"from"`
	FromRanges []Range           `json:"fromRanges"`
}

type CallHierarchyOutgoingCall struct {
	To         CallHierarchyItem `json:"to"`
	FromRanges []Range           `json:"fromRanges"`
}

type DocumentHighlight struct {
	Range Range `json:"range"`
	Kind  int   `json:"kind,omitempty"` // 1=Text, 2=Read, 3=Write
}

// handleWorkspaceSymbol handles workspace/symbol search across all workspace .pnr files (LSP.3)
func (s *Server) handleWorkspaceSymbol(msg JSONRPCMessage) {
	var params struct {
		Query string `json:"query"`
	}
	json.Unmarshal(msg.Params, &params)

	queryLower := strings.ToLower(params.Query)
	results := []WorkspaceSymbolInformation{}

	s.mu.RLock()
	// 1. Search in-memory symbols
	for uri, syms := range s.symbols {
		for _, sym := range syms {
			if queryLower == "" || strings.Contains(strings.ToLower(sym.Name), queryLower) {
				results = append(results, WorkspaceSymbolInformation{
					Name: sym.Name,
					Kind: getSymbolKindInt(sym.Kind),
					Location: Location{
						URI: uri,
						Range: Range{
							Start: Position{Line: sym.Line, Character: sym.Col},
							End:   Position{Line: sym.Line, Character: sym.Col + len(sym.Name)},
						},
					},
				})
			}
		}
	}
	s.mu.RUnlock()

	// 2. Scan workspace directory for unopened .pnr files if any open file path available
	s.mu.RLock()
	var sampleURI string
	for uri := range s.documents {
		sampleURI = uri
		break
	}
	s.mu.RUnlock()

	if sampleURI != "" {
		currentPath := strings.TrimPrefix(sampleURI, "file://")
		workspaceDir := filepath.Dir(currentPath)
		if workspaceDir != "" && workspaceDir != "." {
			_ = filepath.WalkDir(workspaceDir, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(path, ".pnr") {
					return nil
				}
				fileURI := "file://" + filepath.ToSlash(path)
				s.mu.RLock()
				_, open := s.documents[fileURI]
				s.mu.RUnlock()
				if open {
					return nil
				}

				data, readErr := os.ReadFile(path)
				if readErr != nil {
					return nil
				}
				fileSyms := analyzeSymbolsForText(string(data))
				for _, sym := range fileSyms {
					if queryLower == "" || strings.Contains(strings.ToLower(sym.Name), queryLower) {
						results = append(results, WorkspaceSymbolInformation{
							Name: sym.Name,
							Kind: getSymbolKindInt(sym.Kind),
							Location: Location{
								URI: fileURI,
								Range: Range{
									Start: Position{Line: sym.Line, Character: sym.Col},
									End:   Position{Line: sym.Line, Character: sym.Col + len(sym.Name)},
								},
							},
						})
					}
				}
				return nil
			})
		}
	}

	sendResponse(msg.ID, results)
}

// handlePrepareCallHierarchy prepares a call hierarchy item at position (LSP.4)
func (s *Server) handlePrepareCallHierarchy(msg JSONRPCMessage) {
	var params struct {
		TextDocument TextDocumentIdentifier `json:"textDocument"`
		Position     Position               `json:"position"`
	}
	json.Unmarshal(msg.Params, &params)

	s.mu.RLock()
	text := s.documents[params.TextDocument.URI]
	syms := s.symbols[params.TextDocument.URI]
	s.mu.RUnlock()

	word := getWordAtPosition(text, params.Position)
	if word == "" {
		sendResponse(msg.ID, nil)
		return
	}

	for _, sym := range syms {
		if (sym.Kind == "fn" || sym.Kind == "route" || sym.Kind == "method") && sym.Name == word {
			item := CallHierarchyItem{
				Name:   sym.Name,
				Kind:   getSymbolKindInt(sym.Kind),
				Detail: sym.TypeInfo,
				URI:    params.TextDocument.URI,
				Range: Range{
					Start: Position{Line: sym.Line, Character: sym.Col},
					End:   Position{Line: sym.Line, Character: sym.Col + len(sym.Name)},
				},
				SelectionRange: Range{
					Start: Position{Line: sym.Line, Character: sym.Col},
					End:   Position{Line: sym.Line, Character: sym.Col + len(sym.Name)},
				},
			}
			sendResponse(msg.ID, []CallHierarchyItem{item})
			return
		}
	}

	sendResponse(msg.ID, nil)
}

// handleCallHierarchyIncomingCalls returns callers of the item (LSP.4)
func (s *Server) handleCallHierarchyIncomingCalls(msg JSONRPCMessage) {
	var params struct {
		Item CallHierarchyItem `json:"item"`
	}
	json.Unmarshal(msg.Params, &params)

	var incoming []CallHierarchyIncomingCall

	s.mu.RLock()
	defer s.mu.RUnlock()

	for uri, docText := range s.documents {
		lines := strings.Split(docText, "\n")
		for lineNum, line := range lines {
			if strings.Contains(line, params.Item.Name+"(") && !strings.HasPrefix(strings.TrimSpace(line), "fn ") {
				callerSym := findEnclosingSymbol(s.symbols[uri], lineNum)
				callerName := "global"
				callerKind := 12
				if callerSym != nil {
					callerName = callerSym.Name
					callerKind = getSymbolKindInt(callerSym.Kind)
				}

				incoming = append(incoming, CallHierarchyIncomingCall{
					From: CallHierarchyItem{
						Name:  callerName,
						Kind:  callerKind,
						URI:   uri,
						Range: Range{Start: Position{Line: lineNum, Character: 0}, End: Position{Line: lineNum, Character: len(line)}},
					},
					FromRanges: []Range{
						{Start: Position{Line: lineNum, Character: strings.Index(line, params.Item.Name)}, End: Position{Line: lineNum, Character: strings.Index(line, params.Item.Name) + len(params.Item.Name)}},
					},
				})
			}
		}
	}

	sendResponse(msg.ID, incoming)
}

// handleCallHierarchyOutgoingCalls returns calls made by the item (LSP.4)
func (s *Server) handleCallHierarchyOutgoingCalls(msg JSONRPCMessage) {
	var params struct {
		Item CallHierarchyItem `json:"item"`
	}
	json.Unmarshal(msg.Params, &params)

	var outgoing []CallHierarchyOutgoingCall

	s.mu.RLock()
	docText := s.documents[params.Item.URI]
	syms := s.symbols[params.Item.URI]
	s.mu.RUnlock()

	lines := strings.Split(docText, "\n")
	startLine := params.Item.Range.Start.Line
	endLine := len(lines) - 1

	for _, sym := range syms {
		if sym.Line > startLine && sym.Line < endLine {
			endLine = sym.Line - 1
			break
		}
	}

	for i := startLine; i <= endLine && i < len(lines); i++ {
		line := lines[i]
		for _, sym := range syms {
			if sym.Name != params.Item.Name && strings.Contains(line, sym.Name+"(") {
				outgoing = append(outgoing, CallHierarchyOutgoingCall{
					To: CallHierarchyItem{
						Name:  sym.Name,
						Kind:  getSymbolKindInt(sym.Kind),
						URI:   params.Item.URI,
						Range: Range{Start: Position{Line: sym.Line, Character: sym.Col}, End: Position{Line: sym.Line, Character: sym.Col + len(sym.Name)}},
					},
					FromRanges: []Range{
						{Start: Position{Line: i, Character: strings.Index(line, sym.Name)}, End: Position{Line: i, Character: strings.Index(line, sym.Name) + len(sym.Name)}},
					},
				})
			}
		}
	}

	sendResponse(msg.ID, outgoing)
}

// handleDocumentHighlight highlights occurrences of symbol under cursor in document (LSP.6)
func (s *Server) handleDocumentHighlight(msg JSONRPCMessage) {
	var params struct {
		TextDocument TextDocumentIdentifier `json:"textDocument"`
		Position     Position               `json:"position"`
	}
	json.Unmarshal(msg.Params, &params)

	s.mu.RLock()
	text := s.documents[params.TextDocument.URI]
	s.mu.RUnlock()

	word := getWordAtPosition(text, params.Position)
	if word == "" || isBuiltinOrKeyword(word) {
		sendResponse(msg.ID, []DocumentHighlight{})
		return
	}

	var highlights []DocumentHighlight
	lines := strings.Split(text, "\n")

	for lineNum, line := range lines {
		locs := findWordOccurrencesInLine(line, word, lineNum, params.TextDocument.URI)
		for _, loc := range locs {
			highlights = append(highlights, DocumentHighlight{
				Range: loc.Range,
				Kind:  1,
			})
		}
	}

	sendResponse(msg.ID, highlights)
}

func getSymbolKindInt(kind string) int {
	switch kind {
	case "fn":
		return 12
	case "struct":
		return 23
	case "method":
		return 6
	case "interface":
		return 11
	case "route":
		return 12
	case "middleware":
		return 12
	case "enum":
		return 10
	default:
		return 13
	}
}
