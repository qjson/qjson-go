package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/qjson/qjson-go/qjson"
)

func printHelp(w io.Writer) {
	fmt.Fprintf(w, "Usage: qjson <qjson file> | -v | -? | --help\n")
	fmt.Fprintf(w, "Print the qjson file content converted to JSON to stdout. "+
		"In  case of error, print an error message to stderr.\n")
	fmt.Fprintf(w, "  -v           outputs the version.\n")
	fmt.Fprintf(w, "  -?, --help   outputs this help message.\n")
	fmt.Fprintf(w, "\nReturn status is 0 when the convertion was successful, 1 otherwise\n")
}

func argsContain(args []string, val string) bool {
	for _, arg := range args {
		if val == arg {
			return true
		}
	}
	return false
}

func readFile(fileName string) ([]byte, error) {
	st, err := os.Stat(fileName)
	if err != nil {
		return nil, err
	}
	mode := st.Mode()
	if !mode.IsRegular() {
		return nil, fmt.Errorf("error: file '%s' is not a regular file", fileName)
	}
	text, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return text, nil
}

func readStdIn() ([]byte, error) {
	return ioutil.ReadAll(os.Stdin)
}

func main() {
	var qjsonText []byte
	var err error

	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "error: require a file name or an option as argument\n")
		printHelp(os.Stderr)
		os.Exit(1)
	}
	if argsContain(os.Args, "-?") || argsContain(os.Args, "--help") {
		printHelp(os.Stdout)
		os.Exit(0)
	}
	if argsContain(os.Args, "-v") {
		fmt.Println(qjson.Version())
		os.Exit(0)
	}

	if len(os.Args) == 2 {
		qjsonText, err = readFile(os.Args[1])
	} else {
		qjsonText, err = readStdIn()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	jsonText, err := qjson.Decode(qjsonText)
	if err != nil {
		fmt.Fprintf(os.Stderr, "qjson: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonText))
}
