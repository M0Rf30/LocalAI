package main

// This is a wrapper to satisfy the GRPC service interface
// It is meant to be used by the main executable that is the server for the specific backend type

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mudler/LocalAI/pkg/grpc/base"
	pb "github.com/mudler/LocalAI/pkg/grpc/proto"
	"github.com/mudler/LocalAI/pkg/utils"

	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
)

type SherpaONNX struct {
	base.SingleThread
	recognizer   *sherpa.OfflineRecognizer
	vad          *sherpa.VoiceActivityDetector
	sampleRate   int
}

func (s *SherpaONNX) Load(opts *pb.ModelOptions) error {
	// Set default sample rate
	s.sampleRate = 16000

	// Configure the offline recognizer
	config := sherpa.OfflineRecognizerConfig{}
	config.FeatConfig.SampleRate = s.sampleRate
	config.FeatConfig.FeatureDim = 80

	// Set model paths from opts
	modelDir := opts.Model
	if modelDir == "" {
		modelDir = filepath.Dir(opts.ModelFile)
	}

	// Try to detect model type based on available files
	encoderPath := filepath.Join(modelDir, "encoder.onnx")
	decoderPath := filepath.Join(modelDir, "decoder.onnx")
	joinerPath := filepath.Join(modelDir, "joiner.onnx")
	modelPath := filepath.Join(modelDir, "model.onnx")
	tokensPath := filepath.Join(modelDir, "tokens.txt")
	paraformerModelPath := filepath.Join(modelDir, "model.int8.onnx")
	whisperEncoderPath := filepath.Join(modelDir, "encoder.onnx")
	whisperDecoderPath := filepath.Join(modelDir, "decoder.onnx")

	// Configure based on available models
	if fileExists(encoderPath) && fileExists(decoderPath) && fileExists(joinerPath) {
		// Transducer model
		config.ModelConfig.Transducer.Encoder = encoderPath
		config.ModelConfig.Transducer.Decoder = decoderPath
		config.ModelConfig.Transducer.Joiner = joinerPath
	} else if fileExists(whisperEncoderPath) && fileExists(whisperDecoderPath) {
		// Whisper model
		config.ModelConfig.Whisper.Encoder = whisperEncoderPath
		config.ModelConfig.Whisper.Decoder = whisperDecoderPath
	} else if fileExists(paraformerModelPath) {
		// Paraformer model (int8 quantized)
		config.ModelConfig.Paraformer.Model = paraformerModelPath
	} else if fileExists(modelPath) {
		// Try different model types based on filename or directory structure
		if fileExists(tokensPath) {
			// Could be Paraformer, Whisper, or other model types
			// Check if it's a whisper model by looking at the directory name
			if strings.Contains(strings.ToLower(modelDir), "whisper") {
				// Assume it's a whisper model with separate encoder/decoder
				// But we only have one model file, so treat it as a paraformer
				config.ModelConfig.Paraformer.Model = modelPath
			} else {
				// Default to paraformer
				config.ModelConfig.Paraformer.Model = modelPath
			}
		} else {
			// Fallback to paraformer model path
			config.ModelConfig.Paraformer.Model = modelPath
		}
	}

	// Set tokens file
	if fileExists(tokensPath) {
		config.ModelConfig.Tokens = tokensPath
	}

	// Set other config options from opts if available
	if opts.Threads != 0 {
		config.ModelConfig.NumThreads = int(opts.Threads)
	} else {
		config.ModelConfig.NumThreads = 1
	}
	
	config.ModelConfig.Debug = 0
	
	// Set provider based on options or environment
	// Use Type field to determine provider if available
	if opts.Type != "" {
		config.ModelConfig.Provider = opts.Type
	} else {
		config.ModelConfig.Provider = "cpu"
	}
	
	// Set decoding method to greedy_search by default
	config.DecodingMethod = "greedy_search"

	// Create recognizer
	recognizer := sherpa.NewOfflineRecognizer(&config)
	if recognizer == nil {
		return fmt.Errorf("failed to create sherpa-onnx recognizer with config: %+v", config)
	}

	s.recognizer = recognizer

	// Initialize VAD if silero model exists
	sileroVadModel := filepath.Join(modelDir, "silero_vad.onnx")
	if fileExists(sileroVadModel) {
		vadConfig := sherpa.VadModelConfig{}
		vadConfig.SileroVad.Model = sileroVadModel
		vadConfig.SileroVad.Threshold = 0.5
		vadConfig.SileroVad.MinSilenceDuration = 0.5
		vadConfig.SileroVad.MinSpeechDuration = 0.25
		vadConfig.SileroVad.WindowSize = 512
		vadConfig.SampleRate = s.sampleRate
		vadConfig.NumThreads = config.ModelConfig.NumThreads
		vadConfig.Provider = config.ModelConfig.Provider
		vadConfig.Debug = 0

		vad := sherpa.NewVoiceActivityDetector(&vadConfig, 5.0)
		if vad != nil {
			s.vad = vad
		}
	}

	return nil
}

func (s *SherpaONNX) AudioTranscription(req *pb.TranscriptRequest) (pb.TranscriptResult, error) {
	if s.recognizer == nil {
		return pb.TranscriptResult{}, fmt.Errorf("sherpa-onnx recognizer not initialized")
	}

	// Create temporary directory for audio conversion
	dir, err := os.MkdirTemp("", "sherpa-onnx")
	if err != nil {
		return pb.TranscriptResult{}, err
	}
	defer os.RemoveAll(dir)

	// Convert audio to WAV format
	convertedPath := filepath.Join(dir, "converted.wav")
	if err := utils.AudioToWav(req.Dst, convertedPath); err != nil {
		return pb.TranscriptResult{}, err
	}

	// Read the wave file
	wave := sherpa.ReadWave(convertedPath)
	if wave == nil {
		return pb.TranscriptResult{}, fmt.Errorf("failed to read wave file: %s", convertedPath)
	}

	// Create stream and accept waveform
	stream := sherpa.NewOfflineStream(s.recognizer)
	if stream == nil {
		return pb.TranscriptResult{}, fmt.Errorf("failed to create offline stream")
	}
	defer sherpa.DeleteOfflineStream(stream)

	stream.AcceptWaveform(wave.SampleRate, wave.Samples)

	// Decode
	s.recognizer.Decode(stream)

	// Get result
	result := stream.GetResult()
	if result == nil {
		return pb.TranscriptResult{}, fmt.Errorf("failed to get recognition result")
	}

	// Convert to protobuf result
	segments := []*pb.TranscriptSegment{}
	
	// For now, we return the entire result as one segment
	// In the future, we could split this into multiple segments based on timestamps
	segment := &pb.TranscriptSegment{
		Id:     0,
		Text:   result.Text,
		Start:  0,
		End:    int64(float32(len(wave.Samples)) / float32(wave.SampleRate) * 1000), // Convert to milliseconds
		Tokens: make([]int32, 0), // Tokens not available in this API
	}
	
	segments = append(segments, segment)

	return pb.TranscriptResult{
		Segments: segments,
		Text:     result.Text,
	}, nil
}

func (s *SherpaONNX) VAD(req *pb.VADRequest) (pb.VADResponse, error) {
	if s.vad == nil {
		return pb.VADResponse{}, fmt.Errorf("VAD not initialized")
	}

	// Reset VAD
	s.vad.Reset()

	// Accept waveform
	s.vad.AcceptWaveform(req.Audio)

	// Flush to process remaining audio
	s.vad.Flush()

	// Collect segments
	segments := []*pb.VADSegment{}

	// Process all segments
	for !s.vad.IsEmpty() {
		speechSegment := s.vad.Front()
		if speechSegment == nil {
			break
		}
		s.vad.Pop()

		segment := &pb.VADSegment{
			Start: float32(speechSegment.Start) / float32(s.sampleRate),
			End:   float32(speechSegment.Start+len(speechSegment.Samples)) / float32(s.sampleRate),
		}
		segments = append(segments, segment)
	}

	return pb.VADResponse{
		Segments: segments,
	}, nil
}

func (s *SherpaONNX) Status() (pb.StatusResponse, error) {
	return pb.StatusResponse{
		State: pb.StatusResponse_READY,
	}, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}