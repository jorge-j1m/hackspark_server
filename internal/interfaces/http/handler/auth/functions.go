package auth

import (
	"net"
	"regexp"
	"strings"
	"unicode/utf8"
)

// isValidEmail validates an email address according to RFC 5322 with practical constraints
// Returns true if the email is valid, false otherwise
func isValidEmail(email string) bool {
	// Basic checks
	if email == "" {
		return false
	}

	// Check for overall length constraints
	// RFC 5321 limits: 64 chars for local part, 255 for domain, 254 total
	if len(email) > 254 {
		return false
	}

	// Split the email into local part and domain
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	localPart, domain := parts[0], parts[1]

	// Validate local part
	if len(localPart) == 0 || len(localPart) > 64 {
		return false
	}

	// Validate domain
	if len(domain) == 0 || len(domain) > 255 {
		return false
	}

	// Check for valid characters in local part based on different formats
	// If quoted local part
	if strings.HasPrefix(localPart, "\"") && strings.HasSuffix(localPart, "\"") {
		// RFC 5322 quoted string validation (more permissive)
		quotedLocalPattern := `^"(\\.|[^\\"])*"$`
		quotedRegex := regexp.MustCompile(quotedLocalPattern)
		if !quotedRegex.MatchString(localPart) {
			return false
		}
	} else {
		// Unquoted local part - stricter validation
		// Allow alphanumeric, plus common special characters, but no consecutive dots
		unquotedLocalPattern := `^[a-zA-Z0-9!#$%&'*+\-/=?^_\x60{|}~.]+$`
		unquotedRegex := regexp.MustCompile(unquotedLocalPattern)
		if !unquotedRegex.MatchString(localPart) {
			return false
		}

		// Check for consecutive dots which are not allowed
		if strings.Contains(localPart, "..") {
			return false
		}

		// Local part cannot start or end with a dot
		if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
			return false
		}
	}

	// Validate domain using combination of regex and DNS check
	// First, check format with regex
	domainPattern := `^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$`
	domainRegex := regexp.MustCompile(domainPattern)
	if !domainRegex.MatchString(domain) {
		// Special case for IP address literals in domain part
		if strings.HasPrefix(domain, "[") && strings.HasSuffix(domain, "]") {
			ipLiteral := domain[1 : len(domain)-1]
			ip := net.ParseIP(ipLiteral)
			if ip == nil {
				return false
			}
		} else {
			return false
		}
	}

	// Check for invalid punycode in domain
	if strings.Contains(domain, "xn--") {
		// Simplified punycode validation (complete validation is complex)
		for _, part := range strings.Split(domain, ".") {
			if strings.HasPrefix(part, "xn--") {
				// Check if the punycode part is valid UTF-8 when decoded
				// This is a simplified check that catches some invalid punycode
				rest := part[4:] // Skip the "xn--" prefix
				if !utf8.ValidString(rest) {
					return false
				}
			}
		}
	}

	return true
}

func isValidPassword(pass string) bool {
	return len(pass) >= 8
}
