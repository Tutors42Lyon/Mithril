package models

type Exercise struct {
	ID         string           `json:"id" yaml:"id"`
	PoolID     string           `json:"pool_id"`
	Title      string           `json:"title" yaml:"title"`
	Type       string           `json:"type" yaml:"type"` // code, input, qcm, text
	Language   string           `json:"language" yaml:"language"`
	Build      *BuildConfig     `json:"build,omitempty" yaml:"build"`
	Tests      []TestCase       `json:"tests" yaml:"tests"`
	Validation *ValidationRules `json:"validation,omitempty" yaml:"validation"`
	Scoring    *ScoringConfig   `json:"scoring,omitempty" yaml:"scoring"`
}

type BuildConfig struct {
	Command string `json:"command" yaml:"command"`
	Timeout int    `json:"timeout" yaml:"timeout"`
}

type TestCase struct {
	Name           string `json:"name" yaml:"name"`
	Run            string `json:"run" yaml:"run"`
	Input          string `json:"input" yaml:"input"`
	ExpectedOutput string `json:"expected_output" yaml:"expected_output"`
	Timeout        int    `json:"timeout" yaml:"timeout"`
}

type ValidationRules struct {
	CheckValgrind      bool     `json:"check_valgrind" yaml:"check_valgrind"`
	AllowedFunctions   []string `json:"allowed_functions" yaml:"allowed_functions"`
	ForbiddenFunctions []string `json:"forbidden_functions" yaml:"forbidden_functions"`
}

type ScoringConfig struct {
	Compilation   int `json:"compilation" yaml:"compilation"`
	PerTest       int `json:"per_test" yaml:"per_test"`
	ValgrindClean int `json:"valgrind_clean" yaml:"valgrind_clean"`
}

