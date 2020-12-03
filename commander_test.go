package commander_test

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/matthewmueller/commander"
	"github.com/matthewmueller/diff"
	"github.com/pborman/ansi"
	"github.com/tj/assert"
)

// TODO: more tests
var tests = []struct {
	name string
	test func(t testing.TB)
}{
	{
		name: "help",
		test: func(t testing.TB) {
			actual := new(bytes.Buffer)
			cmd := commander.New("say", "same command").Writer(actual)
			err := cmd.Parse([]string{"-h"})
			assert.NoError(t, err)
			equal(t, actual.String(), `
				Usage:

					say [<flags>]

				Flags:

					-h, --help	Output usage information.
			`)
		},
	},
	{
		name: "invalid",
		test: func(t testing.TB) {
			actual := new(bytes.Buffer)
			cmd := commander.New("say", "same command").Writer(actual)
			err := cmd.Parse([]string{"blargle"})
			assert.EqualError(t, err, "unexpected blargle")
			equal(t, actual.String(), ``)
		},
	},
	{
		name: "simple",
		test: func(t testing.TB) {
			actual := new(bytes.Buffer)
			cmd := commander.New("say", "same command").Writer(actual)
			called := 0
			cmd.Run(func() error {
				called++
				return nil
			})
			err := cmd.Parse([]string{})
			assert.NoError(t, err)
			assert.Equal(t, 1, called)
			equal(t, actual.String(), ``)
		},
	},
	{
		name: "run error",
		test: func(t testing.TB) {
			actual := new(bytes.Buffer)
			cmd := commander.New("say", "same command").Writer(actual)
			called := 0
			cmd.Run(func() error {
				called++
				return errors.New("oh noz")
			})
			err := cmd.Parse([]string{})
			assert.EqualError(t, err, "oh noz")
			assert.Equal(t, 1, called)
			equal(t, actual.String(), ``)
		},
	},
	{
		name: "help with example",
		test: func(t testing.TB) {
			actual := new(bytes.Buffer)
			cmd := commander.New("say", "same command").Writer(actual)
			cmd.Example("say <something>", "say something")
			cmd.Example("say <something> [else]", "say something else")
			err := cmd.Parse([]string{"-h"})
			assert.NoError(t, err)
			equal(t, actual.String(), `
				Usage:

					say [<flags>]

				Flags:

					-h, --help	Output usage information.

				Examples:

					say something
					$ say <something>

					say something else
					$ say <something> [else]
			`)
		},
	},
	{
		name: "subcommand help with example",
		test: func(t testing.TB) {
			actual := new(bytes.Buffer)
			cmd := commander.New("say", "same command").Writer(actual)
			en := cmd.Command("en", "say in english")
			en.Example("say en <something>", "say something")
			en.Example("say en <something> [else]", "say something else")
			err := cmd.Parse([]string{"en", "-h"})
			assert.NoError(t, err)
			equal(t, actual.String(), `
				say in english

				Usage:

					say en

				Flags:

					-h, --help	Output usage information.

				Examples:

					say something
					$ say en <something>

					say something else
					$ say en <something> [else]
			`)
		},
	},
	{
		name: "before function",
		test: func(t testing.TB) {
			actual := new(bytes.Buffer)
			cmd := commander.New("say", "same command").Writer(actual)
			called := 0
			cmd.Before(func() error {
				called++
				return nil
			})
			en := cmd.Command("en", "say in english")
			en.Run(func() error {
				called++
				return nil
			})
			err := cmd.Parse([]string{"en"})
			assert.NoError(t, err)
			assert.Equal(t, 2, called)
			equal(t, actual.String(), ``)
		},
	},
}

func Test(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.test(t)
		})
	}
}

func equal(t testing.TB, actual, expected string) {
	is(t, trim(expected), strip(t, trim(actual)))
}

var whitespaceOnly = regexp.MustCompile("(?m)^[ \t]+$")
var leadingWhitespace = regexp.MustCompile("(?m)(^[ \t]*)(?:[^ \t\n])")

func trim(text string) string {
	var margin string
	text = whitespaceOnly.ReplaceAllString(text, "")
	indents := leadingWhitespace.FindAllStringSubmatch(text, -1)
	for i, indent := range indents {
		if i == 0 {
			margin = indent[1]
		} else if strings.HasPrefix(indent[1], margin) {
			continue
		} else if strings.HasPrefix(margin, indent[1]) {
			margin = indent[1]
		} else {
			margin = ""
			break
		}
	}
	if margin != "" {
		text = regexp.MustCompile("(?m)^"+margin).ReplaceAllString(text, "")
	}
	return strings.TrimSpace(strings.Replace(text, "  ", "	", -1))
}

func strip(t testing.TB, ansistr string) string {
	b, err := ansi.Strip([]byte(ansistr))
	assert.NoError(t, err)
	return string(b)
}

// is checks if expect and actual are equal
func is(t testing.TB, expect, actual string) {
	if expect == actual {
		return
	}
	var b bytes.Buffer
	b.WriteString("\n\x1b[4mExpect\x1b[0m:\n")
	b.WriteString(expect)
	b.WriteString("\n\n")
	b.WriteString("\x1b[4mActual\x1b[0m: \n")
	b.WriteString(actual)
	b.WriteString("\n\n")
	b.WriteString("\x1b[4mDifference\x1b[0m: \n")
	b.WriteString(diff.String(expect, actual))
	b.WriteString("\n")
	t.Fatal(b.String())
}
