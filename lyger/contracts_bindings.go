package Lyger

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/l-programming-language/l/contracts"
)

// ContractBindings is a snapshot of the existing resolver, not a replacement
// resolver. The reference selector belongs to contracts and is never called here.
type ContractBindings struct {
	Policy      string                      `json:"policy"`
	Bindings    []contracts.ResolvedBinding `json:"bindings"`
	Targets     []contracts.TargetRef       `json:"targets"`
	Definitions []ContractDefinition        `json:"definitions"`
	Inputs      map[string]string           `json:"inputs"`
}

// ContractDefinition exposes declared meanings even when no concrete adjective
// application exists. It is discovery metadata, not a symbolic certificate.
type ContractDefinition = contracts.DefinitionRecord

type contractDescriptorUse struct {
	Token  *token
	Rule   *descriptorRule
	Target *constructionNode
}

type contractRecorder struct {
	compilation              *Compilation
	origins                  map[*token]contracts.Origin
	statements               map[*token]*lStatement
	lookupReasons            map[*token]string
	descriptors              []contractDescriptorUse
	sourceConstructions      []contractSourceConstruction
	definitions              map[*record]contracts.DefinitionKey
	declarationTokens        map[*record]*token
	occurrences              map[sourceOccurrenceID]contracts.BindingOccurrenceID
	targets                  map[*constructionNode]contracts.TargetRef
	appTargets               map[*App]contracts.TargetRef
	constructionTargets      []contracts.TargetRef
	environments             map[*lexicalScope]contracts.BindingEnvironmentKey
	configurationFingerprint string
	lowering                 []contracts.LoweringStep
	nativeAnalysis           *contractNativeAnalysis
	nativeEmissions          map[*instruction][]contractNativeEmission
	identityReady            bool
}

func newContractRecorder(c *Compilation) *contractRecorder {
	return &contractRecorder{compilation: c, origins: make(map[*token]contracts.Origin), statements: make(map[*token]*lStatement), lookupReasons: make(map[*token]string), definitions: make(map[*record]contracts.DefinitionKey), declarationTokens: make(map[*record]*token), occurrences: make(map[sourceOccurrenceID]contracts.BindingOccurrenceID), targets: make(map[*constructionNode]contracts.TargetRef), appTargets: make(map[*App]contracts.TargetRef), environments: make(map[*lexicalScope]contracts.BindingEnvironmentKey)}
}

func (r *contractRecorder) logicalSource(path string) string {
	if strings.HasPrefix(path, "resource:") {
		return path
	}
	for _, origin := range r.compilation.SourceOrigins {
		if origin.Name == path && strings.HasPrefix(origin.Identity, "resource:") {
			return origin.Identity
		}
	}
	if relative, err := filepath.Rel(r.compilation.RootDir, path); err == nil {
		return "source:" + filepath.ToSlash(relative)
	}
	return "source:" + filepath.ToSlash(path)
}

func (r *contractRecorder) occurrence(id sourceOccurrenceID) contracts.BindingOccurrenceID {
	if key := r.occurrences[id]; key != "" {
		return key
	}
	for _, occurrence := range r.compilation.SourceOccurrences {
		if occurrence.ID != id {
			continue
		}
		var parent contracts.BindingOccurrenceID
		if occurrence.Parent != 0 {
			parent = r.occurrence(occurrence.Parent)
		}
		key := contracts.BindingOccurrenceID(contracts.StableID("source-use", struct {
			Source       string
			Parent       contracts.BindingOccurrenceID
			ImportOffset int
		}{r.logicalSource(occurrence.Path), parent, occurrence.ImportSite.Start.Offset}))
		r.occurrences[id] = key
		return key
	}
	return ""
}

// sourceTokenOffsets maps original lexical words after comment removal. It
// retains quoted bytes as one token just as PrepareSource does before caching.
// A generated alias is explicitly attributed to its import rather than given
// fabricated token coordinates.
func sourceTokenOffsets(statement *SourceStatement) [][2]int {
	parts, err := sourceParts(statement.Raw)
	if err != nil || len(parts) != 1 {
		return nil
	}
	part := parts[0]
	end := part.end
	if part.terminator >= 0 {
		end = part.terminator
	}
	base := statement.Span.Start.Offset - part.codeStart
	var clean []byte
	var offsets []int
	for i := 0; i < end; {
		if bytes.HasPrefix(statement.Raw[i:], []byte("***")) {
			next, err := commentEnd(statement.Raw, i)
			if err != nil {
				return nil
			}
			clean = append(clean, ' ')
			offsets = append(offsets, base+i)
			i = next
			continue
		}
		next := i + 1
		if statement.Raw[i] == '`' || statement.Raw[i] == '"' {
			next, err = quotedEnd(statement.Raw, i)
			if err != nil {
				return nil
			}
		}
		for ; i < next; i++ {
			clean = append(clean, statement.Raw[i])
			offsets = append(offsets, base+i)
		}
	}
	var result [][2]int
	for i := 0; i < len(clean); {
		for i < len(clean) && sourceSpace(clean[i]) {
			i++
		}
		start := i
		for i < len(clean) && !sourceSpace(clean[i]) {
			if clean[i] == '`' || clean[i] == '"' {
				i, err = quotedEnd(clean, i)
				if err != nil {
					return nil
				}
			} else {
				i++
			}
		}
		if start < i {
			result = append(result, [2]int{offsets[start], offsets[i-1] + 1})
		}
	}
	return result
}

func (r *contractRecorder) captureStatement(statement *lStatement) {
	if statement.source == nil {
		return
	}
	source := statement.source
	spans := sourceTokenOffsets(source)
	generated := source.imported != nil || len(spans) != len(statement.tokens)
	for index, tok := range statement.tokens {
		origin := contracts.Origin{Source: r.logicalSource(source.Span.Path), Start: source.Span.Start.Offset, End: source.Span.End.Offset, Parent: r.occurrence(source.occurrence)}
		if !generated {
			origin.Start, origin.End = spans[index][0], spans[index][1]
		} else {
			origin.Label = "generated import binding"
			origin.Expansion = []int{index}
		}
		r.origins[tok], r.statements[tok] = origin, statement
	}
}

func (r *contractRecorder) rebaseConstruction(from, to *constructionNode) {
	for i := range r.descriptors {
		if r.descriptors[i].Target == from {
			r.descriptors[i].Target = to
		}
	}
	for i := range r.sourceConstructions {
		if r.sourceConstructions[i].Target == from {
			// The source operation happened on this fresh contribution before
			// its merge. Its local snapshots are not snapshots of the receiver.
			r.sourceConstructions[i].Detached = true
		}
	}
}

func (r *contractRecorder) origin(tok *token) contracts.Origin {
	if origin, ok := r.origins[tok]; ok {
		return origin
	}
	return contracts.Origin{Label: "generated semantic use; original token unavailable"}
}

func (r *contractRecorder) scopeID(scope *lexicalScope) string {
	if scope == nil {
		return "unresolved-scope"
	}
	var chain []string
	for current := scope; current != nil; current = current.parent {
		unit := ""
		if current.unit != nil {
			unit = r.logicalSource(current.unit.Origin.Name)
		}
		chain = append(chain, fmt.Sprintf("%s:%d:%d", unit, current.kind, current.source.Start.Offset))
	}
	return contracts.StableID("scope", chain)
}

func (r *contractRecorder) definition(rec *record) contracts.DefinitionKey {
	r.prepareIdentities()
	canonical, err := canonicalRecord(rec, r.compilation.Semantics.Aliases)
	if err != nil || canonical == nil {
		return ""
	}
	rec = canonical
	if key := r.definitions[rec]; key != "" {
		return key
	}
	tok := r.declarationTokens[rec]
	origin := r.definitionOrigin(rec)
	scope := rec.storageScope
	if tok != nil {
		scope = tok.scope
	}
	var owners []string
	for owner := rec.declarationOwner; owner != nil; owner = owner.declarationOwner {
		owners = append(owners, owner.ID)
	}
	identity := contracts.DefinitionIdentity{Source: origin.Source, Scope: r.scopeID(scope), Owner: owners, Declaration: origin, Role: rec.Type}
	if origin.Source == "" {
		identity.Source = "configured-or-generated:" + rec.ID
		identity.Role = rec.Type
	}
	key := contracts.DefinitionKey(contracts.StableID("definition", identity))
	r.definitions[rec] = key
	return key
}

func (r *contractRecorder) definitionOrigin(rec *record) contracts.Origin {
	if tok := r.declarationTokens[rec]; tok != nil {
		return r.origin(tok)
	}
	if statement := r.compilation.Semantics.Declarations[rec]; statement != nil && statement.source != nil {
		source := statement.source
		return contracts.Origin{Source: r.logicalSource(source.Span.Path), Start: source.Span.Start.Offset, End: source.Span.End.Offset, Parent: r.occurrence(source.occurrence), Label: "generated member of declaration " + rec.ID}
	}
	return contracts.Origin{Label: "configured or generated declaration " + rec.ID}
}

func (r *contractRecorder) configurationIdentity() string {
	if r.configurationFingerprint == "" {
		r.configurationFingerprint = contracts.StableID("language-configuration", struct {
			Language   string
			Vocabulary *lexicon
			Grammar    map[string]grammeme
		}{r.compilation.Gram.language, r.compilation.Gram.lex, r.compilation.Gram.grammarMap})
	}
	return r.configurationFingerprint
}

func (r *contractRecorder) meaning(rec *record, rule *descriptorRule) contracts.MeaningKey {
	identity := contracts.MeaningIdentity{Definition: r.definition(rec), Version: contracts.TrackAPolicy}
	if rule != nil {
		if rule.Source != nil && rule.Source.source != nil {
			s := rule.Source.source
			identity.Rule = contracts.Origin{Source: r.logicalSource(s.Span.Path), Start: s.Span.Start.Offset, End: s.Span.End.Offset, Parent: r.occurrence(s.occurrence)}
		}
		if rule.Target.definition != nil {
			identity.Applicability = string(r.definition(rule.Target.definition))
		} else {
			identity.Applicability = "native:" + rule.Target.String()
		}
		if descriptor := r.compilation.Semantics.Descriptors[rec]; descriptor != nil {
			for i, candidate := range descriptor.Rules {
				if candidate == rule {
					identity.Rule.Expansion = []int{i}
					break
				}
			}
		}
	}
	return contracts.MeaningKey(contracts.StableID("meaning", identity))
}

func (r *contractRecorder) environment(scope *lexicalScope) contracts.BindingEnvironmentKey {
	if key := r.environments[scope]; key != "" {
		return key
	}
	// Whole visible environments intentionally include candidates not selected,
	// local reservations, and source/resource fingerprints. New imports therefore
	// cannot reuse a selected-word-only cache entry.
	var entries []string
	for current := scope; current != nil; current = current.parent {
		for name, rec := range current.records {
			entries = append(entries, r.scopeID(current)+":"+name+":"+string(r.definition(rec)))
		}
		for _, decl := range current.declarations {
			entries = append(entries, "reservation:"+r.scopeID(current)+":"+decl.SubjectID)
		}
	}
	for _, source := range r.compilation.SourceOrigins {
		entries = append(entries, r.logicalSource(source.Name)+":"+source.SHA256)
	}
	entries = append(entries, r.configurationIdentity())
	sort.Strings(entries)
	key := contracts.BindingEnvironmentKey(contracts.StableID("environment", struct {
		Policy  string
		Entries []string
	}{contracts.TrackAPolicy, entries}))
	r.environments[scope] = key
	return key
}

func (r *contractRecorder) prepareIdentities() {
	if !r.identityReady {
		r.indexDeclarations()
	}
}

func (r *contractRecorder) indexDeclarations() {
	for _, statement := range r.compilation.Statements {
		for _, tok := range statement.tokens {
			if rec := r.compilation.Semantics.Subjects[tok]; rec != nil && r.compilation.Semantics.Aliases[rec] == nil {
				if r.declarationTokens[rec] == nil {
					r.declarationTokens[rec] = tok
				}
			}
		}
	}
	r.identityReady = true
}

func (r *contractRecorder) binding(tok *token, rec *record, rule *descriptorRule, target contracts.TargetRef, ordinal int) contracts.ResolvedBinding {
	canonical, canonicalErr := canonicalRecord(rec, r.compilation.Semantics.Aliases)
	if canonicalErr != nil || canonical == nil {
		canonical = rec
		target.Unknown = true
	}
	origin := r.origin(tok)
	if rule != nil {
		origin.Expansion = append(append([]int(nil), origin.Expansion...), ordinal)
	}
	definitionToken := r.declarationTokens[canonical]
	definitionScope := canonical.storageScope
	if definitionToken != nil {
		definitionScope = definitionToken.scope
	}
	reason := r.lookupReasons[tok]
	if reason == "" {
		reason = "existing resolved declaration/member or definition-site binding"
	}
	if rule != nil {
		reason = "existing ordered descriptor composition: every matching canonical target rule applies"
	}
	if canonicalErr != nil {
		reason = "canonical binding unresolved: " + canonicalErr.Error()
	}
	environment := r.environment(tok.scope)
	return contracts.ResolvedBinding{Occurrence: contracts.BindingOccurrenceID(contracts.StableID("binding-use", origin)), Definition: r.definition(canonical), Meaning: r.meaning(canonical, rule), Spelling: string(tok.rawBytes), CanonicalName: canonical.ID, LexicalRole: tok.tType, Target: target, Origin: origin, DefinitionOrigin: r.definitionOrigin(canonical), DefinitionEnvironment: r.environment(definitionScope), UseEnvironment: environment, Policy: contracts.TrackAPolicy, SelectionReason: reason, CandidateFingerprint: string(environment), Dependencies: []string{string(environment)}}
}

// Construction roots and definition-site identities must already exist when
// workspace emission and lifecycle projection first ask for an App target.
func (r *contractRecorder) indexResolvedTargets() {
	c := r.compilation
	r.indexDeclarations()
	var visit func(*constructionNode, contracts.TargetRef)
	visit = func(node *constructionNode, target contracts.TargetRef) {
		if node == nil {
			return
		}
		target.Type = node.Nominal
		r.targets[node] = target
		r.constructionTargets = append(r.constructionTargets, target)
		for index, child := range node.Children {
			next := target
			next.Path = append(append([]string(nil), target.Path...), fmt.Sprintf("%s[%d]", child.Nominal, index))
			visit(child, next)
		}
	}
	for _, construction := range c.Semantics.Constructions {
		root := contracts.StableID("construction", struct {
			Origin contracts.Origin
			Name   string
		}{r.origin(construction.Root.Source.noun), construction.App.ID})
		visit(construction.Root, contracts.TargetRef{Root: root, Phase: contracts.Construction})
		r.appTargets[construction.App] = r.targets[construction.Root]
	}
}

func (r *contractRecorder) snapshot() ContractBindings {
	c := r.compilation
	result := ContractBindings{Policy: contracts.TrackAPolicy, Inputs: make(map[string]string)}
	for _, source := range c.SourceOrigins {
		result.Inputs[r.logicalSource(source.Name)] = source.SHA256
	}
	result.Inputs["configuration:"+c.Gram.language] = r.configurationIdentity()
	for _, target := range r.constructionTargets {
		target.Path = append([]string(nil), target.Path...)
		result.Targets = append(result.Targets, target)
	}
	for _, statement := range c.Statements {
		for _, tok := range statement.tokens {
			rec := c.Semantics.Subjects[tok]
			if tok.binding != nil {
				rec = tok.binding.Record
			}
			if rec == nil {
				continue
			}
			canonical, err := canonicalRecord(rec, c.Semantics.Aliases)
			if err != nil || canonical == nil {
				continue
			}
			if c.Semantics.Descriptors[canonical] != nil {
				continue
			}
			target := r.recordTarget(canonical, contracts.Elaboration)
			result.Bindings = append(result.Bindings, r.binding(tok, canonical, nil, target, 0))
		}
	}
	for ordinal, use := range r.descriptors {
		rec := use.Token.binding.Record
		canonical, err := canonicalRecord(rec, c.Semantics.Aliases)
		if err != nil || canonical == nil {
			continue
		}
		target, ok := r.targets[use.Target]
		if !ok {
			target = contracts.TargetRef{Unknown: true, Phase: contracts.Construction}
		}
		result.Bindings = append(result.Bindings, r.binding(use.Token, canonical, use.Rule, target, ordinal))
	}
	result.Definitions = r.definitionCatalog()
	result.Bindings = append(result.Bindings, r.instructionBindings()...)
	return result
}

func (r *contractRecorder) recordTarget(rec *record, phase contracts.Phase) contracts.TargetRef {
	canonical, err := canonicalRecord(rec, r.compilation.Semantics.Aliases)
	if err != nil || canonical == nil {
		return contracts.TargetRef{Unknown: true, Phase: phase}
	}
	if value := canonical.Value; value != nil && value.Applicative != nil && value.Applicative.App != nil {
		if target, ok := r.appTargets[value.Applicative.App]; ok {
			target.Phase = phase
			return target
		}
	}
	return contracts.TargetRef{Root: string(r.definition(canonical)), Type: canonical.Type, Phase: phase}
}

func (r *contractRecorder) valueTarget(value *instructionValue) contracts.TargetRef {
	if value == nil {
		return contracts.TargetRef{Unknown: true, Phase: contracts.Invocation}
	}
	if value.Owner != nil {
		target := r.valueTarget(value.Owner)
		target.Path = append(append([]string(nil), target.Path...), value.Member.String())
		target.Type = value.Type.String()
		return target
	}
	if value.Record != nil {
		return r.recordTarget(value.Record, contracts.Invocation)
	}
	if value.Token != nil {
		return contracts.TargetRef{Root: contracts.StableID("value-use", r.origin(value.Token)), Type: value.Type.String(), Phase: contracts.Invocation}
	}
	return contracts.TargetRef{Unknown: true, Type: value.Type.String(), Phase: contracts.Invocation}
}

func (r *contractRecorder) nativeBinding(tok *token, role nativeRole, target contracts.TargetRef, origin contracts.Origin, environment contracts.BindingEnvironmentKey) contracts.ResolvedBinding {
	definition := contracts.DefinitionKey(contracts.StableID("native-definition", struct{ Language, Role string }{r.compilation.Gram.language, role.String()}))
	meaning := contracts.MeaningKey(contracts.StableID("native-meaning", struct {
		Definition    contracts.DefinitionKey
		Role, Version string
	}{definition, role.String(), contracts.TrackAPolicy}))
	return contracts.ResolvedBinding{Occurrence: contracts.BindingOccurrenceID(contracts.StableID("binding-use", origin)), Definition: definition, Meaning: meaning, Spelling: string(tok.rawBytes), CanonicalName: "native:" + role.String(), LexicalRole: tok.tType, Target: target, Origin: origin, DefinitionOrigin: contracts.Origin{Source: "resource:lexicon/" + r.compilation.Gram.language + "/vocabulary.json", Label: role.String()}, DefinitionEnvironment: contracts.BindingEnvironmentKey(r.configurationIdentity()), UseEnvironment: environment, Policy: contracts.TrackAPolicy, SelectionReason: "configured native semantic role retained by the typed instruction planner", CandidateFingerprint: string(environment), Dependencies: []string{string(environment), r.configurationIdentity()}}
}

func (r *contractRecorder) definitionCatalog() []ContractDefinition {
	records := make([]*record, 0, len(r.compilation.Semantics.Declarations))
	for rec := range r.compilation.Semantics.Declarations {
		if r.compilation.Semantics.Aliases[rec] == nil {
			records = append(records, rec)
		}
	}
	sort.Slice(records, func(i, j int) bool { return r.definition(records[i]) < r.definition(records[j]) })
	var result []ContractDefinition
	for _, rec := range records {
		scope := rec.storageScope
		if tok := r.declarationTokens[rec]; tok != nil {
			scope = tok.scope
		}
		base := ContractDefinition{Definition: r.definition(rec), Meaning: r.meaning(rec, nil), Name: rec.ID, Role: rec.Type, Origin: r.definitionOrigin(rec), Environment: r.environment(scope)}
		if descriptor := r.compilation.Semantics.Descriptors[rec]; descriptor != nil {
			base.Descriptor = true
			if len(descriptor.Rules) == 0 {
				result = append(result, base)
			}
			for _, rule := range descriptor.Rules {
				entry := base
				entry.Meaning = r.meaning(rec, rule)
				entry.Applicability = rule.Target.String()
				if rule.Source != nil && rule.Source.source != nil {
					source := rule.Source.source
					entry.Origin = contracts.Origin{Source: r.logicalSource(source.Span.Path), Start: source.Span.Start.Offset, End: source.Span.End.Offset, Parent: r.occurrence(source.occurrence)}
				}
				result = append(result, entry)
			}
		} else {
			result = append(result, base)
		}
	}
	return result
}

func (r *contractRecorder) instructionBindings() []contracts.ResolvedBinding {
	var result []contracts.ResolvedBinding
	seenNative := make(map[*token]bool)
	var visit func([]*instruction, contracts.BindingOccurrenceID, []int, contracts.BindingEnvironmentKey)
	visit = func(instructions []*instruction, parent contracts.BindingOccurrenceID, path []int, callerEnvironment contracts.BindingEnvironmentKey) {
		for index, current := range instructions {
			if current == nil {
				continue
			}
			order := append(append([]int(nil), path...), index)
			tok := current.SourceToken
			useEnvironment := callerEnvironment
			if useEnvironment == "" && tok != nil {
				useEnvironment = r.environment(tok.scope)
			}
			origin := r.origin(tok)
			if parent != "" {
				origin.Parent = parent
				origin.Expansion = append(append([]int(nil), origin.Expansion...), order...)
				origin.Label = "expanded typed instruction"
			}
			target := r.valueTarget(current.Target)
			if target.Unknown && current.Value != nil {
				target = r.valueTarget(current.Value)
			}
			use := contracts.BindingOccurrenceID(contracts.StableID("binding-use", origin))
			if tok != nil {
				if tok.binding == nil {
					if role, ok := tok.nativeRole(); ok {
						binding := r.nativeBinding(tok, role, target, origin, r.environment(tok.scope))
						binding.UseEnvironment = useEnvironment
						binding.Dependencies = append(binding.Dependencies, string(useEnvironment))
						result = append(result, binding)
						seenNative[tok] = true
					}
				} else if parent != "" {
					binding := r.binding(tok, tok.binding.Record, nil, target, 0)
					binding.Origin = origin
					binding.Occurrence = contracts.BindingOccurrenceID(contracts.StableID("binding-use", origin))
					binding.UseEnvironment = useEnvironment
					binding.Dependencies = append(binding.Dependencies, string(useEnvironment))
					binding.SelectionReason = "definition-site resolved word instantiated at an explicit typed callsite"
					result = append(result, binding)
				}
			}
			seenValues := make(map[*instructionValue]bool)
			var operands func(*instructionValue, []int)
			operands = func(value *instructionValue, position []int) {
				if value == nil || seenValues[value] {
					return
				}
				seenValues[value] = true
				if value.Token != nil && value.Record != nil && parent != "" {
					binding := r.binding(value.Token, value.Record, nil, r.valueTarget(value), 0)
					binding.Origin.Parent = use
					binding.Origin.Expansion = append(append([]int(nil), binding.Origin.Expansion...), position...)
					binding.Origin.Label = "expanded typed operand"
					binding.Occurrence = contracts.BindingOccurrenceID(contracts.StableID("binding-use", binding.Origin))
					binding.UseEnvironment = useEnvironment
					binding.Dependencies = append(binding.Dependencies, string(useEnvironment))
					binding.SelectionReason = "definition-site operand binding retained through typed source-action substitution"
					result = append(result, binding)
				}
				operands(value.Owner, append(append([]int(nil), position...), 0))
				operands(value.Payload, append(append([]int(nil), position...), 1))
				operands(value.Argument, append(append([]int(nil), position...), 2))
				operands(value.Key, append(append([]int(nil), position...), 3))
				for i, element := range value.Elements {
					operands(element, append(append([]int(nil), position...), 4, i))
				}
			}
			operands(current.Target, []int{0})
			operands(current.Value, []int{1})
			for i, argument := range current.Arguments {
				operands(argument, []int{2, i})
			}
			visit(current.Actions, use, append(order, 0), useEnvironment)
			visit(current.Then, parent, append(order, 1), callerEnvironment)
			visit(current.Else, parent, append(order, 2), callerEnvironment)
			if current.Block != nil {
				visit(current.Block.Body, parent, append(order, 3), callerEnvironment)
			}
		}
	}
	visit(r.compilation.Semantics.Instructions, "", nil, "")
	for _, statement := range r.compilation.Statements {
		for _, tok := range statement.tokens {
			if tok.binding == nil && !seenNative[tok] {
				if role, ok := tok.nativeRole(); ok {
					origin := r.origin(tok)
					binding := r.nativeBinding(tok, role, contracts.TargetRef{Unknown: true, Phase: contracts.Elaboration}, origin, r.environment(tok.scope))
					binding.SelectionReason = "configured native semantic role in a declaration or unattached lexical context"
					result = append(result, binding)
				}
			}
		}
	}
	return result
}

// ResolveContractBindings runs the same preparation, resolver and construction
// pipeline as Compile while observing its decisions. It never builds or runs an
// executable. Its returned Compilation retains the actual graph for analysis.
func (ls *LygerService) ResolveContractBindings(source []byte, path string) (*Compilation, ContractBindings, error) {
	result, err := ls.CompileForContracts(source, path)
	if err != nil {
		return nil, ContractBindings{}, err
	}
	c := result.Compilation
	return c, c.contractRecorder.snapshot(), nil
}
