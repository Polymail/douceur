package parser

import (
	"fmt"
	"polymail-api/lib/utils"
	"strings"
	"testing"

	"github.com/Polymail/douceur/css"
)

func MustParse(t *testing.T, txt string, nbRules int) *css.Stylesheet {
	stylesheet, err := Parse(txt)
	if err != nil {
		t.Fatal("Failed to parse css", err, txt)
	}

	if len(stylesheet.Rules) != nbRules {
		t.Fatal("Failed to parse Qualified Rules", txt)
	}

	return stylesheet
}

func MustEqualRule(t *testing.T, parsedRule *css.Rule, expectedRule *css.Rule) {
	if !parsedRule.Equal(expectedRule) {
		diff := parsedRule.Diff(expectedRule)

		t.Fatal(fmt.Sprintf("Rule parsing error\nExpected:\n\"%s\"\nGot:\n\"%s\"\nDiff:\n%s", expectedRule, parsedRule, strings.Join(diff, "\n")))
	}
}

func MustEqualCSS(t *testing.T, ruleString string, expected string) {
	if ruleString != expected {
		t.Fatal(fmt.Sprintf("CSS generation error\n   Expected:\n\"%s\"\n   Got:\n\"%s\"", expected, ruleString))
	}
}

func TestExtraSemicolons(t *testing.T) {
	MustParse(t, ".test { color: red;; background: yellow }", 1)
}

func TestQualifiedRule(t *testing.T) {
	input := `/* This is a comment */
p > a {
    color: blue;
    text-decoration: underline; /* This is a comment */
}`

	expectedRule := &css.Rule{
		Kind:      css.QualifiedRule,
		Prelude:   "p > a",
		Selectors: []string{"p > a"},
		Declarations: []*css.Declaration{
			{
				Property: "color",
				Value:    "blue",
			},
			{
				Property: "text-decoration",
				Value:    "underline",
			},
		},
	}

	expectedOutput := `p > a {
  color: blue;
  text-decoration: underline;
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestQualifiedRuleImportant(t *testing.T) {
	input := `/* This is a comment */
p > a {
    color: blue;
    text-decoration: underline !important;
    font-weight: normal   !IMPORTANT    ;
}`

	expectedRule := &css.Rule{
		Kind:      css.QualifiedRule,
		Prelude:   "p > a",
		Selectors: []string{"p > a"},
		Declarations: []*css.Declaration{
			{
				Property:  "color",
				Value:     "blue",
				Important: false,
			},
			{
				Property:  "text-decoration",
				Value:     "underline",
				Important: true,
			},
			{
				Property:  "font-weight",
				Value:     "normal",
				Important: true,
			},
		},
	}

	expectedOutput := `p > a {
  color: blue;
  text-decoration: underline !important;
  font-weight: normal !important;
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestQualifiedRuleSelectors(t *testing.T) {
	input := `table, tr, td {
  padding: 0;
}

body,
  h1,   h2,
    h3   {
  color: #fff;
}`

	expectedRule1 := &css.Rule{
		Kind:      css.QualifiedRule,
		Prelude:   "table, tr, td",
		Selectors: []string{"table", "tr", "td"},
		Declarations: []*css.Declaration{
			{
				Property: "padding",
				Value:    "0",
			},
		},
	}

	expectedRule2 := &css.Rule{
		Kind: css.QualifiedRule,
		Prelude: `body,
  h1,   h2,
    h3`,
		Selectors: []string{"body", "h1", "h2", "h3"},
		Declarations: []*css.Declaration{
			{
				Property: "color",
				Value:    "#fff",
			},
		},
	}

	expectedOutput := `table, tr, td {
  padding: 0;
}
body, h1, h2, h3 {
  color: #fff;
}`

	stylesheet := MustParse(t, input, 2)

	MustEqualRule(t, stylesheet.Rules[0], expectedRule1)
	MustEqualRule(t, stylesheet.Rules[1], expectedRule2)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestAtRuleCharset(t *testing.T) {
	input := `@charset "UTF-8";`

	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@charset",
		Prelude: "\"UTF-8\"",
	}

	expectedOutput := `@charset "UTF-8";`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestAtRuleCounterStyle(t *testing.T) {
	input := `@counter-style footnote {
  system: symbolic;
  symbols: '*' ⁑ † ‡;
  suffix: '';
}`

	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@counter-style",
		Prelude: "footnote",
		Declarations: []*css.Declaration{
			{
				Property: "system",
				Value:    "symbolic",
			},
			{
				Property: "symbols",
				Value:    "'*' ⁑ † ‡",
			},
			{
				Property: "suffix",
				Value:    "''",
			},
		},
	}

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), input)
}

func TestAtRuleDocument(t *testing.T) {
	input := `@document url(http://www.w3.org/),
               url-prefix(http://www.w3.org/Style/),
               domain(mozilla.org),
               regexp("https:.*")
{
  /* CSS rules here apply to:
     + The page "http://www.w3.org/".
     + Any page whose URL begins with "http://www.w3.org/Style/"
     + Any page whose URL's host is "mozilla.org" or ends with
       ".mozilla.org"
     + Any page whose URL starts with "https:" */

  /* make the above-mentioned pages really ugly */
  body { color: purple; background: yellow; }
}`

	expectedRule := &css.Rule{
		Kind: css.AtRule,
		Name: "@document",
		Prelude: `url(http://www.w3.org/),
               url-prefix(http://www.w3.org/Style/),
               domain(mozilla.org),
               regexp("https:.*")`,
		Rules: []*css.Rule{
			{
				Kind:      css.QualifiedRule,
				Prelude:   "body",
				Selectors: []string{"body"},
				Declarations: []*css.Declaration{
					{
						Property: "color",
						Value:    "purple",
					},
					{
						Property: "background",
						Value:    "yellow",
					},
				},
			},
		},
	}

	expectCSS := `@document url(http://www.w3.org/),
               url-prefix(http://www.w3.org/Style/),
               domain(mozilla.org),
               regexp("https:.*") {
  body {
    color: purple;
    background: yellow;
  }
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectCSS)
}

func TestAtRuleFontFace(t *testing.T) {
	input := `@font-face {
  font-family: MyHelvetica;
  src: local("Helvetica Neue Bold"),
       local("HelveticaNeue-Bold"),
       url(MgOpenModernaBold.ttf);
  font-weight: bold;
}`

	expectedRule := &css.Rule{
		Kind: css.AtRule,
		Name: "@font-face",
		Declarations: []*css.Declaration{
			{
				Property: "font-family",
				Value:    "MyHelvetica",
			},
			{
				Property: "src",
				Value: `local("Helvetica Neue Bold"),
       local("HelveticaNeue-Bold"),
       url(MgOpenModernaBold.ttf)`,
			},
			{
				Property: "font-weight",
				Value:    "bold",
			},
		},
	}

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), input)
}

func TestAtRuleFontFeatureValues(t *testing.T) {
	input := `@font-feature-values Font Two { /* How to activate nice-style in Font Two */
  @styleset {
    nice-style: 4;
  }
}`
	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@font-feature-values",
		Prelude: "Font Two",
		Rules: []*css.Rule{
			{
				Kind: css.AtRule,
				Name: "@styleset",
				Declarations: []*css.Declaration{
					{
						Property: "nice-style",
						Value:    "4",
					},
				},
			},
		},
	}

	expectedOutput := `@font-feature-values Font Two {
  @styleset {
    nice-style: 4;
  }
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestAtRuleImport(t *testing.T) {
	input := `@import "my-styles.css";
@import url('landscape.css') screen and (orientation:landscape);`

	expectedRule1 := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@import",
		Prelude: "\"my-styles.css\"",
	}

	expectedRule2 := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@import",
		Prelude: "url('landscape.css') screen and (orientation:landscape)",
	}

	stylesheet := MustParse(t, input, 2)

	MustEqualRule(t, stylesheet.Rules[0], expectedRule1)
	MustEqualRule(t, stylesheet.Rules[1], expectedRule2)

	MustEqualCSS(t, stylesheet.String(), input)
}

func TestAtRuleKeyframes(t *testing.T) {
	input := `@keyframes identifier {
  0% { top: 0; left: 0; }
  100% { top: 100px; left: 100%; }
}`
	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@keyframes",
		Prelude: "identifier",
		Rules: []*css.Rule{
			{
				Kind:      css.QualifiedRule,
				Prelude:   "0%",
				Selectors: []string{"0%"},
				Declarations: []*css.Declaration{
					{
						Property: "top",
						Value:    "0",
					},
					{
						Property: "left",
						Value:    "0",
					},
				},
			},
			{
				Kind:      css.QualifiedRule,
				Prelude:   "100%",
				Selectors: []string{"100%"},
				Declarations: []*css.Declaration{
					{
						Property: "top",
						Value:    "100px",
					},
					{
						Property: "left",
						Value:    "100%",
					},
				},
			},
		},
	}

	expectedOutput := `@keyframes identifier {
  0% {
    top: 0;
    left: 0;
  }
  100% {
    top: 100px;
    left: 100%;
  }
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestAtRuleMedia(t *testing.T) {
	input := `@media screen, print {
  body { line-height: 1.2 }
}`
	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@media",
		Prelude: "screen, print",
		Rules: []*css.Rule{
			{
				Kind:      css.QualifiedRule,
				Prelude:   "body",
				Selectors: []string{"body"},
				Declarations: []*css.Declaration{
					{
						Property: "line-height",
						Value:    "1.2",
					},
				},
			},
		},
	}

	expectedOutput := `@media screen, print {
  body {
    line-height: 1.2;
  }
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestAtRuleNamespace(t *testing.T) {
	input := `@namespace svg url(http://www.w3.org/2000/svg);`
	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@namespace",
		Prelude: "svg url(http://www.w3.org/2000/svg)",
	}

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), input)
}

func TestAtRulePage(t *testing.T) {
	input := `@page :left {
  margin-left: 4cm;
  margin-right: 3cm;
}`
	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@page",
		Prelude: ":left",
		Declarations: []*css.Declaration{
			{
				Property: "margin-left",
				Value:    "4cm",
			},
			{
				Property: "margin-right",
				Value:    "3cm",
			},
		},
	}

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), input)
}

func TestAtRuleSupports(t *testing.T) {
	input := `@supports (animation-name: test) {
    /* specific CSS applied when animations are supported unprefixed */
    @keyframes { /* @supports being a CSS conditional group at-rule, it can includes other relevent at-rules */
      0% { top: 0; left: 0; }
      100% { top: 100px; left: 100%; }
    }
}`
	expectedRule := &css.Rule{
		Kind:    css.AtRule,
		Name:    "@supports",
		Prelude: "(animation-name: test)",
		Rules: []*css.Rule{
			{
				Kind: css.AtRule,
				Name: "@keyframes",
				Rules: []*css.Rule{
					{
						Kind:      css.QualifiedRule,
						Prelude:   "0%",
						Selectors: []string{"0%"},
						Declarations: []*css.Declaration{
							{
								Property: "top",
								Value:    "0",
							},
							{
								Property: "left",
								Value:    "0",
							},
						},
					},
					{
						Kind:      css.QualifiedRule,
						Prelude:   "100%",
						Selectors: []string{"100%"},
						Declarations: []*css.Declaration{
							{
								Property: "top",
								Value:    "100px",
							},
							{
								Property: "left",
								Value:    "100%",
							},
						},
					},
				},
			},
		},
	}

	expectedOutput := `@supports (animation-name: test) {
  @keyframes {
    0% {
      top: 0;
      left: 0;
    }
    100% {
      top: 100px;
      left: 100%;
    }
  }
}`

	stylesheet := MustParse(t, input, 1)
	rule := stylesheet.Rules[0]

	MustEqualRule(t, rule, expectedRule)

	MustEqualCSS(t, stylesheet.String(), expectedOutput)
}

func TestParseDeclarations(t *testing.T) {
	testcases := map[string]struct {
		input    string
		expected []*css.Declaration
	}{
		"basic:": {
			input: `color: blue; text-decoration:underline;`,
			expected: []*css.Declaration{
				{
					Property: "color",
					Value:    "blue",
				},
				{
					Property: "text-decoration",
					Value:    "underline",
				},
			},
		},
		"better handling inline styles: see HACK parser.go:line58": {
			input: `font-family:-apple-system,-webkit-system-font,SFNSText,Segoe UI,system,HelveticaNeue,Helvetica,Arial,sans-serif;font-weight:400;color:#333333;font-size:17px;line-height:26px;-webkit-font-smoothing:antialiased`,
			expected: []*css.Declaration{
				{
					Property:  "font-family",
					Value:     "-apple-system,-webkit-system-font,SFNSText,Segoe UI,system,HelveticaNeue,Helvetica,Arial,sans-serif",
					Important: false,
				},
				{
					Property:  "font-weight",
					Value:     "400",
					Important: false,
				},
				{
					Property:  "color",
					Value:     "#333333",
					Important: false,
				},
				{
					Property:  "font-size",
					Value:     "17px",
					Important: false,
				},
				{
					Property:  "line-height",
					Value:     "26px",
					Important: false,
				},
				{
					Property:  "-webkit-font-smoothing",
					Value:     "antialiased",
					Important: false,
				},
			},
		},
	}
	for _, tc := range testcases {
		declarations, err := ParseDeclarations(tc.input)
		if err != nil {
			t.Fatal("Failed to parse Declarations:", tc.input)
		}
		if len(declarations) != len(tc.expected) {
			utils.PPrintln(declarations)
			t.Fatal("Failed to parse Declarations:", tc.input)
		}
		for i, decl := range declarations {
			if !decl.Equal(tc.expected[i]) {
				t.Fatal("Failed to parse Declarations: ", decl.String(), tc.expected[i].String())
			}
		}
	}
}

func TestParseInvalidCSS(t *testing.T) {
	testcases := map[string]struct {
		input         string
		expected      string
		expectedError string
	}{
		"should ignore extraneous semicolons": {
			input:    "div { background-color: yellow };;;;;",
			expected: "div {\n  background-color: yellow;\n}",
		},
		"should ignore selectors with no declarations": {
			input:    "p; div > p; div { background-color: yellow };",
			expected: "div {\n  background-color: yellow;\n}",
		},
	}

	for msg, tc := range testcases {
		ss, err := Parse(tc.input)
		if tc.expectedError == "" {
			if err != nil {
				t.Fatalf("%v: Parse error: %v", msg, err.Error())
			}
			if tc.expected != ss.String() {
				t.Fatalf("%v: Expected output: %v, Got: %v", msg, tc.expected, ss.String())
			}
		} else {
			if err == nil {
				t.Fatalf("%v: Expected parse error: %v", msg, tc.expectedError)
			}
			if tc.expectedError != err.Error() {
				t.Fatalf("%v: Expected error: %v, Got: %v", msg, tc.expectedError, err.Error())
			}
		}
	}
}
