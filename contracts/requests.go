package contracts

import "fmt"

type CheckRequest struct {
	ID        string            `json:"id"`
	Scope     string            `json:"scope"`
	Profile   Profile           `json:"profile"`
	Footprint *FailureFootprint `json:"footprint,omitempty"`
}

type CheckRequests struct {
	SchemaVersion string         `json:"schema_version"`
	Kind          string         `json:"kind"`
	Producer      string         `json:"producer"`
	Version       string         `json:"version"`
	Checks        []CheckRequest `json:"checks"`
}

type CheckResult struct {
	Request     CheckRequest   `json:"request"`
	Scope       *AnalysisScope `json:"scope,omitempty"`
	Status      Status         `json:"status"`
	Summary     DerivedSummary `json:"summary"`
	Obligations []Obligation   `json:"obligations"`
	Required    []string       `json:"required"`
}

func (document CheckRequests) Validate() error {
	if err := validateDocumentHeader(document.SchemaVersion, document.Kind, "contract_check_requests", document.Producer); err != nil {
		return err
	}
	if document.Version != "1" || len(document.Checks) == 0 {
		return fmt.Errorf("check requests require version 1 and at least one check")
	}
	seen := make(map[string]bool)
	for _, check := range document.Checks {
		if check.ID == "" || seen[check.ID] || check.Scope == "" {
			return fmt.Errorf("check requests require unique IDs and explicit scopes")
		}
		seen[check.ID] = true
		if !SupportedProfile(check.Profile.Name) && check.Profile.Name != "PreservesOnFailure" && check.Profile.Name != "OwnedLifecycle" {
			return fmt.Errorf("unsupported requested profile %s", check.Profile.Name)
		}
		if check.Profile.Name == "PreservesOnFailure" {
			if check.Footprint == nil || len(check.Footprint.Locations) == 0 || len(check.Footprint.Outcomes) == 0 || check.Footprint.Observation == "" {
				return fmt.Errorf("PreservesOnFailure requires a nonempty resource footprint, observation and failure outcomes")
			}
			if check.Footprint.Observation != "state_at_completion" {
				return fmt.Errorf("unsupported failure observation model %q", check.Footprint.Observation)
			}
			if len(check.Profile.Outcomes) > 0 && StableID("outcomes", check.Profile.Outcomes) != StableID("outcomes", check.Footprint.Outcomes) {
				return fmt.Errorf("profile and footprint outcomes disagree")
			}
			for _, outcome := range check.Footprint.Outcomes {
				if outcome == Normal {
					return fmt.Errorf("PreservesOnFailure requires failure outcomes")
				}
			}
			for _, location := range check.Footprint.Locations {
				if location.Resource.ID == "" || location.Resource.Unknown {
					return fmt.Errorf("failure footprint requires resolved resource identities")
				}
			}
		} else if check.Footprint != nil {
			return fmt.Errorf("footprint is only supported for PreservesOnFailure")
		}
	}
	return validateModelEnums(document)
}
