package security

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	keyring "github.com/zalando/go-keyring"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type SecurityTestSuite struct {
	suite.Suite
}

func (suite *SecurityTestSuite) TestAESEncryptionDecryption() {
	for _, t := range []struct {
		input string
	}{
		{input: "R3nD0m5tr!ng"},
		{input: "7H3Qu!ckBr0wnF0x7H3Qu!ckBr0wnF0x"},
		{input: "P@ssw0rd123"},
		{input: "S3cur3P@55w0rd"},
		{input: "3st!ng123"},
	} {
		suite.Run(t.input, func() {
			keyBytes, err := GenerateRandomAESKey()
			require.NoError(suite.T(), err)
			require.Len(suite.T(), keyBytes, 32)

			encryptedBytes, err := AESEncrypt(t.input, keyBytes)
			require.NoError(suite.T(), err)

			decrypted_bytes, err := AESDecrypt(encryptedBytes, keyBytes)
			require.NoError(suite.T(), err)
			require.Equal(suite.T(), t.input, string(decrypted_bytes))

		})
	}
}

func (suite *SecurityTestSuite) TestSetSecretInCredentialsStore() {
	keyring.MockInit()
	secretName := "topsecret"
	secretValue := "securepassword"
	err := SetSecretInCredentialsStore(secretName, secretValue)
	require.NoError(suite.T(), err)
	returnedSecretValue, err := GetSecretFromCredentialsStore(secretName)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), secretValue, returnedSecretValue)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(SecurityTestSuite))
}
