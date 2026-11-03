package contracts

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

var contractEnums = map[reflect.Type][]string{
	reflect.TypeOf(Coverage("")):          {string(Known), string(Partial), string(Unknown), string(NotApplicable)},
	reflect.TypeOf(Status("")):            {string(Established), string(Contradicted), string(Unresolved)},
	reflect.TypeOf(Outcome("")):           {string(Normal), string(ReturnedFailure), string(GuardRejected), string(Panic), string(Abort), string(Cancelled), string(ProcessFailure)},
	reflect.TypeOf(Phase("")):             {"", string(Elaboration), string(Construction), string(Invocation), string(Destruction), string(Cleanup)},
	reflect.TypeOf(EffectKind("")):        {"read", "write", "allocate", "deallocate", "release", "io", "spawn", "transfer", "wait", "block", "suspend", "time", "random", "environment", "foreign", "panic", "opaque", "synchronize", "replace", "cleanup"},
	reflect.TypeOf(ProtocolAuthority("")): {string(TransitionOwner), string(Observer), string(SharedTransition), string(UnknownAuthority)},
	reflect.TypeOf(AppStage("")):          {string(AppIdle), string(AppRunning), string(AppDraining), string(AppFinalized)},
	reflect.TypeOf(ContinuationClaim("")): {string(ClaimUnclaimed), string(ClaimActive), string(ClaimFinished)},
	reflect.TypeOf(NodeKind("")):          {string(Sequence), string(Branch), string(Event), string(Call), string(Callback), string(Native), string(Guard), string(Require), string(Acquire), string(Move), string(Release), string(Scope), string(Repeat), string(Mutation), string(Return), string(CheckedResult)},
}

func documentModel(kind string) (reflect.Type, error) {
	switch kind {
	case "native_environment":
		return reflect.TypeOf(NativeEnvironment{}), nil
	case "contract_check_requests":
		return reflect.TypeOf(CheckRequests{}), nil
	case "native_contract_manifest":
		return reflect.TypeOf(NativeContractManifest{}), nil
	case "facet_and_role_manifest":
		return reflect.TypeOf(FacetRoleManifest{}), nil
	case "trust_policy":
		return reflect.TypeOf(TrustPolicyDocument{}), nil
	case "contract_report":
		return reflect.TypeOf(Report{}), nil
	case "adjective_certificate":
		return reflect.TypeOf(AdjectiveCertificateDocument{}), nil
	default:
		return nil, fmt.Errorf("unsupported schema document kind %q", kind)
	}
}

// JSONSchema derives every property and enum from the actual interchange model.
// Semantic reference/provenance checks remain in the same types' Validate methods.
func JSONSchema(kind string) ([]byte, error) {
	model, err := documentModel(kind)
	if err != nil {
		return nil, err
	}
	definitions := make(map[string]any)
	var schema func(reflect.Type) map[string]any
	schema = func(typ reflect.Type) map[string]any {
		if values, exists := contractEnums[typ]; exists {
			return map[string]any{"type": "string", "enum": values}
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
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)
				if !field.IsExported() {
					continue
				}
				parts := strings.Split(field.Tag.Get("json"), ",")
				if parts[0] == "-" {
					continue
				}
				fieldName := parts[0]
				if fieldName == "" {
					fieldName = field.Name
				}
				properties[fieldName] = schema(field.Type)
				if !containsString(parts[1:], "omitempty") {
					required = append(required, fieldName)
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
			for i := 0; i < current.NumField(); i++ {
				field := current.Type().Field(i)
				if !field.IsExported() || field.Tag.Get("json") == "-" {
					continue
				}
				if containsString(strings.Split(field.Tag.Get("json"), ",")[1:], "omitempty") && current.Field(i).IsZero() {
					continue
				}
				if err := visit(current.Field(i), path+"."+field.Name, depth+1); err != nil {
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
