package Lyger

// A Function value is callable under a new binding only when that binding
// explicitly declares its invocation family. Ordinary value labels stay data.
func (r *record) declareInvocation(label string) error {
	if !r.isFunction() {
		return nil
	}
	word := newLexeme("verb")
	if err := word.reLabel(label); err != nil {
		return err
	}
	r.word = nil
	if len(word.Forms) > 0 {
		r.word = word
	}
	return nil
}

func (r *record) hasInvocationFamily() bool {
	word := r.bindingWord()
	return word != nil && word.POS == "verb" && word.validateFamily() == nil
}
