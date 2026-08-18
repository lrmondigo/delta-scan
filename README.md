# delta-scan - Narrows analysis reports to what has actually changed.

``` shell
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
```
