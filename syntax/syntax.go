package syntax

import (
	"github.com/aretext/aretext/syntax/languages"
	"github.com/aretext/aretext/syntax/parser"
)

// Language is an enum of languages that we can parse.
type Language string

// AllLanguages lists every available language.
var AllLanguages []Language

const (
	LanguagePlaintext = Language("plaintext")
	LanguageJson      = Language("json")
	LanguageYaml      = Language("yaml")
	LanguageGo        = Language("go")
	LanguageGitCommit = Language("gitcommit")
	LanguageGitRebase = Language("gitrebase")
	LanguageDevlog    = Language("devlog")
)

// languageToParseFunc maps each language to its parse func.
var languageToParseFunc map[Language]parser.Func

func init() {
	languageToParseFunc = map[Language]parser.Func{
		LanguagePlaintext: nil,
		LanguageJson:      languages.JsonParseFunc(),
		LanguageYaml:      languages.YamlParseFunc(),
		LanguageGo:        languages.GolangParseFunc(),
		LanguageGitCommit: languages.GitCommitParseFunc(),
		LanguageGitRebase: languages.GitRebaseParseFunc(),
		LanguageDevlog:    languages.DevlogParseFunc(),
	}

	for language, _ := range languageToParseFunc {
		AllLanguages = append(AllLanguages, language)
	}
}

// ParseForLanguage creates a parser for a syntax language.
// If no parser is available (e.g. for LanguagePlaintext) this returns nil.
func ParserForLanguage(language Language) *parser.P {
	parseFunc := languageToParseFunc[language]
	if parseFunc == nil {
		return nil
	}
	return parser.New(parseFunc)
}
