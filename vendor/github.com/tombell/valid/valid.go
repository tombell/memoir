package valid

import (
	"strings"
)

// Case represents a validation test case. Contains a condition that is a the
// boolean result of the check, and an error message.
type Case struct {
	Cond bool
	Msg  string
}

// Error is a map of validation error messages.
type Error map[string][]string

// Error returns a formatted string of all the validatin error messages.
func (e Error) Error() string {
	var combined []string

	for key, val := range e {
		for _, msg := range val {
			combined = append(combined, key+": "+msg)
		}
	}

	return strings.Join(combined, ", ")
}

// Validator performs a number of validation checks and determines if eveything
// is valid or not. Contains the map of error messages.
type Validator struct {
	Errors Error `json:"errors"`
}

// New returns an instance of Validator.
func New() *Validator {
	return &Validator{
		Errors: make(map[string][]string),
	}
}

// Check goes through the validation test cases and adds any validation errors
// for failing cases.
func (v *Validator) Check(key string, cases ...Case) {
	for _, c := range cases {
		if !c.Cond {
			v.Errors[key] = append(v.Errors[key], c.Msg)
		}
	}
}

// Valid returns if the validation has passed or not.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}
