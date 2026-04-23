package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateOutputFile_AcceptsYamlExtension(t *testing.T) {
	dir := t.TempDir()
	assert.NoError(t, validateOutputFile(filepath.Join(dir, "policy.yaml")))
}

func TestValidateOutputFile_AcceptsYmlExtension(t *testing.T) {
	dir := t.TempDir()
	assert.NoError(t, validateOutputFile(filepath.Join(dir, "policy.yml")))
}

func TestValidateOutputFile_AcceptsUppercaseExtension(t *testing.T) {
	dir := t.TempDir()
	assert.NoError(t, validateOutputFile(filepath.Join(dir, "policy.YAML")))
}

func TestValidateOutputFile_RejectsNonYamlExtension(t *testing.T) {
	dir := t.TempDir()
	err := validateOutputFile(filepath.Join(dir, "policy.json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".yaml or .yml")
}

func TestValidateOutputFile_RejectsNoExtension(t *testing.T) {
	dir := t.TempDir()
	err := validateOutputFile(filepath.Join(dir, "policy"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), ".yaml or .yml")
}

func TestValidateOutputFile_RejectsExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.yaml")
	require.NoError(t, os.WriteFile(path, []byte("test"), 0644))

	err := validateOutputFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}
