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
	result.AppMeta.Index.Keys = append(original.AppMeta.Index.Keys[:0:0], original.AppMeta.Index.Keys...)
	result.AppMeta.Index.Indexes = append(original.AppMeta.Index.Indexes[:0:0], original.AppMeta.Index.Indexes...)
	if original.AppMeta.Metas != nil {
		result.AppMeta.Metas = make([]*MetaObj, len(original.AppMeta.Metas))
		for index, original := range original.AppMeta.Metas {
			if original != nil {
				entry := *original
				result.AppMeta.Metas[index] = &entry
			}
		}
	}
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
