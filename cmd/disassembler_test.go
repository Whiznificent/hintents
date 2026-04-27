// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDisassemblerCommand(t *testing.T) {
	// Create a simple WASM module for testing
	// Function body: i32.const 1, i32.const 2, i32.add, drop
	body := []byte{
		0x41, 0x01, // i32.const 1
		0x41, 0x02, // i32.const 2
		0x6a,       // i32.add
		0x1a,       // drop
	}
	wasm := buildMinimalWasm(body)
	wasmBase64 := base64.StdEncoding.EncodeToString(wasm)

	tests := []struct {
		name           string
		args           []string
		flags          map[string]string
		expectedOutput []string
		expectError    bool
	}{
		{
			name: "gas analysis with base64 wasm",
			args: []string{"gas-analysis"},
			flags: map[string]string{
				"wasm-base64": wasmBase64,
			},
			expectedOutput: []string{
				"Analyzing base64-encoded WASM",
				"WASM module validated successfully",
				"Gas Cost Analysis",
				"Total Instructions:",
				"Total Gas Cost:",
				"Average Gas Cost:",
				"Top Instructions by Total Gas Cost:",
				"i32.const",
				"i32.add",
				"drop",
			},
			expectError: false,
		},
		{
			name: "gas analysis with sorting by average cost",
			args: []string{"gas-analysis"},
			flags: map[string]string{
				"wasm-base64": wasmBase64,
				"sort-by":      "average-cost",
			},
			expectedOutput: []string{
				"Top Instructions by Average Gas Cost:",
			},
			expectError: false,
		},
		{
			name: "gas analysis with sorting by count",
			args: []string{"gas-analysis"},
			flags: map[string]string{
				"wasm-base64": wasmBase64,
				"sort-by":      "count",
			},
			expectedOutput: []string{
				"Top Instructions by Count:",
			},
			expectError: false,
		},
		{
			name: "gas analysis filtered by instruction type",
			args: []string{"gas-analysis"},
			flags: map[string]string{
				"wasm-base64":     wasmBase64,
				"instruction-type": "i32.add",
			},
			expectedOutput: []string{
				"Analysis filtered for instruction type: i32.add",
				"Count:",
				"Total Gas Cost:",
				"Average Gas Cost:",
			},
			expectError: false,
		},
		{
			name:        "error when no wasm provided",
			args:        []string{"gas-analysis"},
			flags:       map[string]string{},
			expectError: true,
		},
		{
			name: "error when both wasm file and base64 provided",
			args: []string{"gas-analysis"},
			flags: map[string]string{
				"wasm-file":    "test.wasm",
				"wasm-base64": wasmBase64,
			},
			expectError: true,
		},
		{
			name: "error with invalid sort option",
			args: []string{"gas-analysis"},
			flags: map[string]string{
				"wasm-base64": wasmBase64,
				"sort-by":     "invalid",
			},
			expectError: true,
		},
		{
			name: "error with invalid instruction type",
			args: []string{"gas-analysis"},
			flags: map[string]string{
				"wasm-base64":     wasmBase64,
				"instruction-type": "invalid.instruction",
			},
			expectError: true,
		},
		{
			name: "error with invalid base64",
			args: []string{"gas-analysis"},
			flags: map[string]string{
				"wasm-base64": "invalid-base64",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary buffer for output
			var buf bytes.Buffer
			disassemblerCmd.SetOut(&buf)

			// Set flags
			for flag, value := range tt.flags {
				disassemblerCmd.Flags().Set(flag, value)
			}

			// Execute command
			err := disassemblerCmd.ParseAndRun(tt.args)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Check expected output
			output := buf.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain '%s', but got: %s", expected, output)
				}
			}
		})
	}
}

// buildMinimalWasm creates a minimal WASM module with the given function body
// This is a simplified version that creates a valid WASM module for testing
func buildMinimalWasm(body []byte) []byte {
	// This is a minimal WASM module structure
	// Magic number + version
	wasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	
	// Type section (one function type: [] -> [])
	wasm = append(wasm, 0x01) // section count
	wasm = append(wasm, 0x60) // section ID (type)
	wasm = append(wasm, 0x00) // param count
	wasm = append(wasm, 0x00) // result count
	wasm = append(wasm, 0x04) // section size
	
	// Function section (one function at index 0)
	wasm = append(wasm, 0x03) // section ID (function)
	wasm = append(wasm, 0x02) // section size
	wasm = append(wasm, 0x01) // count
	wasm = append(wasm, 0x00) // type index
	
	// Export section (export function as "main")
	wasm = append(wasm, 0x07) // section ID (export)
	wasm = append(wasm, 0x07) // section size
	wasm = append(wasm, 0x01) // count
	wasm = append(wasm, 0x04) // name length
	wasm = append(wasm, 'm', 'a', 'i', 'n') // name
	wasm = append(wasm, 0x00) // export kind (function)
	wasm = append(wasm, 0x00) // function index
	
	// Code section
	wasm = append(wasm, 0x0a) // section ID (code)
	
	// Function body
	funcBody := append([]byte{0x00}, body...) // 0 locals
	funcBody = append(funcBody, 0x0b)         // end opcode
	
	bodySize := len(funcBody)
	if bodySize < 128 {
		wasm = append(wasm, 0x01) // count
		wasm = append(wasm, byte(bodySize)) // body size (single byte)
	} else {
		// Handle larger bodies with LEB128 encoding
		wasm = append(wasm, 0x01) // count
		for bodySize >= 128 {
			wasm = append(wasm, byte(bodySize|0x80))
			bodySize >>= 7
		}
		wasm = append(wasm, byte(bodySize))
	}
	wasm = append(wasm, funcBody...) // actual body code
	
	return wasm
}
