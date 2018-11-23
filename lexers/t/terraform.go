package t

import (
	. "github.com/alecthomas/chroma" // nolint
	"github.com/alecthomas/chroma/lexers/internal"
)

// Terraform lexer.
var Terraform = internal.Register(MustNewLexer(
	&Config{
		Name:      "Terraform",
		Aliases:   []string{"terraform", "tf"},
		Filenames: []string{"*.tf"},
		MimeTypes: []string{"application/x-tf", "application/x-terraform"},
	},
	Rules{
		"root": {
			Include("general"),
			{`(resource|data)(\s*)`, ByGroups(KeywordDeclaration, Text), Push("resourceBlock")},
			{`(module)(\s*)`, ByGroups(KeywordDeclaration, Text), Push("moduleBlock")},
			{`(variable)(\s*)`, ByGroups(KeywordDeclaration, Text), Push("variableBlock")},
			{`(output)(\s*)`, ByGroups(KeywordDeclaration, Text), Push("outputBlock")},
			{`(provider)(\s*)`, ByGroups(KeywordDeclaration, Text), Push("providerBlock")},
			{`(terraform)(\s*)`, ByGroups(KeywordDeclaration, Text), Push("terraformBlock")},
			Include("bodyContentLocals"),

			// We'll just match everything else as normal body content, even though
			// the top-level is much more restrictive, since that'll give some
			// reasonable formatting for any future Terraform features we didn't
			// account for here.
			Include("bodyContent"),
		},
		"general": {
			{`\s+`, Text, nil},
			{`(//|#.*)\n`, CommentSingle, nil},
			{`(/\*.*\*/)\n`, CommentMultiline, nil},
		},
		"label": {
			Include("general"),
			{`([-\w]+|\"(?:[^"]|\\")*\")(\s*)`, ByGroups(LiteralStringName, Text), nil},
		},
		"bodyContent": {
			Include("general"),

			// Since arguments don't have an ending marker of their
			// own (it's implied by the end of the line or enclosing block) we
			// don't have a state specifically for them but rather just match
			// expression-like content throughout the body. However, we do still
			// look for what appear to be argument definitions.
			{`([-\w]+)(\s*)(=)(\s*)`, ByGroups(NameAttribute, Text, Punctuation, Text), nil},
			{`([-\w]+)(\s*)`, ByGroups(NameAttribute, Text), Push("nestedBlock")},

			Include("expr"),
		},
		"bodyContentLifecycle": {
			{`(lifecycle)(\s*)`, ByGroups(Keyword, Text), Push("lifecycleBlock")},
		},
		"bodyContentLocals": {
			{`(locals)(\s*)`, ByGroups(Keyword, Text), Push("localsBlock")},
		},
		"expr": {
			{`"`, LiteralStringDelimiter, Push("templateQuoted")},
			{`<<-?(\w+)\n.*\n\1\n`, LiteralStringHeredoc, nil},
			{`(\w+)(\()`, ByGroups(NameFunction, Punctuation), nil},
			{`(\.)(\w+)`, ByGroups(Punctuation, NameAttribute), nil},
			{`\w+`, NameVariable, nil},
			{`\{`, Punctuation, Push("exprBrace")},
			{`[\[\]()},.]`, Punctuation, nil},
			{`(\+|\-|\*|\\|%|&&|\|\||!|==|<=?|>=?)`, Operator, nil},
		},
		"typeExpr": {
			{`(string|number|bool|list|set|map|tuple|object)`, KeywordType, nil},
			Include("expr"),
		},
		"exprBrace": {
			{`\}`, Punctuation, Pop(1)},
			Include("expr"),
		},
		"templateQuoted": {
			{`"`, LiteralStringDelimiter, Pop(1)},
			{`$\{`, LiteralStringInterpol, Push("templateInterp")},
			{`%\{`, LiteralStringInterpol, Push("templateControl")},
			{`[^"]+`, LiteralString, nil},
		},
		"templateInterp": {
			{`\}`, Punctuation, Pop(1)},
			Include("expr"),
		},
		"templateControl": {
			{`\}`, Punctuation, Pop(1)},
			Include("expr"),
		},

		// Block types have two rules each: one for the block itself, which
		// matches the labels and the opening brace, and then one for the body
		// which matches all of the content up to the closing brace, often
		// including recognizing certain special items inside that Terraform
		// itself defines, as opposed to plugins defining them.
		"resourceBlock": {
			Include("label"),
			{`\{`, Punctuation, Push("resourceBody")},
		},
		"resourceBody": {
			{`(count|for_each|depends_on)(\s*)(=)(\s*)`, ByGroups(Keyword, Text, Punctuation, Text), nil},
			{`(provisioner)(\s*)`, ByGroups(KeywordDeclaration, Text), Push("provisionerBlock")},
			Include("bodyContentLifecycle"),
			Include("bodyContentLocals"),
			Include("bodyContent"),
			{`\}`, Punctuation, Pop(2)}, // pop out of both resourceBody and resourceBlock
		},
		"moduleBlock": {
			Include("label"),
			{`\{`, Punctuation, Push("moduleBody")},
		},
		"moduleBody": {
			{`(source|version|count|for_each|depends_on)(\s*)(=)(\s*)`, ByGroups(Keyword, Text, Punctuation, Text), nil},
			Include("bodyContentLifecycle"),
			Include("bodyContentLocals"),
			Include("bodyContent"),
			{`\}`, Punctuation, Pop(2)}, // pop out of both moduleBody and moduleBlock
		},
		"variableBlock": {
			Include("label"),
			{`\{`, Punctuation, Push("variableBody")},
		},
		"variableBody": {
			{`(type|default|description|sensitive)(\s*)(=)(\s*)`, ByGroups(Keyword, Text, Punctuation, Text), nil},
			Include("typeExpr"),
			Include("bodyContent"),
			{`\}`, Punctuation, Pop(2)}, // pop out of both variableBody and variableBlock
		},
		"outputBlock": {
			Include("label"),
			{`\{`, Punctuation, Push("outputBody")},
		},
		"outputBody": {
			{`(value|description|sensitive|depends_on)(\s*)(=)(\s*)`, ByGroups(Keyword, Text, Punctuation, Text), nil},
			Include("bodyContent"),
			{`\}`, Punctuation, Pop(2)}, // pop out of both outputBody and outputBlock
		},
		"providerBlock": {
			Include("label"),
			{`\{`, Punctuation, Push("providerBody")},
		},
		"providerBody": {
			{`(source|version|alias)(\s*)(=)(\s*)`, ByGroups(Keyword, Text, Punctuation, Text), nil},
			Include("bodyContentLifecycle"),
			Include("bodyContent"),
			{`\}`, Punctuation, Pop(2)}, // pop out of both providerBody and providerBlock
		},
		"localsBlock": {
			Include("label"),
			{`\{`, Punctuation, Push("localsBody")},
		},
		"localsBody": {
			Include("bodyContent"),
			{`\}`, Punctuation, Pop(2)}, // pop out of both localsBody and localsBlock
		},
		"terraformBlock": {
			Include("label"),
			{`\{`, Punctuation, Push("terraformBody")},
		},
		"terraformBody": {
			{`(required_version|required_providers)(\s*)(=)(\s*)`, ByGroups(Keyword, Text, Punctuation, Text), nil},
			{`(backend)(\s*)`, ByGroups(KeywordDeclaration, Text), Push("backendBlock")},
			Include("bodyContent"),
			{`\}`, Punctuation, Pop(2)}, // pop out of both terraformBody and terraformBlock
		},
		"backendBlock": {
			Include("label"),
			{`\{`, Punctuation, Push("backendBody")},
		},
		"backendBody": {
			Include("bodyContent"),
			{`\}`, Punctuation, Pop(2)}, // pop out of both backendBody and backendBlock
		},
		"lifecycleBlock": {
			Include("label"),
			{`\{`, Punctuation, Push("lifecycleBody")},
		},
		"lifecycleBody": {
			Include("bodyContent"),
			{`\}`, Punctuation, Pop(2)}, // pop out of both lifecycleBody and lifecycleBlock
		},
		"provisionerBlock": {
			Include("label"),
			{`\{`, Punctuation, Push("provisionerBody")},
		},
		"provisionerBody": {
			{`(when|on_failure)(\s*)(=)(\s*)`, ByGroups(Keyword, Text, Punctuation, Text), nil},
			Include("bodyContent"),
			{`\}`, Punctuation, Pop(2)}, // pop out of both provisionerBody and provisionerBlock
		},
		"connectionBlock": {
			Include("label"),
			{`\{`, Punctuation, Push("connectionBody")},
		},
		"connectionBody": {
			Include("bodyContent"),
			{`\}`, Punctuation, Pop(2)}, // pop out of both connectionBody and connectionBlock
		},
		"nestedBlock": { // nestedBlock is a plugin-defined block within one of the language's toplevels
			Include("label"),
			{`\{`, Punctuation, Push("nestedBody")},
		},
		"nestedBody": {
			Include("bodyContent"),
			{`\}`, Punctuation, Pop(2)}, // pop out of both nestedBody and nestedBlock
		},
	},
))
