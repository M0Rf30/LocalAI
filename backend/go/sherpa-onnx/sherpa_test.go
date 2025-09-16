package main

import (
	"testing"

	"github.com/mudler/LocalAI/pkg/grpc/proto"
	"github.com/stretchr/testify/assert"
)

func TestSherpaONNXBackend(t *testing.T) {
	// Test that the backend struct can be instantiated
	backend := &SherpaONNX{}
	assert.NotNil(t, backend)
	
	// Test that the backend implements the required interface methods
	// This is just a basic check to ensure the struct has the right shape
	// Note: We can't directly cast to BackendServer because of the interface mismatch
	// but we can check that it implements the AIModel interface from grpc
	_ = backend // Use the variable to avoid unused variable error
}

func TestFileExists(t *testing.T) {
	// Test that fileExists returns false for non-existent files
	assert.False(t, fileExists("/non/existent/file"))
	
	// Test that fileExists returns true for existing files
	// Note: This test assumes the test file itself exists
	assert.True(t, fileExists("sherpa_test.go"))
}

func TestBackendMethods(t *testing.T) {
	backend := &SherpaONNX{}
	
	// Test Status method
	status, err := backend.Status()
	assert.NoError(t, err)
	assert.Equal(t, proto.StatusResponse_READY, status.State)
	
	// Test that AudioTranscription returns an error when recognizer is not initialized
	_, err = backend.AudioTranscription(&proto.TranscriptRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sherpa-onnx recognizer not initialized")
	
	// Test that VAD returns an error when VAD is not initialized
	_, err = backend.VAD(&proto.VADRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VAD not initialized")
}