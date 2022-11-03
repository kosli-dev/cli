package test_support

import (
	"fmt"
	"github.com/kosli-dev/cli/internal/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

const ImageName = "library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5"

func PullExampleImage(suite *testing.T) {
	err := utils.PullDockerImage(ImageName)
	require.NoError(suite, err, fmt.Sprintf("PullExampleImage: %s", ImageName))
}
