// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/dotandev/hintents/internal/errors"
	"github.com/dotandev/hintents/internal/wat"
	"github.com/spf13/cobra"
)

var (
	wasmFileFlag       string
	wasmBase64Flag    string
	instructionTypeFlag string
	sortByFlag        string
)

var disassemblerCmd = &cobra.Command{
	Use:   "disassembler",
	Short: "WASM disassembler with gas cost analysis",
	Long: `Disassemble WebAssembly bytecode and analyze gas costs.

This command provides detailed analysis of WASM instructions including:
  - Instruction disassembly with gas costs
  - Average gas cost per instruction type
  - Gas cost statistics and breakdowns

The gas analysis helps identify expensive operations in smart contracts and optimize gas usage.

Examples:
  erst disassembler gas-analysis ./contract.wasm
  erst disassembler gas-analysis --wasm-base64 <base64-wasm>
  erst disassembler gas-analysis ./contract.wasm --sort-by total-cost
  erst disassembler gas-analysis ./contract.wasm --instruction-type i32.add`,
}

var gasAnalysisCmd = &cobra.Command{
	Use:   "gas-analysis",
	Short: "Analyze gas costs per WASM instruction type",
	Long: `Perform comprehensive gas cost analysis on WASM bytecode.

This command disassembles the WASM module and provides detailed statistics about gas usage,
including average gas cost per instruction type, total costs, and instruction frequency.

The analysis helps identify:
  - Most expensive instruction types
  - Instruction frequency distribution
  - Overall gas consumption patterns

Output includes:
  - Total instructions and gas cost
  - Average gas cost per instruction type
  - Top instructions by total gas cost
  - Percentage breakdown of gas usage`,
	Args: cobra.ExactArgs(0),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if wasmFileFlag == "" && wasmBase64Flag == "" {
			return errors.WrapCliArgumentRequired("either --wasm-file or --wasm-base64 is required")
		}
		if wasmFileFlag != "" && wasmBase64Flag != "" {
			return errors.WrapValidationError("cannot specify both --wasm-file and --wasm-base64")
		}
		return nil
	},
	RunE: runGasAnalysis,
}

func init() {
	gasAnalysisCmd.Flags().StringVar(&wasmFileFlag, "wasm-file", "", "Path to WASM file to analyze")
	gasAnalysisCmd.Flags().StringVar(&wasmBase64Flag, "wasm-base64", "", "Base64-encoded WASM bytecode to analyze")
	gasAnalysisCmd.Flags().StringVar(&instructionTypeFlag, "instruction-type", "", "Filter analysis to specific instruction type (e.g., i32.add, call)")
	gasAnalysisCmd.Flags().StringVar(&sortByFlag, "sort-by", "total-cost", "Sort instructions by: total-cost, average-cost, count")
	disassemblerCmd.AddCommand(gasAnalysisCmd)
	rootCmd.AddCommand(disassemblerCmd)
}

func runGasAnalysis(cmd *cobra.Command, args []string) error {
	var wasmBytes []byte
	var err error

	if wasmFileFlag != "" {
		wasmBytes, err = os.ReadFile(wasmFileFlag)
		if err != nil {
			return errors.WrapValidationError(fmt.Sprintf("failed to read WASM file: %v", err))
		}
		fmt.Printf("Analyzing WASM file: %s\n", wasmFileFlag)
	} else {
		wasmBytes, err = base64.StdEncoding.DecodeString(wasmBase64Flag)
		if err != nil {
			return errors.WrapValidationError(fmt.Sprintf("failed to decode base64 WASM: %v", err))
		}
		fmt.Printf("Analyzing base64-encoded WASM (%d bytes)\n", len(wasmBytes))
	}

	// Create disassembler
	disassembler := wat.NewDisassembler(wasmBytes)
	if !disassembler.IsValidWasm() {
		return errors.WrapValidationError("invalid WASM module")
	}

	fmt.Printf("WASM module validated successfully\n\n")

	// Perform gas analysis
	analysis, err := disassembler.AnalyzeGasCosts()
	if err != nil {
		return fmt.Errorf("gas analysis failed: %w", err)
	}

	// Filter by instruction type if specified
	if instructionTypeFlag != "" {
		if _, exists := analysis.InstructionCounts[instructionTypeFlag]; !exists {
			return errors.WrapValidationError(fmt.Sprintf("instruction type '%s' not found in WASM module", instructionTypeFlag))
		}
		fmt.Printf("Analysis filtered for instruction type: %s\n\n", instructionTypeFlag)
		fmt.Printf("Count: %d\n", analysis.InstructionCounts[instructionTypeFlag])
		fmt.Printf("Total Gas Cost: %d particles (%.4f gas units)\n", 
			analysis.InstructionGasCosts[instructionTypeFlag], 
			float64(analysis.InstructionGasCosts[instructionTypeFlag])/10000.0)
		fmt.Printf("Average Gas Cost: %d particles (%.4f gas units)\n",
			analysis.AverageGasCostByInstruction[instructionTypeFlag],
			float64(analysis.AverageGasCostByInstruction[instructionTypeFlag])/10000.0)
		return nil
	}

	// Validate sort-by flag
	validSortOptions := map[string]bool{
		"total-cost":  true,
		"average-cost": true,
		"count":       true,
	}
	if !validSortOptions[sortByFlag] {
		return errors.WrapValidationError(fmt.Sprintf("invalid sort-by option: %s. Valid options: total-cost, average-cost, count", sortByFlag))
	}

	// Display full analysis
	fmt.Println(analysis.FormatWithCustomSorting(sortByFlag))
	return nil
}
