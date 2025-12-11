package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nats-io/nats.go"
	"gopkg.in/yaml.v2"
)

// var nc *nats.Conn

func main() {

	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	_, err = nc.QueueSubscribe("worker.*.grade", "worker.grade", handler)
	if err != nil {
		log.Fatal(err)
	}

	//pull exercices from DB

	log.Println("Worker is running and listening...")

	select {}
}

type Result struct {
	Success bool         `json:"success"`
	Tests   []TestResult `json:"tests"`
}

type TestResult struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Logs    string `json:"logs"`
}

type Spec struct {
	Id         string      `yaml:"id"`
	Type       string      `yaml:"type"`
	Build      *Build      `yaml:"build"`
	Tests      []Test      `yaml:"tests"`
	Validation *Validation `yaml:"validation"`
}

type Build struct {
	Cmd     string `yaml:"command"`
	Timeout string `yaml:"timeout"`
}

type Test struct {
	Name     string `yaml:"name"`
	Run      string `yaml:"run"`
	Input    string `yaml:"input"`
	Expected string `yaml:"expected_output"`
	Timeout  int    `yaml:"timeout"`
}

type Validation struct {
	Valgrind bool     `yaml:"check_valgrind"`
	Allowed  []string `yaml:"allowed_functions"`
}

func handler(m *nats.Msg) {

	var inputs []string

	subject := strings.Split(m.Subject, ".")
	exerciseId := subject[1]

	//parse test

	spec, err := searchExerciseTest(exerciseId)
	if err != nil {
		log.Println("Search exercise Error:", err)
		m.Respond([]byte(`{"success": false, "error": "Invalid input format"}`))
		return
	}

	err = json.Unmarshal(m.Data, &inputs)
	if err != nil {
		log.Println("Unmarshal JSON Error:", err)
		m.Respond([]byte(`{"success": false, "error": "Invalid input format"}`))
		return
	}

	//test
	testResults, err := testInput(*spec, inputs)
	if err != nil {
		log.Printf("Error test input: %v", err)
	}

	result := prepareResponse(testResults)
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Printf("Marshal Error for response: %v", err)
		m.Respond([]byte(`{"success": false, "error": "Internal server error during JSON encoding"}`))
		return
	}

	m.Respond(resultJSON)
	// if err != nil {
	// 	log.Printf("Error sending response to NATS: %v", err)
	// }
}

func searchExerciseTest(exerciseId string) (*Spec, error) {

	var foundSpec *Spec = nil
	err := filepath.Walk("app/exercises", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}

			var spec Spec

			if unmarshalErr := yaml.Unmarshal(data, &spec); unmarshalErr == nil {
				fmt.Printf("YAML unmarshal error in %s: %v", path, unmarshalErr)
				return nil
			}

			if spec.Id == exerciseId {
				foundSpec = &spec
				return fmt.Errorf("spec found")
			}
		}
		return nil
	})

	if err != nil && err.Error() == "spec found" {
		return foundSpec, nil
	}

	if err != nil {
		return nil, fmt.Errorf("File parsing error: %w", err)
	}

	return nil, fmt.Errorf("Can't find exercise with id: %s", exerciseId)
}

func testInput(spec Spec, inputs []string) ([]TestResult, error) {

	switch {
	case spec.Type == "program":
		return programHandler(inputs, spec)
	case spec.Type == "function":
		return funcHandler(inputs, spec)
	case spec.Type == "text":
		return textHandler(inputs, spec)
	case spec.Type == "mcq":
		return mcqHandler(inputs, spec)
	}
	var tab []TestResult
	return tab, nil
}

func programHandler(inputs []string, spec Spec) ([]TestResult, error) {
	var testResults []TestResult

	path := "exercises/"
	for _, file := range inputs {

		split := strings.SplitN(file, "\n", 2)
		fileName := path + strings.TrimSpace(split[0])
		fileContent := split[1]

		if fileName == "" {
			err := errors.New("File unnamed")
			return nil, err
		}

		if len(split) == 2 {
			os.WriteFile(fileName, []byte(fileContent), 0644)
		}
	}

	splitBuild := strings.SplitN(spec.Build.Cmd, " ", 2)
	args := strings.Split(splitBuild[1], " ")
	output := args[len(args)-1]
	buildCmd := exec.Command(splitBuild[0], args...)

	var stdout, stderr bytes.Buffer
	buildCmd.Stdout = &stdout
	buildCmd.Stderr = &stderr

	err := buildCmd.Run()
	if err != nil {
		log.Fatalf("Build failed: %v\nstderr: %s\n", err, stderr.String())
		return nil, err
	}

	for _, test := range spec.Tests {

		inputContent, err := os.ReadFile(test.Input)
		args := strings.Split(string(inputContent), " ")
		execCmd := exec.Command(test.Run, args...)

		stdout.Reset()
		stderr.Reset()
		execCmd.Stdout = &stdout
		execCmd.Stderr = &stderr
		err = execCmd.Run()

		if err != nil {
			log.Fatalf("Exec failed: %v\nstderr: %s\n", err, stderr.String())
			return nil, err
		}

		currentOutput := strings.Split(stdout.String(), "\n")
		var result TestResult
		expectedContent, err := os.ReadFile("/exercises/" + test.Expected)
		if err != nil {
			log.Fatalf("Read file \"expected output\" failed: %v\n", err)
			return nil, err
		}
		expectedOutput := strings.Split(string(expectedContent), "\n")
		result.Name = test.Name
		result.Success = true
		if len(expectedOutput) != len(currentOutput) {
			result.Success = false
		} else {
			for i := range expectedOutput {
				if expectedOutput[i] != currentOutput[i] {
					result.Success = false
					break
				}
			}
		}
		testResults = append(testResults, result)
	}

	return testResults, err
}

func funcHandler(inputs []string, spec Spec) ([]TestResult, error) {
	var testResult []TestResult

	//compile and compare result

	return testResult, nil
}

func textHandler(inputs []string, spec Spec) ([]TestResult, error) {
	var result []TestResult

	return result, nil
}

func mcqHandler(inputs []string, spec Spec) ([]TestResult, error) {
	var result []TestResult

	expectedContent, err := os.ReadFile("/exercises/" + spec.Tests[0].Expected)
	if err != nil {
		log.Fatalf("Read file \"expected output\" failed: %v\n", err)
		return nil, err
	}
	expectedOutput := strings.Split(string(expectedContent), "\n")

	result[0].Name = spec.Tests[0].Name
	result[0].Success = true
	for i := range expectedOutput {
		if expectedOutput[i] != inputs[i] {
			result[0].Success = false
			break
		}
	}

	return result, nil
}

func prepareResponse(testResults []TestResult) Result {

	overallSuccess := true

	for _, tr := range testResults {
		if !tr.Success {
			overallSuccess = false
			break
		}
	}

	responsePayload := Result{
		Success: overallSuccess,
		Tests:   testResults,
	}

	return responsePayload
}
