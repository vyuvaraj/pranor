package lsp

import (
	"github.com/vyuvaraj/pranor/lang/compiler"
)

func analyzeSymbolsForText(text string) []symbolInfo {
	lexer := compiler.NewLexer(text)
	parser := compiler.NewParser(lexer)
	program := parser.ParseProgram()

	var symbols []symbolInfo
	var collectSymbols func(statements []compiler.Statement)
	collectSymbols = func(statements []compiler.Statement) {
		for _, stmt := range statements {
			if stmt == nil {
				continue
			}
			sym := extractSymbol(stmt)
			if sym.Name != "" {
				symbols = append(symbols, sym)
			}
			switch s := stmt.(type) {
			case *compiler.AppStmt:
				if s.Body != nil {
					collectSymbols(s.Body.Statements)
				}
			case *compiler.FnDecl:
				if s.Body != nil {
					collectSymbols(s.Body.Statements)
				}
			case *compiler.RouteStmt:
				if s.Body != nil {
					collectSymbols(s.Body.Statements)
				}
			}
		}
	}

	collectSymbols(program.Statements)
	return symbols
}
