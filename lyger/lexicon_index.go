package Lyger

// Configured vocabulary is stable within expansion and tokenization. Index its
// surface forms for each stage, retaining category order and the normal matcher
// for candidate collisions. Restore direct lookup afterward: callers can change
// configured words between stages, including before Compilation.Make.
func (l *lexicon) indexLookups() func() {
	prior := l.lookupIndex
	l.lookupIndex = make(map[string]map[string]map[string]lexeme)
	for _, category := range l.categories() {
		index := make(map[string]map[string]lexeme)
		for key, word := range *category.words {
			for _, form := range word.forms() {
				if index[form.Spelling] == nil {
					index[form.Spelling] = make(map[string]lexeme)
				}
				index[form.Spelling][key] = word
			}
		}
		l.lookupIndex[category.name] = index
	}
	return func() { l.lookupIndex = prior }
}
