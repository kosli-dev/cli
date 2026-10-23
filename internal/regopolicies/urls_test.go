package regopolicies

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPolicyURLKeepsTheIdentifierInOneSegment(t *testing.T) {
	endpoint, err := PolicyURL("https://app.kosli.com", "my-org", "../other/x")

	require.NoError(t, err)
	require.Equal(t, "https://app.kosli.com/api/v2/rego-policies/my-org/..%2Fother%2Fx", endpoint)
}

func TestPolicyURLRefusesAnIdentifierThatIsNotASegment(t *testing.T) {
	for _, identifier := range []string{"", ".", ".."} {
		_, err := PolicyURL("https://app.kosli.com", "my-org", identifier)

		require.Error(t, err, identifier)
	}
}

func TestVersionURLNamesTheVersion(t *testing.T) {
	endpoint, err := VersionURL("https://app.kosli.com", "my-org", "pr-approved", 3)

	require.NoError(t, err)
	require.Equal(t, "https://app.kosli.com/api/v2/rego-policies/my-org/pr-approved/versions/3", endpoint)
}

func TestListURLNamesTheOrg(t *testing.T) {
	endpoint, err := ListURL("https://app.kosli.com", "my org")

	require.NoError(t, err)
	require.Equal(t, "https://app.kosli.com/api/v2/rego-policies/my%20org", endpoint)
}

func TestBundleURLNamesTheBundleOfTheVersion(t *testing.T) {
	endpoint, err := BundleURL("https://app.kosli.com", "my-org", "pr-approved", 3)

	require.NoError(t, err)
	require.Equal(t, "https://app.kosli.com/api/v2/rego-policies/my-org/pr-approved/versions/3/bundle", endpoint)
}
