package security

import (
	"net/http"
	"testing"
	"time"
)

func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https URL",
			url:     "https://ci.example.com/webhook",
			wantErr: false,
		},
		{
			name:    "valid http URL with port",
			url:     "http://ci.internal.corp:8443/events",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			url:     "   ",
			wantErr: true,
		},
		{
			name:    "file scheme rejected",
			url:     "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "ftp scheme rejected",
			url:     "ftp://ftp.example.com/upload",
			wantErr: true,
		},
		{
			name:    "gopher scheme rejected",
			url:     "gopher://127.0.0.1:6379",
			wantErr: true,
		},
		{
			name:    "cloud metadata IP rejected (AWS/GCP/Azure)",
			url:     "http://169.254.169.254/latest/meta-data/",
			wantErr: true,
		},
		{
			name:    "gcp metadata hostname rejected",
			url:     "http://metadata.google.internal/computeMetadata/v1/",
			wantErr: true,
		},
		{
			name:    "instance-data alias rejected",
			url:     "http://instance-data/latest/meta-data/",
			wantErr: true,
		},
		{
			name:    "rfc1918 10.0.0.0/8 rejected",
			url:     "http://10.0.0.1:8080/webhook",
			wantErr: true,
		},
		{
			name:    "rfc1918 172.16.0.0/12 rejected",
			url:     "http://172.16.0.1:8080/webhook",
			wantErr: true,
		},
		{
			name:    "rfc1918 192.168.0.0/16 rejected",
			url:     "http://192.168.1.1:8080/webhook",
			wantErr: true,
		},
		{
			name:    "carrier-grade nat 100.64.0.0/10 rejected",
			url:     "http://100.64.0.1:8080/webhook",
			wantErr: true,
		},
		{
			name:    "ipv6 unique local address (ULA) rejected",
			url:     "http://[fd00::1]:8080/webhook",
			wantErr: true,
		},
		{
			name:    "ipv6 link local unicast metadata rejected",
			url:     "http://[fe80::1]/test",
			wantErr: true,
		},
		{
			name:    "loopback IPv4 rejected",
			url:     "http://127.0.0.1:8080/events",
			wantErr: true,
		},
		{
			name:    "loopback IPv4 other address in /8 rejected",
			url:     "http://127.0.0.2:8080/events",
			wantErr: true,
		},
		{
			name:    "localhost rejected",
			url:     "http://localhost:8080/events",
			wantErr: true,
		},
		{
			name:    "subdomain of localhost rejected",
			url:     "http://api.localhost:3000/webhook",
			wantErr: true,
		},
		{
			name:    "unspecified IPv4 0.0.0.0 rejected",
			url:     "http://0.0.0.0:8080/events",
			wantErr: true,
		},
		{
			name:    "loopback IPv6 rejected",
			url:     "http://[::1]:8080/events",
			wantErr: true,
		},
		{
			name:    "unspecified IPv6 rejected",
			url:     "http://[::]:8080/events",
			wantErr: true,
		},
		{
			name:    "no scheme",
			url:     "ci.example.com/webhook",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWebhookURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWebhookURL(%q) err = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestSafeTransport_BlocksPrivateNetwork(t *testing.T) {
	transport := NewSafeTransport(false)
	client := &http.Client{
		Transport: transport,
		Timeout:   2 * time.Second,
	}

	// Dialing an RFC 1918 private IP should be blocked by DialContext immediately without network wait
	_, err := client.Get("http://10.255.255.1:8080/events")
	if err == nil {
		t.Errorf("expected error connecting to private network, got nil")
	}

	// Dialing loopback should also be blocked by default
	_, err = client.Get("http://127.0.0.1:8080/events")
	if err == nil {
		t.Errorf("expected error connecting to loopback, got nil")
	}
}
