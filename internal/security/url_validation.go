// Package security provides enterprise security scanning, push protection, and URL validation for refweir.dev.
package security

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	// ErrInvalidURLScheme indicates the URL does not use an allowed http or https scheme.
	ErrInvalidURLScheme = errors.New("webhook URL scheme must be http or https")
	// ErrEmptyURLHost indicates the URL is missing a hostname.
	ErrEmptyURLHost = errors.New("webhook URL host cannot be empty")
	// ErrBlockedMetadataHost indicates the URL targets a restricted cloud metadata address.
	ErrBlockedMetadataHost = errors.New("webhook URL host targets a restricted cloud metadata address")
	// ErrBlockedLoopbackHost indicates the URL targets loopback or localhost.
	ErrBlockedLoopbackHost = errors.New("webhook URL cannot target loopback or localhost")
	// ErrBlockedUnspecifiedHost indicates the URL targets an unspecified network address.
	ErrBlockedUnspecifiedHost = errors.New("webhook URL cannot target unspecified network address")
	// ErrBlockedPrivateNetwork indicates the URL targets a private or restricted network address (RFC 1918 / ULA).
	ErrBlockedPrivateNetwork = errors.New("webhook URL targets a private or restricted network address")

	cgnatNet = &net.IPNet{
		IP:   net.ParseIP("100.64.0.0"),
		Mask: net.CIDRMask(10, 32),
	}
)

func isPrivateOrBlockedIP(ip net.IP, allowPrivate bool, allowLoopback bool) error {
	if ip == nil {
		return nil
	}
	if !allowLoopback && ip.IsLoopback() {
		return ErrBlockedLoopbackHost
	}
	if ip.IsUnspecified() {
		return ErrBlockedUnspecifiedHost
	}
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return ErrBlockedMetadataHost
	}
	if !allowPrivate {
		if ip.IsPrivate() || (cgnatNet != nil && cgnatNet.Contains(ip)) {
			return ErrBlockedPrivateNetwork
		}
	}
	return nil
}

// ValidateWebhookURL verifies that a target webhook URL is syntactically valid,
// uses an allowed HTTP/HTTPS scheme, and does not target cloud instance metadata endpoints,
// private internal networks, or loopback (SSRF defense).
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

	lowerHost := strings.ToLower(hostname)

	// Check if loopback or private networks are permitted for testing or isolated local harnesses
	allowLoopback := os.Getenv("REFWEIR_ALLOW_LOOPBACK_WEBHOOKS") == "true"
	allowPrivate := os.Getenv("REFWEIR_ALLOW_PRIVATE_WEBHOOKS") == "true" || allowLoopback

	// Block loopback hostname and subdomains (localhost, *.localhost)
	if !allowLoopback && (lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".localhost")) {
		return ErrBlockedLoopbackHost
	}

	// Block AWS / GCP / Azure link-local metadata addresses and aliases
	if lowerHost == "169.254.169.254" || lowerHost == "metadata.google.internal" || lowerHost == "instance-data" {
		return ErrBlockedMetadataHost
	}

	// Check if IP matches loopback, unspecified, link-local metadata, or private internal network
	if ip := net.ParseIP(hostname); ip != nil {
		if err := isPrivateOrBlockedIP(ip, allowPrivate, allowLoopback); err != nil {
			return err
		}
	}

	return nil
}

// NewSafeTransport returns an http.Transport configured with a custom DialContext that validates
// IP addresses at connection time to prevent DNS rebinding attacks and SSRF into private networks.
func NewSafeTransport(insecureSSL bool) *http.Transport {
	allowLoopback := os.Getenv("REFWEIR_ALLOW_LOOPBACK_WEBHOOKS") == "true"
	allowPrivate := os.Getenv("REFWEIR_ALLOW_PRIVATE_WEBHOOKS") == "true" || allowLoopback

	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	return &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSSL, // #nosec G402 -- controlled by repo webhook setting
		},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, fmt.Errorf("invalid target address %q: %w", addr, err)
			}

			// Validate IP literal directly
			if ip := net.ParseIP(host); ip != nil {
				if err := isPrivateOrBlockedIP(ip, allowPrivate, allowLoopback); err != nil {
					return nil, fmt.Errorf("connection to %s blocked: %w", host, err)
				}
				return dialer.DialContext(ctx, network, addr)
			}

			// Resolve host and verify that all resolved IP addresses are safe
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, fmt.Errorf("resolve host %s: %w", host, err)
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("no IP addresses found for host %s", host)
			}

			for _, ip := range ips {
				if err := isPrivateOrBlockedIP(ip, allowPrivate, allowLoopback); err != nil {
					return nil, fmt.Errorf("connection to %s (%s) blocked: %w", host, ip.String(), err)
				}
			}

			// Connect to the first verified IP literal to eliminate DNS rebinding
			targetAddr := net.JoinHostPort(ips[0].String(), port)
			return dialer.DialContext(ctx, network, targetAddr)
		},
	}
}
