package Lyger

func (l *lexicon) scanMap(raw []byte, category string, words map[string]lexeme) (lexeme, string) {
	if l.lookupIndex != nil {
		words = l.lookupIndex[category][string(raw)]
	}
	return checkLexemeMap(raw, words)
}

// Configured vocabulary is stable within expansion and tokenization. Index its
// surface forms for each stage, retaining category order and the normal matcher
// for candidate collisions. Restore direct lookup afterward: callers can change
// configured words between stages, including before Compilation.Make.
func (l *lexicon) indexLookups() func() {
	prior := l.lookupIndex
	l.lookupIndex = make(map[string]map[string]map[string]lexeme)
	for category, words := range map[string]map[string]lexeme{
		"Nouns": l.Nouns, "Verbs": l.Verbs, "Adjectives": l.Adjectives, "Adverbs": l.Adverbs,
		"Prepositions": l.Prepositions, "Articles": l.Articles, "Conjunctions": l.Conjunctions, "Evaluations": l.Evaluations,
	} {
		index := make(map[string]map[string]lexeme)
		for key, word := range words {
			forms := []string{word.Root}
			for fix := range word.Postfixes {
				forms = append(forms, processLexFix(word.Root, fix, "postfix"))
			}
			for fix := range word.Prefixes {
				forms = append(forms, processLexFix(word.Root, fix, "prefix"))
			}
			for _, form := range forms {
				if index[form] == nil {
					index[form] = make(map[string]lexeme)
				}
				index[form][key] = word
			}
		}
		l.lookupIndex[category] = index
	}
	return func() { l.lookupIndex = prior }
}
