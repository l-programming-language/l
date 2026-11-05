package contracts

// ProfileDescriptor selects a fixed checker and its required request parameters.
// Values are copied out of the table; callers cannot change profile semantics.
type ProfileDescriptor struct {
	ForbiddenEffect   EffectKind
	RequiresFootprint bool
	WholeEntryOnly    bool
}

var profileDescriptors = map[string]ProfileDescriptor{
	"NoHeapAllocation":   {ForbiddenEffect: "allocate"},
	"NoBlocking":         {ForbiddenEffect: "block"},
	"NoExternalIO":       {ForbiddenEffect: "io"},
	"NoPanic":            {ForbiddenEffect: "panic"},
	"OwnedLifecycle":     {WholeEntryOnly: true},
	"PreservesOnFailure": {RequiresFootprint: true},
}

func LookupProfile(name string) (ProfileDescriptor, bool) {
	descriptor, ok := profileDescriptors[name]
	return descriptor, ok
}

func (descriptor ProfileDescriptor) SupportsScope(scope string) bool {
	if scope == "" || scope == "whole_region" {
		return true
	}
	if descriptor.WholeEntryOnly {
		return false
	}
	switch scope {
	case "body_only", "invocation", "source_invocation", "source_operation", "native_body":
		return true
	}
	return false
}
