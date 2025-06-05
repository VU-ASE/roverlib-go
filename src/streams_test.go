package roverlib

import (

	"testing"

)

// A simple helper function to create a small Service with a single input and output stream.
// Dummy values are used for the addresses and names, which can be adjusted as needed.
func sampleServiceStream() Service {
	outputName := "testOutput"
	outputAddress := "tcp://localhost:5555"

	inputService := "testService"
	inputName := "testInput"
	inputAddress := "tcp://unix:6666"

	inputs := []Input{{Service: &inputService, Streams: []Stream{{Name: &inputName, Address: &inputAddress}}}}
	outputs := []Output{{Name: &outputName, Address: &outputAddress}}
	
	return Service{
		Inputs:  inputs,
		Outputs: outputs,
	}
}

// Tests whether GetWriteStream works correctly for an existing stream.
// It checks that it returns the correct address,
// and that it returns the same instance for the same stream name.
func TestGetWriteStreamHappy(t *testing.T) {
	service := sampleServiceStream()

	// Get the write stream for the existing output
	write_stream := service.GetWriteStream("testOutput")
	if write_stream == nil {
		t.Fatalf("GetWriteStream returned nil for existing stream")
	}

	// Check the address
	if write_stream.stream.address != "tcp://*:5555" {
		t.Fatalf("Expected address tcp://*:5555, got %s", write_stream.stream.address)
	}

	// test that the stream is a singleton
	write_stream2 := service.GetWriteStream("testOutput")
	if write_stream != write_stream2 {
		t.Fatalf("Expected GetWriteStream to return the same instance for the same name, got different instances")
	}
}

// Same as for GetWriteStream, but for GetReadStream.
func TestGetReadStreamHappy(t *testing.T) {
	service := sampleServiceStream()

	read_stream := service.GetReadStream("testService", "testInput")
	if read_stream == nil {
		t.Fatalf("GetReadStream returned nil for existing stream")
	}

	if read_stream.stream.address != "tcp://unix:6666" {
		t.Fatalf("Expected address tcp://unix:6666, got %s", read_stream.stream.address)
	}

	// test that the stream again is a singleton
	read_stream2 := service.GetReadStream("testService", "testInput")
	if read_stream != read_stream2 {
		t.Fatalf("Expected GetReadStream to return the same instance for the same name, got different instances")
	}
}

// Tests that GetWriteStream returns nil for a non-existent stream.
func TestGetwriteStreamMissing(t *testing.T) {
	service := sampleServiceStream()

	// Try to get a write stream that does not exist
	output_stream := service.GetWriteStream("nonExistentOutput")
	if output_stream != nil {
		t.Fatalf("GetWriteStream should return nil for non-existent stream, got %v", output_stream)
	}
}

// Tests that GetReadStream returns nil for a non-existent stream.
func TestGetReadStreamMissing(t *testing.T) {
	service := sampleServiceStream()

	input_stream := service.GetReadStream("nonExistentService", "nonExistentInput")
	if input_stream != nil {
		t.Fatalf("GetReadStream should return nil for non-existent stream, got %v", input_stream)
	}
}