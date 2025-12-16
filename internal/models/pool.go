package models

type Pool struct {
	ID          string   `json:"id" yaml:"slug"`
	Name        string   `json:"name" yaml:"name"`
	Theme       string   `json:"theme" yaml:"theme"`
	Description string   `json:"description" yaml:"description"`
	Difficulty  string   `json:"difficulty" yaml:"difficulty"`
	Tags        []string `json:"tags" yaml:"tags"`
	Maintainer  string   `json:"maintainer" yaml:"maintainer"`
	Exercises   []string `json:"exercises"`
}

