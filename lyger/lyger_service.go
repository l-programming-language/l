package Lyger

import (
	"io/fs"

	distribution "github.com/l-programming-language/l"
)

// DefaultLanguage selects the English resources shipped with L.
const DefaultLanguage = "eng"

// Options selects a language and one complete compiler distribution. Its zero
// value uses English and embedded resources. AssetRoot records physical
// provenance; it never selects a resource provider by itself.
type Options struct {
	Language  string
	Resources fs.FS
	AssetRoot string
}

func (options Options) normalized() Options {
	if options.Language == "" {
		options.Language = DefaultLanguage
	}
	if options.Resources == nil {
		options.Resources = distribution.Files()
	}
	return options
}

// LygerService packages the functionality of the lyger L compiler for use in the main package
type LygerService struct {
	options Options
}

// NewLygerService loads a lyger service
func NewLygerService(options Options) *LygerService {
	return &LygerService{
		options: options.normalized(),
	}
}

// NewCompilation caches string literals, removes comments, recursively scans content for import statements and builds
func (ls *LygerService) NewCompilation(mainBytes []byte, mainPath string) *Compilation {
	comp := newCompilation(mainBytes, mainPath)
	comp.init(ls.options)
	if comp.Error != nil {
		return comp
	}
	// preprocess - cache texts, add file line number IDs, and remove comments for mainBytes
	comp.preProcess()
	if comp.Error != nil {
		return comp
	}
	// build compiledBytes via any defined import statements found in mainBytes
	comp.build()
	return comp
}
