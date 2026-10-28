package Lyger

import (
	"strings"

	"github.com/l-programming-language/l/lerrors"
)

// nativeTypeDefinition binds a source-selected native declaration as a type atom.
// It never guesses a declaration from a Rust type spelling or creates an L value
// kind. The complete selected HY remains an ordinary declaration dependency.
func (c *Compilation) nativeTypeDefinition(tok *token) (string, *DefinitionInstance, error) {
	if tok == nil || tok.binding == nil {
		return "", nil, nil
	}
	return c.nativeDeclaration(tok.binding.Record)
}

func (c *Compilation) nativeDeclaration(rec *record) (string, *DefinitionInstance, error) {
	rec, err := canonicalRecord(rec, c.Semantics.Aliases)
	if err != nil {
		return "", nil, err
	}
	if c.Semantics.Declarations[rec] == nil || rec.function != nil {
		return "", nil, nil
	}
	if c.nativeDeclarations[rec] {
		return "", nil, lerrors.Cycle.Errorf("recursive native declaration: %s", rec.ID)
	}
	if c.nativeDeclarations == nil {
		c.nativeDeclarations = make(map[*record]bool)
	}
	c.nativeDeclarations[rec] = true
	defer delete(c.nativeDeclarations, rec)
	template, err := c.recordTemplate(rec)
	if err != nil {
		return "", nil, err
	}
	var declaration DefinitionInstance
	if template != nil {
		if len(template.Slots) != 0 {
			return "", nil, nil
		}
		declaration, err = template.Instantiate(nil)
	} else if object, objectErr := c.definitionObject(rec); objectErr != nil {
		return "", nil, objectErr
	} else if object != nil {
		node, err := c.constructObject(object, make(map[*descriptorRule]bool))
		if err != nil {
			return "", nil, err
		}
		nominal, err := c.recordTemplate(node.Definition)
		if err != nil {
			return "", nil, err
		}
		if nominal != nil && (len(nominal.parts) == 0 || !strings.HasPrefix(strings.TrimSpace(nominal.parts[0].text), "HY{|")) {
			return "", nil, nil
		}
		declaration, err = c.instantiateCall(rec, nil, make(map[*record]bool))
		if err != nil {
			return "", nil, err
		}
	} else {
		return "", nil, nil
	}
	if err != nil {
		return "", nil, err
	}
	name, err := nativeDeclarationName(strings.TrimSpace(declaration.Expression))
	if err != nil || name == "" {
		return "", nil, err
	}
	declaration.Symbol, declaration.definition = rec.ID, rec
	return name, &declaration, nil
}

// A selected member retains native declarations belonging to its canonical
// source owner chain. Sibling declarations and opaque native type names do not
// participate; their dependencies must be selected through actual source words.
func (c *Compilation) declarationDependencies(rec *record) ([]DefinitionInstance, error) {
	var dependencies []DefinitionInstance
	for owner := rec.declarationOwner; owner != nil; owner = owner.declarationOwner {
		_, declaration, err := c.nativeDeclaration(owner)
		if err != nil {
			return nil, err
		}
		if declaration != nil {
			dependencies = append(dependencies, *declaration)
		}
	}
	return dependencies, nil
}

func nativeDeclarationName(expression string) (string, error) {
	if !strings.HasPrefix(expression, "HY{|") {
		return "", nil
	}
	end, err := definitionMetadataEnd(expression, 2)
	if err != nil {
		return "", err
	}
	if end+2 != len(expression) {
		return "", lerrors.InvalidInput.Errorf("native type binding requires one declaration")
	}
	metadata := expression[4:end]
	name := ""
	for metadata != "" {
		key, value, found := strings.Cut(metadata, ":")
		value = strings.TrimSpace(value)
		if !found || !strings.HasPrefix(value, "~") {
			return "", lerrors.InvalidInput.Errorf("native declaration metadata requires authored String atoms")
		}
		close := strings.IndexByte(value[1:], '~')
		if close < 0 {
			return "", lerrors.Syntax.Errorf("unterminated native declaration metadata")
		}
		if strings.TrimSpace(key) == "Key" {
			if name != "" {
				return "", lerrors.InvalidInput.Errorf("native declaration has duplicate Key")
			}
			name = value[1 : close+1]
		}
		metadata = strings.TrimSpace(value[close+2:])
		if metadata != "" {
			if metadata[0] != ',' {
				return "", lerrors.Syntax.Errorf("native declaration metadata has no field separator")
			}
			metadata = strings.TrimSpace(metadata[1:])
		}
	}
	if !validDefinitionNativeName(name) || name == "_" {
		return "", lerrors.InvalidInput.Errorf("native declaration requires a valid Key atom")
	}
	return name, nil
}
