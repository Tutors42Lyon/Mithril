package main

import (
	"bytes"
	"encoding/json"
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
	Id    string `yaml:"id"`
	Type  string `yaml:"type"`
	Tests []Test `yaml:"tests"`
}

type Build struct {
	Cmd     string `yaml:"command"`
	Timeout string `yaml:"timeout"`
}

type Test struct {
	Name     string `yaml:"name"`
	Run      string `yaml:"run"`
	Expected string `yaml:"expected_output"`
	Timeout  int    `yaml:"timeout"`
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
	testResults := testInput(spec.Type, inputs)

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

func testInput(exerciseCase string, inputs []string, build Build) []TestResult {

	switch {
	case exerciseCase == "program":
		return programHandler(inputs, build)
	case exerciseCase == "function":
		return funcHandler(inputs, build)
	case exerciseCase == "text":
		return textHandler(inputs)
	case exerciseCase == "mcq":
		return mcqHandler(inputs)
	}
	var tab []TestResult
	return tab
}

func programHandler(inputs []string, build Build) []TestResult {
	var testResult []TestResult

	path := "exercises/"
	for _, file := range inputs {

		split := strings.SplitN(file, "\n", 2)
		fileName := path + strings.TrimSpace(split[0])
		fileContent := split[1]

		if fileName == "" {
			return nil
		}

		if len(split) == 2 {
			os.WriteFile(fileName, []byte(fileContent), 0644)
		}
	}

	splitBuild := strings.SplitN(build.Cmd, "\n", 2)
	args := strings.Split(splitBuild[1], "\n")
	cmd := exec.Command(splitBuild[0], args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Build failed: %v\nstderr: %s\n", err, stderr.String())
	}
	
	//check allowed/forbidden function

	return testResult
}

func funcHandler(inputs []string, build Build) []TestResult {
	var testResult []TestResult

	path := "exercises/"
	for _, file := range inputs {

		split := strings.SplitN(file, "\n", 2)
		fileName := path + strings.TrimSpace(split[0])
		fileContent := split[1]

		if fileName == "" {
			return nil
		}

		if len(split) == 2 {
			os.WriteFile(fileName, []byte(fileContent), 0644)
		}
	}

	splitBuild := strings.SplitN(build.Cmd, "\n", 2)
	args := strings.Split(splitBuild[1], "\n")
	cmd := exec.Command(splitBuild[0], args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Build failed: %v\nstderr: %s\n", err, stderr.String())
	}

	//compile and compare result

	return testResult
}

func textHandler(inputs []string) []TestResult {
	var testResult []TestResult

	//execute test with input and compare result

	return testResult
}

func mcqHandler(inputs []string) []TestResult {
	var testResult []TestResult

	//compare input with test

	return testResult
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
