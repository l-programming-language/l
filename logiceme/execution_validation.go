package Logiceme

// glyphForm uses parsed symbols, before Compile assigns the rendered category.
func glyphForm(g *Glyph) string {
	if g == nil {
		return ""
	}
	form := g.Determinant.Symbol
	for _, vowel := range g.Vowels {
		form += vowel.Symbol
	}
	return form + g.Consonant.Symbol
}

func glyphHasVowel(g *Glyph, symbol string) bool {
	for _, vowel := range g.Vowels {
		if vowel.Symbol == symbol {
			return true
		}
	}
	return false
}
