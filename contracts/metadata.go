package contracts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

type DeclaredEffectRow struct {
	Effects     []Effect `json:"effects,omitempty"`
	UnknownTail *bool    `json:"unknown_tail,omitempty"`
	Parameters  []string `json:"parameters,omitempty"`
}

type OwnershipDeclaration struct {
	Access    []string `json:"access,omitempty"`
	Escape    []string `json:"escape,omitempty"`
	Release   []string `json:"release,omitempty"`
	Authority []string `json:"authority,omitempty"`
}

type NativeSignature struct {
	Inputs      []string     `json:"inputs"`
	Outputs     []string     `json:"outputs"`
	InputTypes  []NativeType `json:"input_types,omitempty"`
	OutputTypes []NativeType `json:"output_types,omitempty"`
}

// DeclaredContract cannot contain Status or evidence. Favorable claims here
// remain requirements on the implementation until separately discharged.
type DeclaredContract struct {
	ID                 string                  `json:"id"`
	Meaning            MeaningKey              `json:"meaning"`
	Name               string                  `json:"name"`
	Signature          *NativeSignature        `json:"signature,omitempty"`
	Requirements       []Predicate             `json:"requirements,omitempty"`
	Guarantees         []Claim                 `json:"guarantees,omitempty"`
	Effects            *DeclaredEffectRow      `json:"effects,omitempty"`
	Costs              *ResourceCost           `json:"costs,omitempty"`
	Demand             map[string]Multiplicity `json:"demand,omitempty"`
	Ownership          *OwnershipDeclaration   `json:"ownership,omitempty"`
	Outcomes           []Outcome               `json:"outcomes,omitempty"`
	OutcomeContracts   []NativeOutcomeContract `json:"outcome_contracts,omitempty"`
	Frames             []FrameContract         `json:"frames,omitempty"`
	Coverage           map[string]Coverage     `json:"coverage,omitempty"`
	Dependencies       []string                `json:"dependencies,omitempty"`
	TargetRestrictions []Predicate             `json:"target_restrictions,omitempty"`
	Origin             Origin                  `json:"origin"`
}

type NativeContractManifest struct {
	SchemaVersion string             `json:"schema_version"`
	Kind          string             `json:"kind"`
	Producer      string             `json:"producer"`
	Declarations  []DeclaredContract `json:"declarations"`
}

type FacetRoleManifest struct {
	SchemaVersion      string                         `json:"schema_version"`
	Kind               string                         `json:"kind"`
	Producer           string                         `json:"producer"`
	Facets             []FacetSchema                  `json:"facets"`
	Roles              []RoleInterface                `json:"roles"`
	Transformers       []ContractTransformer          `json:"transformers,omitempty"`
	SourceTransformers []SourceTransformerDeclaration `json:"source_transformers,omitempty"`
}

type TrustPolicyDocument struct {
	SchemaVersion string          `json:"schema_version"`
	Kind          string          `json:"kind"`
	Producer      string          `json:"producer"`
	Policy        TrustPolicy     `json:"policy"`
	Witnesses     []NativeWitness `json:"witnesses"`
}

type AdjectiveCertificateDocument struct {
	SchemaVersion string                 `json:"schema_version"`
	Kind          string                 `json:"kind"`
	Producer      string                 `json:"producer"`
	Certificates  []AdjectiveCertificate `json:"certificates"`
}

type NormalizedNativeContract struct {
	Declaration DeclaredContract    `json:"declaration"`
	Coverage    map[string]Coverage `json:"coverage"`
	Effects     DeclaredEffectRow   `json:"effects"`
	Unknowns    []UnknownReason     `json:"unknowns"`
	Assurance   Status              `json:"assurance"`
}

func NormalizeNativeContract(declaration DeclaredContract) NormalizedNativeContract {
	unknown := true
	result := NormalizedNativeContract{Declaration: declaration, Coverage: make(map[string]Coverage), Effects: DeclaredEffectRow{UnknownTail: &unknown}, Assurance: Unresolved}
	for i := 1; i <= 10; i++ {
		family := fmt.Sprintf("C%d", i)
		coverage := declaration.Coverage[family]
		if coverage == "" || coverage == NotApplicable {
			coverage = Unknown
			message := "native contract family was not supplied"
			if declaration.Coverage[family] == NotApplicable {
				message = "authored not_applicable assertion requires an independent supported applicability decision"
			}
			result.Unknowns = append(result.Unknowns, UnknownReason{Code: "native_coverage", Message: family + ": " + message, Boundary: declaration.ID, Origin: declaration.Origin, Components: []string{family}})
		}
		result.Coverage[family] = coverage
	}
	present := map[string]bool{"C1": declaration.Signature != nil, "C2": declaration.Demand != nil, "C3": declaration.Ownership != nil, "C4": declaration.Effects != nil, "C5": len(declaration.Outcomes) != 0, "C6": false, "C7": len(declaration.Frames) != 0, "C8": declaration.Costs != nil, "C9": declaration.Signature != nil, "C10": declaration.Meaning != ""}
	for family, supplied := range present {
		if !supplied {
			result.Coverage[family] = Unknown
		}
	}
	if declaration.Effects != nil {
		result.Effects = *declaration.Effects
		if result.Effects.UnknownTail == nil {
			result.Effects.UnknownTail = &unknown
		}
	}
	if *result.Effects.UnknownTail {
		result.Coverage["C4"] = Unknown
		result.Unknowns = append(result.Unknowns, UnknownReason{Code: "native_effect_tail", Message: "native effect row remains open", Boundary: declaration.ID, Origin: declaration.Origin, Components: []string{"effects"}})
	}
	result.Unknowns = append(result.Unknowns, UnknownReason{Code: "unapproved_native", Message: "declaration has no implementation-conformance evidence by itself", Boundary: declaration.ID, Origin: declaration.Origin})
	return result
}

func validateDocumentHeader(version, kind, expected, producer string) error {
	if version != SchemaVersion {
		return fmt.Errorf("unsupported contracts schema version %q", version)
	}
	if kind != expected {
		return fmt.Errorf("expected document kind %q, got %q", expected, kind)
	}
	if producer == "" {
		return fmt.Errorf("contract document requires producer/version identity")
	}
	return nil
}

func (manifest NativeContractManifest) Validate() error {
	if err := validateDocumentHeader(manifest.SchemaVersion, manifest.Kind, "native_contract_manifest", manifest.Producer); err != nil {
		return err
	}
	if err := validateModelEnums(manifest); err != nil {
		return err
	}
	identities, meanings := make(map[string]bool), make(map[MeaningKey]bool)
	for _, declaration := range manifest.Declarations {
		if declaration.ID == "" || declaration.Meaning == "" || declaration.Name == "" || identities[declaration.ID] || meanings[declaration.Meaning] {
			return fmt.Errorf("native declaration requires unambiguous canonical contract and meaning identities")
		}
		identities[declaration.ID], meanings[declaration.Meaning] = true, true
		outcomes := make(map[Outcome]bool)
		for _, outcome := range declaration.Outcomes {
			if outcomes[outcome] {
				return fmt.Errorf("duplicate native outcome %s", outcome)
			}
			outcomes[outcome] = true
		}
		components := make(map[Outcome]bool)
		for _, component := range declaration.OutcomeContracts {
			if !outcomes[component.Outcome] || components[component.Outcome] {
				return fmt.Errorf("native outcome components require unique declared outcomes")
			}
			components[component.Outcome] = true
			for _, relation := range component.States {
				if relation.Location.Resource.ID == "" || relation.Location.Resource.Unknown || relation.Observation != "state_at_completion" || !containsString([]string{"unchanged", "different", "unknown"}, relation.Relation) || len(relation.Evidence) > 0 {
					return fmt.Errorf("native state declarations require an exact resource, supported observation/relation and no manufactured evidence")
				}
			}
			for _, absent := range component.AbsentEffects {
				if component.Effects != nil {
					for _, effect := range component.Effects.Effects {
						if effect.Kind == absent {
							return fmt.Errorf("native effect both declared and excluded: %s", absent)
						}
					}
				}
			}
		}
		for family := range declaration.Coverage {
			valid := false
			for i := 1; i <= 10; i++ {
				valid = valid || family == fmt.Sprintf("C%d", i)
			}
			if !valid {
				return fmt.Errorf("unknown native contract family %q", family)
			}
		}
		for _, dependency := range declaration.Dependencies {
			if dependency == "" {
				return fmt.Errorf("native declaration has empty dependency reference")
			}
		}
	}
	return nil
}

func (manifest FacetRoleManifest) Validate() error {
	if err := validateDocumentHeader(manifest.SchemaVersion, manifest.Kind, "facet_and_role_manifest", manifest.Producer); err != nil {
		return err
	}
	if err := validateModelEnums(manifest); err != nil {
		return err
	}
	sourceMeanings := make(map[MeaningKey]bool)
	for _, declaration := range manifest.SourceTransformers {
		if declaration.Meaning == "" || declaration.Version == "" || declaration.TargetNominal == "" || sourceMeanings[declaration.Meaning] {
			return fmt.Errorf("source transformer requires unique canonical rule, version and nominal identity")
		}
		sourceMeanings[declaration.Meaning] = true
	}
	seen := make(map[string]bool)
	for _, facet := range manifest.Facets {
		if seen[facet.ID] {
			return fmt.Errorf("duplicate facet reference %s", facet.ID)
		}
		seen[facet.ID] = true
		if err := facet.Validate(); err != nil {
			return err
		}
	}
	seen = make(map[string]bool)
	for _, role := range manifest.Roles {
		if seen[role.ID] {
			return fmt.Errorf("duplicate role reference %s", role.ID)
		}
		seen[role.ID] = true
		if err := role.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (document TrustPolicyDocument) Validate() error {
	if err := validateDocumentHeader(document.SchemaVersion, document.Kind, "trust_policy", document.Producer); err != nil {
		return err
	}
	if document.Policy.Version != SchemaVersion || document.Policy.ID == "" {
		return fmt.Errorf("trust policy requires supported version and identity")
	}
	seen := make(map[string]NativeWitness)
	for _, witness := range document.Witnesses {
		if _, exists := seen[witness.ID]; exists || witness.ID == "" {
			return fmt.Errorf("duplicate or empty native witness identity")
		}
		if witness.Authority == "" || witness.Rationale == "" || witness.ClaimFingerprint == "" || !validEvidenceFingerprints(EvidenceFingerprints{Contract: witness.ContractFingerprint, Implementation: witness.ImplementationFingerprint, Target: witness.TargetFingerprint, Dependencies: witness.Dependencies}) {
			return fmt.Errorf("native witness %s lacks proposition, authority, rationale or current dependencies", witness.ID)
		}
		seen[witness.ID] = witness
	}
	for _, identity := range document.Policy.ApprovedWitnesses {
		witness, exists := seen[identity]
		if !exists || !containsString(document.Policy.Authorities, witness.Authority) {
			return fmt.Errorf("policy approval %s has missing witness or unapproved authority", identity)
		}
	}
	return nil
}

func (document AdjectiveCertificateDocument) Validate() error {
	if err := validateDocumentHeader(document.SchemaVersion, document.Kind, "adjective_certificate", document.Producer); err != nil {
		return err
	}
	if err := validateModelEnums(document); err != nil {
		return err
	}
	seen := make(map[MeaningKey]bool)
	for _, certificate := range document.Certificates {
		if certificate.Meaning == "" || certificate.TransformerFingerprint == "" || seen[certificate.Meaning] {
			return fmt.Errorf("ambiguous or incomplete adjective certificate identity")
		}
		seen[certificate.Meaning] = true
		if certificate.Status == Established && (!AllEstablished(certificate.Obligations) || certificate.Mode != "symbolic") {
			return fmt.Errorf("adjective certificate hides unresolved obligations or unsupported symbolic analysis")
		}
	}
	return nil
}

func ReadNativeManifest(data []byte) (NativeContractManifest, error) {
	value, err := DecodeDocument(data)
	if err != nil {
		return NativeContractManifest{}, err
	}
	manifest, ok := value.(NativeContractManifest)
	if !ok {
		return NativeContractManifest{}, fmt.Errorf("expected native_contract_manifest")
	}
	return manifest, nil
}

func DecodeDocument(data []byte) (any, error) {
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return nil, err
	}
	var header struct {
		SchemaVersion string `json:"schema_version"`
		Kind          string `json:"kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("contract document: %w", err)
	}
	if header.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("unsupported contracts schema version %q", header.SchemaVersion)
	}
	var target any
	switch header.Kind {
	case "native_environment":
		target = &NativeEnvironment{}
	case "contract_check_requests":
		target = &CheckRequests{}
	case "native_contract_manifest":
		target = &NativeContractManifest{}
	case "facet_and_role_manifest":
		target = &FacetRoleManifest{}
	case "trust_policy":
		target = &TrustPolicyDocument{}
	case "contract_report":
		target = &Report{}
	case "adjective_certificate":
		target = &AdjectiveCertificateDocument{}
	default:
		return nil, fmt.Errorf("unsupported contract document kind %q", header.Kind)
	}
	if err := validateJSONShape(data, reflect.TypeOf(target).Elem(), "document", 0); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return nil, fmt.Errorf("contract document: %w", err)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return nil, fmt.Errorf("contract document requires exactly one JSON value")
	}
	if err := validateDocument(target); err != nil {
		return nil, err
	}
	return reflect.ValueOf(target).Elem().Interface(), nil
}

func EncodeDocument(document any) ([]byte, error) {
	if err := validateDocument(document); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func validateDocument(document any) error {
	validatable, ok := document.(interface{ Validate() error })
	if !ok {
		return fmt.Errorf("unsupported contracts document type %T", document)
	}
	return validatable.Validate()
}

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value func(int) error
	value = func(depth int) error {
		if depth > 256 {
			return fmt.Errorf("contract JSON nesting exceeds supported bound")
		}
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delimiter, compound := token.(json.Delim)
		if !compound {
			return nil
		}
		switch delimiter {
		case '{':
			keys := make(map[string]bool)
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok || keys[key] {
					return fmt.Errorf("duplicate or invalid contract JSON key %q", key)
				}
				keys[key] = true
				if err := value(depth + 1); err != nil {
					return err
				}
			}
		case '[':
			for decoder.More() {
				if err := value(depth + 1); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unexpected contract JSON delimiter %q", delimiter)
		}
		_, err = decoder.Token()
		return err
	}
	if err := value(0); err != nil {
		return fmt.Errorf("contract JSON: %w", err)
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("contract JSON requires exactly one value")
	}
	return nil
}

func validateJSONShape(data []byte, typ reflect.Type, path string, depth int) error {
	if depth > 256 {
		return fmt.Errorf("contract document exceeds supported depth at %s", path)
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		if typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map {
			return nil
		}
		return fmt.Errorf("required contract value %s cannot be null", path)
	}
	if typ.Kind() == reflect.Pointer {
		return validateJSONShape(data, typ.Elem(), path, depth+1)
	}
	switch typ.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return fmt.Errorf("%s requires an object: %w", path, err)
		}
		expected := make(map[string]reflect.StructField)
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			if !field.IsExported() {
				continue
			}
			parts := strings.Split(field.Tag.Get("json"), ",")
			if parts[0] == "-" {
				continue
			}
			name := parts[0]
			if name == "" {
				name = field.Name
			}
			expected[name] = field
			if _, exists := fields[name]; !exists && !containsString(parts[1:], "omitempty") {
				return fmt.Errorf("required contract field %s.%s is missing", path, name)
			}
		}
		for name, raw := range fields {
			field, exists := expected[name]
			if !exists {
				return fmt.Errorf("unknown contract field %s.%s", path, name)
			}
			if err := validateJSONShape(raw, field.Type, path+"."+name, depth+1); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		var values []json.RawMessage
		if err := json.Unmarshal(data, &values); err != nil {
			return err
		}
		for i, raw := range values {
			if err := validateJSONShape(raw, typ.Elem(), fmt.Sprintf("%s[%d]", path, i), depth+1); err != nil {
				return err
			}
		}
	case reflect.Map:
		var values map[string]json.RawMessage
		if err := json.Unmarshal(data, &values); err != nil {
			return err
		}
		for name, raw := range values {
			if err := validateJSONShape(raw, typ.Elem(), path+"["+name+"]", depth+1); err != nil {
				return err
			}
		}
	}
	return nil
}
