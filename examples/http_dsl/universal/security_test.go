package universal

import (
	"net"
	"testing"
)

func TestVuln17SafeXPathAndPrivateIPGuard(t *testing.T) {
	he := NewHTTPEngine()
	he.lastResponseBody = `<div class="x">hello</div>`

	// Malformed regex meta-characters in XPath tag name must NEVER panic!
	if got := he.Extract("xpath", "//[unclosed"); got != nil {
		t.Fatalf("expected nil on non-matching meta tag, got %v", got)
	}
	if got := he.Extract("xpath", "//div"); got != "hello" {
		t.Fatalf("expected 'hello' from //div, got %v", got)
	}

	// Private / loopback / cloud-metadata IPs are detected by SSRF guard.
	for _, rawIP := range []string{"127.0.0.1", "10.0.0.5", "192.168.1.1", "169.254.169.254", "::1"} {
		ip := net.ParseIP(rawIP)
		if !isDisallowedPrivateIP(ip) {
			t.Errorf("expected %s to be disallowed by SSRF guard", rawIP)
		}
	}
	if isDisallowedPrivateIP(net.ParseIP("93.184.216.34")) {
		t.Errorf("expected public IP 93.184.216.34 to be allowed")
	}
}
