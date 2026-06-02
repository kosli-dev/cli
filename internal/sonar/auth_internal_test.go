package sonar

import "testing"

func TestIsSonarCloudHost(t *testing.T) {
	cases := []struct {
		host string
		want bool
	}{
		{"sonarcloud.io", true},
		{"SonarCloud.io", true},
		{"sonarcloud.io:443", true},
		{"api.sonarcloud.io", true},
		{"www.sonarcloud.io", true},
		{"sonarqube.mycorp.local", false},
		{"adcm5929.adcbmis.local:9000", false},
		{"localhost:9000", false},
		{"notsonarcloud.io", false},       // must not match by accident
		{"sonarcloud.io.evil.com", false}, // suffix-spoof must be rejected
	}
	for _, c := range cases {
		if got := isSonarCloudHost(c.host); got != c.want {
			t.Errorf("isSonarCloudHost(%q) = %v, want %v", c.host, got, c.want)
		}
	}
}
