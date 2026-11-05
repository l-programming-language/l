package contracts

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var contractEnums = map[reflect.Type][]string{
	reflect.TypeOf(Coverage("")):          {string(Known), string(Partial), string(Unknown), string(NotApplicable)},
	reflect.TypeOf(Status("")):            {string(Established), string(Contradicted), string(Unresolved)},
	reflect.TypeOf(Outcome("")):           {string(Normal), string(ReturnedFailure), string(GuardRejected), string(Panic), string(Abort), string(Cancelled), string(ProcessFailure)},
	reflect.TypeOf(Phase("")):             {"", string(Elaboration), string(Construction), string(Invocation), string(Destruction), string(Cleanup)},
	reflect.TypeOf(EffectKind("")):        {"read", "write", "allocate", "deallocate", "release", "io", "spawn", string(EffectSignal), "transfer", "wait", "block", "suspend", "time", "random", "environment", "foreign", "panic", "opaque", "synchronize", "replace", "cleanup"},
	reflect.TypeOf(ProtocolAuthority("")): {string(TransitionOwner), string(Observer), string(SharedTransition), string(UnknownAuthority)},
	reflect.TypeOf(AppStage("")):          {string(AppIdle), string(AppRunning), string(AppDraining), string(AppFinalized)},
	reflect.TypeOf(ContinuationClaim("")): {string(ClaimUnclaimed), string(ClaimActive), string(ClaimFinished)},
	reflect.TypeOf(NodeKind("")):          {string(Sequence), string(Branch), string(Event), string(Call), string(Callback), string(Native), string(Guard), string(Require), string(Acquire), string(Move), string(Release), string(Scope), string(Repeat), string(Mutation), string(Return), string(CheckedResult)},
}

// The same model selects the strict decoder and the offline schema.
var documentModels = map[string]reflect.Type{
	"native_environment":       reflect.TypeOf(NativeEnvironment{}),
	"contract_check_requests":  reflect.TypeOf(CheckRequests{}),
	"native_contract_manifest": reflect.TypeOf(NativeContractManifest{}),
	"facet_and_role_manifest":  reflect.TypeOf(FacetRoleManifest{}),
	"trust_policy":             reflect.TypeOf(TrustPolicyDocument{}),
	"contract_report":          reflect.TypeOf(Report{}),
	"adjective_certificate":    reflect.TypeOf(AdjectiveCertificateDocument{}),
}

type modelField struct {
	name     string
	typ      reflect.Type
	index    int
	optional bool
}

var fieldLayouts sync.Map // reflect.Type -> []modelField, immutable after publication

func modelFields(typ reflect.Type) []modelField {
	cached, _ := fieldLayouts.Load(typ)
	if fields, ok := cached.([]modelField); ok {
		return fields
	}
	fields := make([]modelField, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
		if !field.IsExported() || name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		fields = append(fields, modelField{name, field.Type, i, containsString(strings.Split(options, ","), "omitempty")})
	}
	fieldLayouts.Store(typ, fields)
	return fields
}

// JSONSchema derives every property and enum from the actual interchange model.
// Semantic reference/provenance checks remain in the same types' Validate methods.
func JSONSchema(kind string) ([]byte, error) {
	model := documentModels[kind]
	if model == nil {
		return nil, fmt.Errorf("unsupported contract document kind %q", kind)
	}
	definitions := make(map[string]any)
	var schema func(reflect.Type) map[string]any
	schema = func(typ reflect.Type) map[string]any {
		if values, exists := contractEnums[typ]; exists {
			definitions[typ.Name()] = map[string]any{"type": "string", "enum": values}
			return map[string]any{"$ref": "#/$defs/" + typ.Name()}
		}
		if typ.Kind() == reflect.Pointer {
			return map[string]any{"anyOf": []any{schema(typ.Elem()), map[string]any{"type": "null"}}}
		}
		switch typ.Kind() {
		case reflect.String:
			return map[string]any{"type": "string"}
		case reflect.Bool:
			return map[string]any{"type": "boolean"}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return map[string]any{"type": "integer"}
		case reflect.Float32, reflect.Float64:
			return map[string]any{"type": "number"}
		case reflect.Slice, reflect.Array:
			return map[string]any{"type": []string{"array", "null"}, "items": schema(typ.Elem())}
		case reflect.Map:
			return map[string]any{"type": []string{"object", "null"}, "additionalProperties": schema(typ.Elem())}
		case reflect.Struct:
			name := typ.Name()
			if _, exists := definitions[name]; exists {
				return map[string]any{"$ref": "#/$defs/" + name}
			}
			definitions[name] = map[string]any{}
			properties := make(map[string]any)
			var required []string
			for _, field := range modelFields(typ) {
				properties[field.name] = schema(field.typ)
				if !field.optional {
					required = append(required, field.name)
				}
			}
			definition := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
			if len(required) != 0 {
				definition["required"] = required
			}
			definitions[name] = definition
			return map[string]any{"$ref": "#/$defs/" + name}
		default:
			return map[string]any{}
		}
	}
	root := schema(model)
	root["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	root["$id"] = "l-contracts/" + SchemaVersion + "/" + kind
	root["$defs"] = definitions
	definition, ok := definitions[model.Name()].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing document definition for %s", model.Name())
	}
	properties, ok := definition["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing document properties for %s", model.Name())
	}
	properties["schema_version"] = map[string]any{"const": SchemaVersion}
	properties["kind"] = map[string]any{"const": kind}
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func validateModelEnums(value any) error {
	var visit func(reflect.Value, string, int) error
	visit = func(current reflect.Value, path string, depth int) error {
		if !current.IsValid() {
			return nil
		}
		if depth > 256 {
			return fmt.Errorf("model nesting exceeds supported bound at %s", path)
		}
		if values, exists := contractEnums[current.Type()]; exists {
			if !containsString(values, current.String()) {
				return fmt.Errorf("invalid %s value %q at %s", current.Type().Name(), current.String(), path)
			}
			return nil
		}
		switch current.Kind() {
		case reflect.Pointer, reflect.Interface:
			if !current.IsNil() {
				return visit(current.Elem(), path, depth+1)
			}
		case reflect.Struct:
			for _, field := range modelFields(current.Type()) {
				value := current.Field(field.index)
				if field.optional && value.IsZero() {
					continue
				}
				if err := visit(value, path+"."+field.name, depth+1); err != nil {
					return err
				}
			}
		case reflect.Slice, reflect.Array:
			for i := 0; i < current.Len(); i++ {
				if err := visit(current.Index(i), fmt.Sprintf("%s[%d]", path, i), depth+1); err != nil {
					return err
				}
			}
		case reflect.Map:
			iterator := current.MapRange()
			for iterator.Next() {
				if err := visit(iterator.Value(), path+"["+fmt.Sprint(iterator.Key().Interface())+"]", depth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return visit(reflect.ValueOf(value), "document", 0)
}
