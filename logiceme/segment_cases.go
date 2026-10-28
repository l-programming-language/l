package Logiceme

type segmentCase struct {
	condition string
	tag       *LucidTag
}

// The selected resource owns immutable parsed cases. Rendering receives a
// fresh tree for each selected case, never shared between glyphs or providers.
type segmentCases struct {
	source string
	cases  []segmentCase
}

func parseSegmentCases(content []byte) *segmentCases {
	parsed := &segmentCases{source: string(content)}
	for _, tag := range CompileMainTags(content) {
		if len(tag.Expressions) > 1 {
			parsed.cases = append(parsed.cases, segmentCase{tag.Expressions[1].ExpressionString, tag})
		}
	}
	return parsed
}

func (s Segment) hasTemplateCase(system *SegmentalSystem, selected string) bool {
	condition := `if Type == "` + selected + `"`
	// Preserve caller-supplied Tags and direct Content mutation. A copied
	// Segment cannot reuse framing for different bytes.
	tags := s.Tags
	if len(tags) == 0 {
		if cases := s.loadedCases(system); cases != nil {
			for _, candidate := range cases.cases {
				if candidate.condition == condition {
					return true
				}
			}
			return false
		}
		tags = CompileMainTags(s.Content)
	}
	for _, tag := range tags {
		if len(tag.Expressions) > 1 && tag.Expressions[1].ExpressionString == condition {
			return true
		}
	}
	return false
}

func (s Segment) selectedTemplateTags(system *SegmentalSystem, selected string, headers bool) []*LucidTag {
	cases := s.loadedCases(system)
	if cases == nil {
		return CompileMainTags(s.Content)
	}
	condition := `if Type == "` + selected + `"`
	var tags []*LucidTag
	for _, candidate := range cases.cases {
		if candidate.condition == condition || headers && (candidate.condition == "Headers" || candidate.condition == "Inputs") {
			tags = append(tags, cloneSegmentTag(candidate.tag))
		}
	}
	return tags
}

func (s Segment) loadedCases(system *SegmentalSystem) *segmentCases {
	if system != nil {
		if cases := system.templateCases[s.Symbol]; cases != nil && cases.source == string(s.Content) {
			return cases
		}
	}
	return nil
}
