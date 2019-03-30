package commander_test

import (
	"bytes"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/matthewmueller/diff"
	"github.com/pborman/ansi"
	"github.com/tj/assert"
)

// TODO: more tests
var tests = []struct {
	name   string
	input  string
	stderr string
	stdout string
	err    string
	main   string
}{
	{
		name:  "help",
		input: "-h",
		stderr: `
			Usage:

				say [<flags>]

			Flags:

				-h, --help  Output usage information.
		`,
		main: `
			func main() {
				cmd := commander.New("say", "same command")
				cmd.MustParse(os.Args[1:])
			}
		`,
	},
	{
		name:  "invalid",
		input: "blargle",
		stderr: `
				say: error: unexpected blargle

				Usage:

					say [<flags>]

				Flags:

					-h, --help	Output usage information.

			exit status 1
		`,
		err: `exit status 1`,
		main: `
			func main() {
				cmd := commander.New("say", "same command")
				cmd.MustParse(os.Args[1:])
			}
		`,
	},
	{
		name:   "simple",
		input:  "",
		stdout: `hi!`,
		main: `
			func main() {
				cmd := commander.New("say", "same command")
				cmd.Run(func() error {
					fmt.Printf("hi!")
					return nil
				})
				cmd.MustParse(os.Args[1:])
			}
		`,
	},
	{
		name:   "subcommand",
		input:  "english",
		stdout: `hi!`,
		main: `
			func main() {
				cmd := commander.New("say", "same command")
				sub := cmd.Command("english", "say in english")
				sub.Run(func() error {
					fmt.Printf("hi!")
					return nil
				})
				cmd.MustParse(os.Args[1:])
			}
		`,
	},
	{
		name:  "help with example",
		input: "-h",
		stderr: `
			Usage:

				say [<flags>]

			Flags:

				-h, --help  Output usage information.

			Examples:

				say something
				$ say <something>

				say something else
				$ say <something> [else]
		`,
		main: `
			func main() {
				cmd := commander.New("say", "same command")
				cmd.Example("say <something>", "say something")
				cmd.Example("say <something> [else]", "say something else")
				cmd.MustParse(os.Args[1:])
			}
		`,
	},
	{
		name:  "subcommand help with example",
		input: "en -h",
		stderr: `
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
		`,
		main: `
			func main() {
				cmd := commander.New("say", "same command")
				en := cmd.Command("en", "say in english")
				en.Example("say en <something>", "say something")
				en.Example("say en <something> [else]", "say something else")
				cmd.MustParse(os.Args[1:])
			}
		`,
	},
}

func Test(t *testing.T) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pkg := pkgpath(t)
			mainfile := filepath.Join("tmp", "main.go")
			var args []string
			if test.input != "" {
				args = strings.Split(test.input, " ")
			}
			stdout, stderr, err, cleanup := gorun(t, mainfile, code(pkg, test.main), args...)
			is(t, trim(test.stderr), strip(t, trim(stderr)))
			is(t, trim(test.stdout), strip(t, trim(stdout)))
			if err != nil {
				is(t, trim(test.err), strip(t, trim(err.Error())))
			}
			if !t.Failed() {
				cleanup()
			}
		})
	}
}

func code(imp, main string) string {
	return fmt.Sprintf(`
package main
import commander %q
%s
`, imp, main)
}

func pkgpath(t testing.TB) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to get file")
	}
	gosrc := path.Join(build.Default.GOPATH, "src")
	rel, err := filepath.Rel(gosrc, filepath.Dir(file))
	assert.NoError(t, err)
	return rel
}

// gorun write a main.go file runs it, returning the result
func gorun(t testing.TB, path, main string, args ...string) (string, string, *exec.ExitError, func()) {
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0755)
	assert.NoError(t, err)

	code := imports(t, main)
	assert.NoError(t, err)

	err = ioutil.WriteFile(path, []byte(code), 0644)
	assert.NoError(t, err)

	gobin, err := exec.LookPath("go")
	assert.NoError(t, err)

	// add the arguments
	var arguments []string
	arguments = append(arguments, "run", path)
	arguments = append(arguments, args...)

	// go run
	cmd := exec.Command(gobin, arguments...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// run the command
	var eerr *exec.ExitError
	if err := cmd.Run(); err != nil {
		eerr = err.(*exec.ExitError)
	}

	return stdout.String(), stderr.String(), eerr, func() {
		assert.NoError(t, os.RemoveAll(dir))
	}
}

// Format the output using goimports
func imports(t testing.TB, input string) (output string) {
	goimports, err := exec.LookPath("goimports")
	assert.NoError(t, err)

	cmd := exec.Command(goimports)
	stdin, err := cmd.StdinPipe()
	assert.NoError(t, err)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	reader := bytes.NewBufferString(input)
	err = cmd.Start()
	assert.NoError(t, err)

	_, err = io.Copy(stdin, reader)
	assert.NoError(t, err)
	assert.NoError(t, stdin.Close())

	if err := cmd.Wait(); err != nil {
		t.Fatal(stderr.String())
	}
	return stdout.String()
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
	b.WriteString("\n\x1b[4mexpect\x1b[0m:\n")
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
