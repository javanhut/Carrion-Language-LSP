package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	glspServer "github.com/tliron/glsp/server"

	// Ensure a backend logger is available.
	_ "github.com/tliron/commonlog/simple"

	// Carrion language packages.
	"github.com/javanhut/TheCarrionLanguage/src/ast"
	"github.com/javanhut/TheCarrionLanguage/src/lexer"
	"github.com/javanhut/TheCarrionLanguage/src/parser"
)

// documents stores the current text for each document (keyed by URI).
var documents = make(map[string]string)

const lsName = "CarrionLang"

var version string = "0.0.1"

// myHandler is our custom handler type that embeds the base protocol.Handler.
type myHandler struct {
	protocol.Handler
}

// Handle implements the glsp.Handler interface with signature:
//
//	Handle(ctx *glsp.Context) (result any, handled bool, synchronous bool, err error)
//
// It reads the method from ctx.Method and uses our custom unmarshalParams function
// (which uses ctx.Params) to decode the incoming request parameters.
func (h *myHandler) Handle(ctx *glsp.Context) (any, bool, bool, error) {
	method := ctx.Method
	fmt.Printf("Received method: %s\n", method)
	switch method {
	case "textDocument/didOpen":
		var p protocol.DidOpenTextDocumentParams
		if err := unmarshalParams(ctx, &p); err != nil {
			return nil, true, false, err
		}
		err := didOpen(ctx, &p)
		return nil, true, false, err
	case "textDocument/didChange":
		var p protocol.DidChangeTextDocumentParams
		if err := unmarshalParams(ctx, &p); err != nil {
			return nil, true, false, err
		}
		err := didChange(ctx, &p)
		return nil, true, false, err
	case "textDocument/completion":
		var p protocol.CompletionParams
		if err := unmarshalParams(ctx, &p); err != nil {
			return nil, true, false, err
		}
		result, err := completion(ctx, &p)
		return result, true, false, err
	default:
		// Delegate any unhandled methods to the embedded base handler.
		return h.Handler.Handle(ctx)
	}
}

// unmarshalParams converts the JSON in ctx.Params into the target structure.
// It marshals ctx.Params to JSON and then unmarshals that JSON into target.
func unmarshalParams(ctx *glsp.Context, target interface{}) error {
	if ctx.Params == nil {
		return fmt.Errorf("no params found in context")
	}
	b, err := json.Marshal(ctx.Params)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

var myH *myHandler

func main() {
	// Create the base handler with core endpoints.
	baseHandler := protocol.Handler{
		Initialize:  initialize,
		Initialized: initialized,
		Shutdown:    shutdown,
		SetTrace:    setTrace,
	}
	// Create our custom handler that embeds the base handler.
	myH = &myHandler{
		Handler: baseHandler,
	}

	// Create the LSP server instance using our custom handler.
	lspServer := glspServer.NewServer(myH, lsName, false)

	// Run the LSP server over stdio.
	if err := lspServer.RunStdio(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running server: %s\n", err)
		os.Exit(1)
	}
}

//
// Core LSP Endpoints
//

func initialize(ctx *glsp.Context, params *protocol.InitializeParams) (any, error) {
	// Create default capabilities and add completion support.
	capabilities := myH.CreateServerCapabilities()
	resolveProvider := false
	capabilities.CompletionProvider = &protocol.CompletionOptions{
		ResolveProvider:   &resolveProvider,
		TriggerCharacters: []string{".", "("},
	}
	fmt.Println("Initialize called, sending capabilities with completion support.")
	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}, nil
}

func initialized(ctx *glsp.Context, params *protocol.InitializedParams) error {
	fmt.Println("Server initialized.")
	return nil
}

func shutdown(ctx *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(ctx *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}

//
// Custom Document Handlers & Auto-Completion
//

// didOpen stores the text of a document when it is opened.
func didOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	fmt.Printf("didOpen: %s\n", params.TextDocument.URI)
	documents[params.TextDocument.URI] = params.TextDocument.Text
	return nil
}

// didChange updates the document text when changes occur.
func didChange(ctx *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	fmt.Printf("didChange: %s\n", params.TextDocument.URI)
	if len(params.ContentChanges) > 0 {
		var change protocol.TextDocumentContentChangeEvent
		b, err := json.Marshal(params.ContentChanges[0])
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &change); err != nil {
			return err
		}
		documents[params.TextDocument.URI] = change.Text
	}
	return nil
}

// completion re-parses the document, extracts symbols from the AST,
// and returns them as auto-complete suggestions.
func completion(ctx *glsp.Context, params *protocol.CompletionParams) (any, error) {
	fmt.Printf("completion requested for: %s\n", params.TextDocument.URI)
	uri := params.TextDocument.URI
	doc, ok := documents[uri]
	if !ok {
		fmt.Printf("No document found for URI: %s\n", uri)
		return &protocol.CompletionList{
			IsIncomplete: false,
			Items:        []protocol.CompletionItem{},
		}, nil
	}

	// Parse the document using the Carrion language lexer and parser.
	lex := lexer.New(doc)
	p := parser.New(lex)
	program := p.ParseProgram()

	// If there are parser errors, return an empty completion list.
	if len(p.Errors()) > 0 {
		fmt.Printf("Parser errors: %v\n", p.Errors())
		return &protocol.CompletionList{
			IsIncomplete: false,
			Items:        []protocol.CompletionItem{},
		}, nil
	}

	// Extract symbols from the AST.
	symbols := extractSymbols(program)
	fmt.Printf("Extracted symbols: %v\n", symbols)

	// Build completion items.
	var items []protocol.CompletionItem
	for _, sym := range symbols {
		kind := protocol.CompletionItemKindVariable
		items = append(items, protocol.CompletionItem{
			Label: sym,
			Kind:  &kind,
		})
	}
	return &protocol.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

// extractSymbols traverses the AST and collects top-level symbols.
// This basic implementation collects function names and assignment targets.
func extractSymbols(program *ast.Program) []string {
	symbolSet := make(map[string]struct{})
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.FunctionDefinition:
			if s.Name != nil {
				symbolSet[s.Name.Value] = struct{}{}
			}
		case *ast.AssignStatement:
			if ident, ok := s.Name.(*ast.Identifier); ok {
				symbolSet[ident.Value] = struct{}{}
			}
		}
	}
	var symbols []string
	for sym := range symbolSet {
		symbols = append(symbols, sym)
	}
	return symbols
}

