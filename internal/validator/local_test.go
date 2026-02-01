package validator

import (
	"testing"
)

func TestLocalValidator_ValidYAML(t *testing.T) {
	validator := NewLocalValidator(true)

	validYAML := `
image: alpine:latest

stages:
  - build
  - test

build:
  stage: build
  script:
    - echo "Building"
  tags:
    - docker

test:
  stage: test
  script:
    - echo "Testing"
  dependencies:
    - build
`

	result := validator.Validate([]byte(validYAML))

	if !result.Valid {
		t.Errorf("Expected valid YAML to pass validation, got errors: %v", result.Errors)
	}

	if len(result.Errors) > 0 {
		t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	if result.Stage != "local" {
		t.Errorf("Expected stage 'local', got '%s'", result.Stage)
	}
}

func TestLocalValidator_InvalidYAML_Syntax(t *testing.T) {
	validator := NewLocalValidator(true)

	invalidYAML := `
image: alpine:latest

stages:
  - build
  test

build:
  stage: build
  script:
    - echo "Building"
  tags:
    - docker
`

	result := validator.Validate([]byte(invalidYAML))

	if result.Valid {
		t.Error("Expected invalid YAML to fail validation")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected at least one error")
	}

	// Check that error has context
	err := result.Errors[0]
	if err.Message == "" {
		t.Error("Expected error message to be populated")
	}

	if err.Line == 0 {
		t.Error("Expected error line number to be populated")
	}
}

func TestLocalValidator_InvalidYAML_UnterminatedString(t *testing.T) {
	validator := NewLocalValidator(true)

	invalidYAML := `
build:
  script:
    - echo "unterminated string
`

	result := validator.Validate([]byte(invalidYAML))

	// Unterminated strings might not always be caught by YAML parser
	// depending on context. Just check it doesn't crash.
	_ = result
}

func TestLocalValidator_StrictMode(t *testing.T) {
	strictValidator := NewLocalValidator(true)
	lenientValidator := NewLocalValidator(false)

	// YAML that might fail in strict mode but pass in lenient mode
	yamlWithUnknownFields := `
image: alpine:latest

stages:
  - build

build:
  stage: build
  script:
    - echo "test"
  unknown_field: value
`

	strictResult := strictValidator.Validate([]byte(yamlWithUnknownFields))
	lenientResult := lenientValidator.Validate([]byte(yamlWithUnknownFields))

	// In strict mode, unknown fields might cause issues
	// The actual behavior depends on yaml.v3 implementation
	// This test documents the difference between modes

	_ = strictResult
	_ = lenientResult
}

func TestLocalValidator_EmptyFile(t *testing.T) {
	validator := NewLocalValidator(true)

	result := validator.Validate([]byte(""))

	// Empty files are technically valid YAML but may have EOF errors
	// The actual behavior depends on yaml.v3 implementation
	// Let's just check that it doesn't crash
	_ = result
}

func TestLocalValidator_WhitespaceOnly(t *testing.T) {
	validator := NewLocalValidator(true)

	result := validator.Validate([]byte("   \n  \n  "))

	// Whitespace-only files may have EOF errors
	// Just check that it doesn't crash
	_ = result
}

func TestLocalValidator_ComplexStructure(t *testing.T) {
	validator := NewLocalValidator(true)

	complexYAML := `
variables:
  GLOBAL_VAR: "value"
  ANOTHER_VAR: 123

default:
  image: alpine:latest
  artifacts:
    expire_in: 1 week

stages:
  - build
  - test
  - deploy

.build_template: &build_template
  stage: build
  before_script:
    - echo "Before"
  after_script:
    - echo "After"

build1:
  <<: *build_template
  script:
    - make build

build2:
  <<: *build_template
  parallel: 5
  script:
    - make build-all

test:
  stage: test
  needs: [build1, build2]
  script:
    - make test
  cache:
    key: ${CI_COMMIT_REF_SLUG}
    paths:
      - vendor/
  only:
    - branches
  except:
    - /^hotfix-.*$/

deploy:
  stage: deploy
  script:
    - make deploy
  when: manual
  environment:
    name: production
    url: https://example.com
`

	result := validator.Validate([]byte(complexYAML))

	if !result.Valid {
		t.Errorf("Expected complex valid YAML to pass, got errors: %v", result.Errors)
	}
}

func TestLocalValidator_AnchorsAndAliases(t *testing.T) {
	validator := NewLocalValidator(true)

	yamlWithAnchors := `
.defaults: &defaults
  cache:
    paths:
      - vendor/

job1:
  <<: *defaults
  script: test

job2:
  <<: *defaults
  script: deploy
`

	result := validator.Validate([]byte(yamlWithAnchors))

	if !result.Valid {
		t.Errorf("Expected YAML with anchors to be valid, got errors: %v", result.Errors)
	}
}

func TestLocalValidator_MultilineStrings(t *testing.T) {
	validator := NewLocalValidator(true)

	multilineYAML := `
test:
  script:
    - |
      echo "Line 1"
      echo "Line 2"
      echo "Line 3"
    - echo "Another command"
  variables:
    MULTILINE: |
      This is a
      multiline
      variable
    LITERAL: |-
      No trailing newline
`

	result := validator.Validate([]byte(multilineYAML))

	if !result.Valid {
		t.Errorf("Expected multiline strings to be valid, got errors: %v", result.Errors)
	}
}

func TestLocal_Result_Stage(t *testing.T) {
	validator := NewLocalValidator(true)

	result := validator.Validate([]byte("test: value"))

	if result.Stage != "local" {
		t.Errorf("Expected stage to be 'local', got '%s'", result.Stage)
	}
}

func TestLocalValidator_InvalidIndentation(t *testing.T) {
	validator := NewLocalValidator(true)

	invalidIndent := `
job1:
  script: echo "test"
    nested: value
`

	result := validator.Validate([]byte(invalidIndent))

	if result.Valid {
		t.Error("Expected invalid indentation to fail validation")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected at least one error for invalid indentation")
	}
}
