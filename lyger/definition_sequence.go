package Lyger

func definitionSteps(instance DefinitionInstance) []DefinitionInstance {
	var steps []DefinitionInstance
	for _, prior := range instance.before {
		steps = append(steps, definitionSteps(prior)...)
	}
	instance.before = nil
	if instance.result != nil && instance.Expression == "" && instance.Words == nil {
		return steps
	}
	return append(steps, instance)
}
