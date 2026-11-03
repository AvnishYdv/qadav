package checker

import (
	"fmt"

	"github.com/qadav-lang/qadav/internal/ast"
	"github.com/qadav-lang/qadav/internal/types"
)

// Symbol represents an identifier definition in a scope.
type Symbol struct {
	Name    string
	Type    *types.Type
	Pos     ast.Position
	IsConst bool
	IsParam bool
}

// Scope represents a lexical environment holding symbols and linking to an outer scope.
type Scope struct {
	parent  *Scope
	symbols map[string]*Symbol
}

// NewScope creates a new Scope with an optional parent.
func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent:  parent,
		symbols: make(map[string]*Symbol),
	}
}

// Define inserts a new symbol into the current scope.
// It returns an error if the symbol was already declared in this exact scope.
func (s *Scope) Define(sym *Symbol) error {
	if sym == nil {
		return fmt.Errorf("cannot define nil symbol")
	}
	if existing, ok := s.symbols[sym.Name]; ok {
		return fmt.Errorf("redeclaration of identifier %q (previously declared at %s)", sym.Name, existing.Pos)
	}
	s.symbols[sym.Name] = sym
	return nil
}

// Lookup resolves a symbol by name in the current scope or any ancestor.
func (s *Scope) Lookup(name string) (*Symbol, bool) {
	curr := s
	for curr != nil {
		if sym, ok := curr.symbols[name]; ok {
			return sym, true
		}
		curr = curr.parent
	}
	return nil, false
}

// LookupCurrent resolves a symbol by name in the current scope only.
func (s *Scope) LookupCurrent(name string) (*Symbol, bool) {
	sym, ok := s.symbols[name]
	return sym, ok
}

// Parent returns the enclosing scope.
func (s *Scope) Parent() *Scope {
	return s.parent
}
