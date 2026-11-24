// Package security provides enterprise security scanning, push protection, and URL validation for refweir.dev.
package security

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

var (
	// ErrInvalidURLScheme indicates the URL does not use an allowed http or https scheme.
	ErrInvalidURLScheme = errors.New("webhook URL scheme must be http or https")
	// ErrEmptyURLHost indicates the URL is missing a hostname.
	ErrEmptyURLHost = errors.New("webhook URL host cannot be empty")
	// ErrBlockedMetadataHost indicates the URL targets a restricted cloud metadata address.
	ErrBlockedMetadataHost = errors.New("webhook URL host targets a restricted cloud metadata address")
)

// ValidateWebhookURL verifies that a target webhook URL is syntactically valid,
// uses an allowed HTTP/HTTPS scheme, and does not target cloud instance metadata endpoints (SSRF defense).
func ValidateWebhookURL(rawURL string) error {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return errors.New("webhook URL cannot be empty")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return ErrInvalidURLScheme
	}

	hostname := u.Hostname()
	if hostname == "" {
		return ErrEmptyURLHost
	}

	// Block AWS / GCP / Azure link-local metadata addresses and aliases
	lowerHost := strings.ToLower(hostname)
	if lowerHost == "169.254.169.254" || lowerHost == "metadata.google.internal" || lowerHost == "instance-data" {
		return ErrBlockedMetadataHost
	}

	// Check if IP matches link-local metadata subnet (169.254.0.0/16 or IPv6 link-local)
	if ip := net.ParseIP(hostname); ip != nil {
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return ErrBlockedMetadataHost
		}
	}

	return nil
}
