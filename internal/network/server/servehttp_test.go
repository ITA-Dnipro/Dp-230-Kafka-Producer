package server

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestServerHTTP(t *testing.T) {
	originalCertPath := os.Getenv("PATH_TO_CERTS")
	defer os.Setenv("PATH_TO_CERTS", originalCertPath)

	tests := []struct {
		Name       string
		CertPath   string
		Timeout    time.Duration
		Addr       string
		tlsEnabled bool
	}{
		{
			Name:       "with tls",
			CertPath:   "../../../certs/",
			Timeout:    time.Second,
			Addr:       fmt.Sprintf(":%d", 8989),
			tlsEnabled: true,
		},
		{
			Name:       "http without tls",
			CertPath:   "fakepath",
			Timeout:    time.Second,
			Addr:       fmt.Sprintf(":%d", 9898),
			tlsEnabled: false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			os.Setenv("PATH_TO_CERTS", test.CertPath)
			server := NewServerHTTP(test.Addr)
			server.SetShutdownTimeout(test.Timeout)
			if server.shutdownTimeout != test.Timeout {
				t.Errorf("error setting server shutdown timeout. Expected: %s, But got: %s", test.Timeout, server.shutdownTimeout)
			}
			if test.tlsEnabled && server.Server.TLSConfig == nil ||
				!test.tlsEnabled && server.Server.TLSConfig != nil {
				t.Errorf("error setting server TLS parameter. Expected: %v, But got: %v", test.tlsEnabled, !test.tlsEnabled)
			}
			go func() { server.Start(nil) }()
			<-time.After(100 * time.Millisecond)
			errClose := server.Close()
			if errClose != nil {
				t.Error("error shutting down")
			}
		})
	}
}
