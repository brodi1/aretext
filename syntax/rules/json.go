package rules

import (
	"strings"

	"github.com/aretext/aretext/syntax/parser"
)

var JsonRules []parser.TokenizerRule

func init() {
	JsonRules = []parser.TokenizerRule{
		{
			Regexp:    `true|false|null`,
			TokenRole: parser.TokenRoleKeyword,
		},
		{
			Regexp:    `-?[0-9]+(\.[0-9]+)?((e|E)-?[0-9]+)?`,
			TokenRole: parser.TokenRoleNumber,
		},
		{
			Regexp:    `"([^\"\n]|\\")*"`,
			TokenRole: parser.TokenRoleString,
			SubRules: []parser.TokenizerRule{
				{
					Regexp:    `^"`,
					TokenRole: parser.TokenRoleStringQuote,
				},
				{
					Regexp:    `"$`,
					TokenRole: parser.TokenRoleStringQuote,
				},
			},
		},
		{
			Regexp:    `"([^\"\n]|\\")*"[ \t]*:`,
			TokenRole: parser.TokenRoleKey,
		},
		{
			Regexp:    strings.Join([]string{`\{`, `\}`, `\[`, `\]`, `,`}, "|"),
			TokenRole: parser.TokenRolePunctuation,
		},

		// This prevents the number and keyword rules from matching substrings of a symbol.
		{
			Regexp:    `-?[a-zA-Z0-9_]+`,
			TokenRole: parser.TokenRoleNone,
		},
	}
}
