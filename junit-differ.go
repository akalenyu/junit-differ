package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2/reporters"
	flag "github.com/spf13/pflag"
)

func main() {

	if path := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); path != "" {
		if err := os.Chdir(path); err != nil {
			panic(err)
		}
	}

	var output string
	flag.StringVarP(&output, "output", "o", "-", "File to write the resulting junit file to, defaults to stdout (-)")
	flag.Parse()
	junitFiles := flag.Args()

	if len(junitFiles) != 2 {
		log.Panicln("Diffing junit reports requires exactly 2 junit files")
	}

	suites, err := loadJUnitFiles(junitFiles)
	if err != nil {
		log.Panicln("Could not load JUnit files.")
	}

	result, err := diffJUnitFiles(suites)
	if err != nil {
		log.Panicln("Could not diff JUnit files")
	}

	writer, err := prepareOutput(output)
	if err != nil {
		log.Panicln("Failed to prepare the output file")
	}

	err = writeJunitFile(writer, result)
	if err != nil {
		log.Panicln("Failed to write the diffed junit report")
	}
}

func loadJUnitFiles(fileGlobs []string) (suites []reporters.JUnitTestSuite, err error) {
	for _, fileglob := range fileGlobs {
		files, err := filepath.Glob(fileglob)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			f, err := os.Open(file)
			if err != nil {
				return nil, fmt.Errorf("failed to open file %s: %v", file, err)
			}
			suite := reporters.JUnitTestSuite{}
			err = xml.NewDecoder(f).Decode(&suite)
			if err != nil {
				return nil, fmt.Errorf("failed to decode suite %s: %v", file, err)
			}
			suites = append(suites, suite)
		}
	}
	return suites, nil
}

type DeprecatedJUnitTestSuite struct {
	XMLName   xml.Name                  `xml:"testsuite"`
	TestCases []reporters.JUnitTestCase `xml:"testcase"`
	Name      string                    `xml:"name,attr"`
	Tests     int                       `xml:"tests,attr"`
	Failures  int                       `xml:"failures,attr"`
	Errors    int                       `xml:"errors,attr"`
	Time      float64                   `xml:"time,attr"`
}

func diffJUnitFiles(suites []reporters.JUnitTestSuite) (result *DeprecatedJUnitTestSuite, err error) {
	result = &DeprecatedJUnitTestSuite{}

	a := suites[0].TestCases
	b := suites[1].TestCases
	m := make(map[string]bool)
	for _, tc := range b {
		m[tc.Name] = tc.Skipped != nil
	}
	for _, tc := range a {
		skipped := tc.Skipped != nil
		if val, ok := m[tc.Name]; !ok || skipped != val {
			result.TestCases = append(result.TestCases, tc)
			result.Tests += 1
		}
	}

	result.Name = "Diffed Test Suite"
	for _, suite := range suites {
		result.Time += suite.Time
		result.Failures += suite.Failures
		result.Errors += suite.Errors
	}

	return result, nil
}

func prepareOutput(output string) (writer io.Writer, err error) {
	writer = os.Stdout
	if output != "-" && output != "" {
		writer, err = os.Create(output)
		if err != nil {
			return nil, err
		}
	}
	return writer, nil
}

func writeJunitFile(writer io.Writer, suite *DeprecatedJUnitTestSuite) error {
	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")
	err := encoder.Encode(suite)
	if err != nil {
		return err
	}
	return nil
}
