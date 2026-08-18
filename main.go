package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const usage = `delta-scan - Narrows linter reports to what has actually changed.

Usage:
    delta-scan -r report_file [OPTIONS...]

The report file must be in SARIF json format.

Examples:
  Read diff from stdin and output the rules violations in the terminal
    git diff feature..main | delta-scan -r sast-report.sarif

  Read diff from a file
    delta-scan -d change.diff -r sast-report.sarif

  Export the matched violations to a new SARIF file
    delta-scan -d change.diff -r sast-report.sarif -x -f outfile.sarif

Return values:
  0 - no violation found
  1 - one or more violation found
  2 - error during command execution

Options:
  -d, --diff file
                Diff file location. Use '-' to read from
                stdin (defaults to stdin)
  -r, --report file
                SARIF report file location (required)
  -x, --export  Export matched violations to a new SARIF file
  -f, --file file
                Output SARIF filename (required with -x)
  -h, --help    Display this message
`

type options struct {
	diff   string
	report string
	export bool
	file   string
	help   bool
}

func parseOptions(args []string) (options, error) {
	var opts options

	flags := flag.NewFlagSet("delta-scan", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&opts.diff, "d", "", "")
	flags.StringVar(&opts.diff, "diff", "", "")
	flags.StringVar(&opts.report, "r", "", "")
	flags.StringVar(&opts.report, "report", "", "")
	flags.BoolVar(&opts.export, "x", false, "")
	flags.BoolVar(&opts.export, "export", false, "")
	flags.StringVar(&opts.file, "f", "", "")
	flags.StringVar(&opts.file, "file", "", "")
	flags.BoolVar(&opts.help, "h", false, "")
	flags.BoolVar(&opts.help, "help", false, "")
	flags.Usage = func() {}

	if err := flags.Parse(args); err != nil {
		return opts, err
	}

	if opts.help {
		return opts, nil
	}

	if len(flags.Args()) != 0 {
		return opts, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if opts.report == "" {
		return opts, errors.New("no report file provided (use -r)")
	}
	if opts.export && opts.file == "" {
		return opts, errors.New("-f is required when -x is set")
	}
	if opts.file != "" && !opts.export {
		return opts, errors.New("-f can only be used with -x")
	}

	return opts, nil
}

func main() {
	os.Exit(run(os.Stdin, os.Stdout, os.Stderr))
}

func run(r io.Reader, wout, werr io.Writer) int {
	opts, err := parseOptions(os.Args[1:])
	if err != nil {
		fmt.Fprintf(werr, "%s\n", usage)
		fmt.Fprintf(werr, "error: %s\n", err)
		return 2
	}

	if opts.help {
		fmt.Fprintf(wout, "%s\n", usage)
		return 0
	}

	var content contentChecker
	switch opts.diff {
	case "", "-":
		content, err = readDiff(r)
	default:
		content, err = readDiffFile(opts.diff)
	}
	if err != nil {
		fmt.Fprintf(werr, "error: could not read diff: %s\n", err)
		return 2
	}

	report, err := readPMDFile(opts.report)
	if err != nil {
		fmt.Fprintf(werr, "error: could not read report: %s\n", err)
		return 2
	}

	if opts.export {
		if err := writeSARIFFile(opts.file, report.filter(content)); err != nil {
			fmt.Fprintf(werr, "error: could not write sarif: %s\n", err)
			return 2
		}
	}

	reporter := newTextReporter(wout)
	cmd := command{
		content:  content,
		reporter: reporter,
		linter:   report,
	}

	violationCount, err := cmd.run()
	if err != nil {
		fmt.Fprintf(werr, "error: command failed: %s\n", err)
		return 2
	}

	if violationCount > 0 {
		fmt.Fprintf(werr, "%d violations found.\n", violationCount)
		return 1
	}
	fmt.Print("No violations found.")
	return 0
}
