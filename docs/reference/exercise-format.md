---
layout: default
title: Exercise Format
parent: Reference
nav_order: 2
---

# Exercise Format

Specification for exercise files and structure.

## File Structure

## Metadata

## Test Cases

## Example Exercise

```yaml
id: "ex01_hello_world"
title: "Hello World"
type: "code"  # or "qcm", "text"
language: "c"

# Compilation (if needed)
build:
  command: "gcc -Wall -Wextra -Werror {submission} -o {output}"
  timeout: 30

# Test execution
tests:
  - name: "Basic test"
    run: "./{output}"
    input: "tests/test1_input.txt"
    expected_output: "tests/test1_expected.txt"
    timeout: 5
    
  - name: "Edge case"
    run: "./{output}"
    input: "tests/test2_input.txt"
    expected_output: "tests/test2_expected.txt"

# Validation rules
validation:
  check_valgrind: true
  allowed_functions: ["write", "malloc", "free"]
  forbidden_functions: ["printf"]  # For 42 exercises

# Scoring
scoring:
  compilation: 10
  per_test: 15
  valgrind_clean: 10
```


```yaml
id: "ex01_regex"
title: "regex"
type: "input"  # or "qcm", "text"
language: "none"

# Test execution
tests:
  - name: "Basic test"
    run: "./{output}"
    input: "tests/test1_input.txt"
    expected_output: "tests/test1_expected.txt"
    timeout: 5
    
  - name: "Edge case"
    run: "./{output}"
    input: "tests/test2_input.txt"
    expected_output: "tests/test2_expected.txt"

# Scoring
scoring:
  compilation: 10
  per_test: 15
  valgrind_clean: 10
```
regex/
    exo1.yaml
    test/

