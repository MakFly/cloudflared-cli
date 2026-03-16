package config

import (
	"fmt"
	"regexp"
	"strings"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ValidationError represents a single validation issue.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) String() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult holds all validation errors.
type ValidationResult struct {
	Errors []ValidationError
}

// IsValid returns true if there are no validation errors.
func (r ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// Error returns a formatted string of all validation errors.
func (r ValidationResult) Error() string {
	if r.IsValid() {
		return ""
	}
	var msgs []string
	for _, e := range r.Errors {
		msgs = append(msgs, "  - "+e.String())
	}
	return fmt.Sprintf("validation failed:\n%s", strings.Join(msgs, "\n"))
}

// ValidateTunnelConfig checks a tunnel configuration for common issues.
func ValidateTunnelConfig(cfg *TunnelConfig) ValidationResult {
	var result ValidationResult

	if cfg.Tunnel == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "tunnel",
			Message: "tunnel ID is required",
		})
	} else if !isValidUUID(cfg.Tunnel) && !isValidTunnelName(cfg.Tunnel) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "tunnel",
			Message: "must be a valid UUID or tunnel name",
		})
	}

	if cfg.CredentialsFile == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "credentials-file",
			Message: "credentials file path is required",
		})
	}

	if len(cfg.Ingress) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "ingress",
			Message: "at least one ingress rule is required",
		})
	} else {
		lastRule := cfg.Ingress[len(cfg.Ingress)-1]
		if !lastRule.IsCatchAll() {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "ingress",
				Message: "last ingress rule must be a catch-all (no hostname or path)",
			})
		}

		for i, rule := range cfg.Ingress {
			if rule.Service == "" {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("ingress[%d].service", i),
					Message: "service is required",
				})
			}
			// Non-catch-all rules should have a hostname
			if i < len(cfg.Ingress)-1 && rule.Hostname == "" && rule.Path == "" {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("ingress[%d]", i),
					Message: "non-catch-all rules must have a hostname or path",
				})
			}
		}
	}

	return result
}

func isValidUUID(s string) bool {
	return uuidRegex.MatchString(strings.ToLower(s))
}

func isValidTunnelName(s string) bool {
	return len(s) > 0 && len(s) <= 253 && !strings.ContainsAny(s, " \t\n\r")
}
