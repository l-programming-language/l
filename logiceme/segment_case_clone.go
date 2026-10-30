package Logiceme

// Case parsing shares one TagMeta through its expression tree. Preserve that
// relationship within each copy while giving every selected case its own data.
type segmentTagCopy map[*TagMeta]*TagMeta

func (copying segmentTagCopy) metadata(original *TagMeta) *TagMeta {
	if original == nil {
		return nil
	}
	if found := copying[original]; found != nil {
		return found
	}
	result := *original
	result.AppMeta = original.AppMeta.clone()
	copying[original] = &result
	return &result
}

func (copying segmentTagCopy) expressions(original []*TagExpression) []*TagExpression {
	if original == nil {
		return nil
	}
	result := make([]*TagExpression, len(original))
	for index, original := range original {
		expression := *original
		expression.TagMeta = copying.metadata(original.TagMeta)
		if original.ExpressionSubTags != nil {
			expression.ExpressionSubTags = make([]*SubTag, len(original.ExpressionSubTags))
			for index, original := range original.ExpressionSubTags {
				subtag := *original
				subtag.TagMeta = copying.metadata(original.TagMeta)
				subtag.Expressions = copying.expressions(original.Expressions)
				expression.ExpressionSubTags[index] = &subtag
			}
		}
		result[index] = &expression
	}
	return result
}

func cloneSegmentTag(original *LucidTag) *LucidTag {
	copying := make(segmentTagCopy)
	result := *original
	result.TagMeta = copying.metadata(original.TagMeta)
	result.Expressions = copying.expressions(original.Expressions)
	return &result
}
