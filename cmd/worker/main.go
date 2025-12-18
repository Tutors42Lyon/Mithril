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

var path_ string

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
	Executable *string     `yaml:"executable"`
	Build      *Build      `yaml:"build"`
	Tests      []Test      `yaml:"tests"`
	Validation *Validation `yaml:"validation"`
}

type Build struct {
	Output  string `yaml:"output"`
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

	testResults, err := testInput(*spec, inputs)
	if err != nil {
		log.Printf("Error test input: %v", err)
		return
	}

	result := prepareResponse(testResults)
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Printf("Marshal Error for response: %v", err)
		m.Respond([]byte(`{"success": false, "error": "Internal server error during JSON encoding"}`))
		return
	}

	m.Respond(resultJSON)
}

func searchExerciseTest(exerciseId string) (*Spec, error) {

	var foundSpec *Spec = nil
	err := filepath.Walk("exercises", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".yaml") {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}

			var spec Spec

			if unmarshalErr := yaml.Unmarshal(data, &spec); unmarshalErr != nil {
				fmt.Printf("YAML unmarshal error in %s: %v", path, unmarshalErr)
				return nil
			}

			if spec.Id == exerciseId {
				foundSpec = &spec
				path_ = filepath.Dir(path)
				return fmt.Errorf("spec found")
			}
		}
		return nil
	})

	if err != nil && err.Error() == "spec found" {
		return foundSpec, nil
	}

	if err != nil {
		return nil, fmt.Errorf("file parsing error: %w", err)
	}

	return nil, fmt.Errorf("can't exercise with id: %s", exerciseId)
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

	path := "app/exercises/"
	for _, file := range inputs {

		split := strings.SplitN(file, "\n", 2)
		fileName := path + strings.TrimSpace(split[0])
		fileContent := split[1]

		if fileName == "" {
			err := errors.New("file unnamed")
			return nil, err
		}

		if len(split) == 2 {
			os.WriteFile(fileName, []byte(fileContent), 0644)
		}
	}

	splitBuild := strings.SplitN(spec.Build.Cmd, " ", 2)
	args := strings.Split(splitBuild[1], " ")
	cmd := strings.Replace(splitBuild[0], "{output}", spec.Build.Output, -1)
	buildCmd := exec.Command(cmd, args...)

	var stdout, stderr bytes.Buffer
	buildCmd.Stdout = &stdout
	buildCmd.Stderr = &stderr

	err := buildCmd.Run()
	if err != nil {
		log.Printf("Build failed: %v\nstderr: %s\n", err, stderr.String())
		return nil, err
	}

	for _, test := range spec.Tests {

		inputContent, err := os.ReadFile(test.Input)
		if err != nil {
			log.Printf("ReadFile failed: %v", err)
			return nil, err
		}
		args := strings.Split(string(inputContent), " ")
		execCmd := exec.Command(test.Run, args...)

		stdout.Reset()
		stderr.Reset()
		execCmd.Stdout = &stdout
		execCmd.Stderr = &stderr
		err = execCmd.Run()

		if err != nil {
			log.Printf("1Exec failed: %v\nstderr: %s\n", err, stderr.String())
			return nil, err
		}

		currentOutput := strings.Split(stdout.String(), "\n")
		var result TestResult
		expectedContent, err := os.ReadFile("/app/exercises/" + test.Expected)
		if err != nil {
			log.Printf("Read file \"expected output\" failed: %v\n", err)
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

	return testResult, nil
}

func textHandler(inputs []string, spec Spec) ([]TestResult, error) {
	var results []TestResult

	for i, test := range spec.Tests {

		/* Get test inputs and set them in a string by line */
		inputContent, err := os.ReadFile(path_ + "/" + test.Input)
		if err != nil {
			log.Printf("ReadFile failed: %v", err)
			return nil, err
		}
		argTest := strings.Split(string(inputContent), "\n")

		/* Create args for command */
		args := argTest
		args = append(args, inputs[i])
		cmd := path_ + "/" + test.Run
		buildCmd := exec.Command(cmd, args...)

		log.Printf("_____cmd : %v, args : %v", cmd, args)

		for i, arg := range args {
			log.Printf("arg[%v] : %v", i, arg)
		}


		var stdout, stderr bytes.Buffer
		buildCmd.Stdout = &stdout
		buildCmd.Stderr = &stderr
		stdout.Reset()
		stderr.Reset()
		buildCmd.Stdout = &stdout
		buildCmd.Stderr = &stderr
		err = buildCmd.Run()
		if err != nil {
			log.Printf("Exec failed: %v", err)
			return nil, err
		}

		var result TestResult
		result.Name = test.Name

		currentOutput := strings.Split(stdout.String(), "\n")
		expectedContent, err := os.ReadFile(path_ + "/" + test.Expected)
		if err != nil {
			log.Printf("Read file \"expected output\" failed: %v\n", err)
			return nil, err
		}
		expectedOutput := strings.Split(string(expectedContent), "\n")
		log.Printf("output expected: %v", expectedOutput)
		log.Printf("output current: %v", currentOutput)
		result.Success = true
		if len(expectedOutput) != len(currentOutput) {
			log.Printf("len difference")
			result.Success = false
		} else {
			for i := range expectedOutput {
				if expectedOutput[i] != currentOutput[i] {
					result.Success = false
					break
				}
			}
		}
		results = append(results, result)
	}

	return results, nil
}

func mcqHandler(inputs []string, spec Spec) ([]TestResult, error) {
	var results []TestResult
	var result TestResult

	expectedContent, err := os.ReadFile("/app/exercises/" + spec.Tests[0].Expected)
	if err != nil {
		log.Printf("Read file \"expected output\" failed: %v\n", err)
		return nil, err
	}
	expectedOutput := strings.Split(string(expectedContent), "\n")

	result.Name = spec.Tests[0].Name
	result.Success = true
	for i := range expectedOutput {
		if expectedOutput[i] != inputs[i] {
			result.Success = false
			break
		}
	}

	results = append(results, result)
	return results, nil
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
