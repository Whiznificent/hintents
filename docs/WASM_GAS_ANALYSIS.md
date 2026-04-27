# WASM Gas Analysis

This document describes the WASM gas analysis functionality added in issue #1221.

## Overview

The WASM gas analysis feature provides detailed insights into gas consumption patterns of WebAssembly smart contracts. It helps developers identify expensive operations and optimize gas usage by analyzing instruction frequency and costs.

## Features

### Gas Cost Mapping
- Comprehensive gas cost mapping for all WASM instruction types
- Costs based on ewasm documentation and typical execution patterns
- Precision: 1/10000 of a gas unit (particles)

### Analysis Capabilities
- **Total gas cost analysis**: Overall gas consumption
- **Average gas cost per instruction type**: Identifies expensive operations
- **Instruction frequency**: Shows most common operations
- **Percentage breakdown**: Gas usage distribution
- **Custom sorting**: Sort by total cost, average cost, or count

## CLI Usage

### Basic Gas Analysis

```bash
# Analyze a WASM file
erst disassembler gas-analysis ./contract.wasm

# Analyze base64-encoded WASM
erst disassembler gas-analysis --wasm-base64 <base64-wasm>
```

### Advanced Options

```bash
# Sort by average gas cost
erst disassembler gas-analysis ./contract.wasm --sort-by average-cost

# Sort by instruction count
erst disassembler gas-analysis ./contract.wasm --sort-by count

# Filter by specific instruction type
erst disassembler gas-analysis ./contract.wasm --instruction-type i32.add
```

### Supported Sort Options

- `total-cost` (default): Sort instructions by total gas consumption
- `average-cost`: Sort by average gas cost per instruction
- `count`: Sort by frequency of instruction execution

## Output Format

### Full Analysis Output

```
Gas Cost Analysis
================
Total Instructions: 150
Total Gas Cost: 450 particles (0.0450 gas units)
Average Gas Cost: 3 particles (0.0003 gas units)

Top Instructions by Total Gas Cost:
Instruction           Count    Total Cost    Avg Cost     % of Total
----------           -----    -----------    ---------    ----------
i32.const            45       45            1             10.00%
i32.add              30       30            1             6.67%
local.get            25       50            2             11.11%
call                 10       50            5             11.11%
i32.mul              5        15            3             3.33%
```

### Filtered Instruction Output

```
Analysis filtered for instruction type: i32.add
Count: 30
Total Gas Cost: 30 particles (0.0030 gas units)
Average Gas Cost: 1 particles (0.0001 gas units)
```

## Gas Cost Reference

### Control Flow Instructions
- `unreachable`, `nop`, `block`, `loop`, `if`, `else`, `end`: 1 particle
- `br`: 2 particles
- `br_if`: 3 particles
- `br_table`: 4 particles
- `return`: 1 particle
- `call`: 5 particles
- `call_indirect`: 10 particles

### Variable Access
- `local.get`, `local.set`, `local.tee`: 2 particles
- `global.get`, `global.set`: 3 particles

### Memory Operations
- `i32.load`, `i64.load`, `f32.load`, `f64.load`: 3 particles
- `i32.store`, `i64.store`, `f32.store`, `f64.store`: 3 particles
- `memory.size`: 5 particles
- `memory.grow`: 10 particles

### Constants
- `i32.const`, `i64.const`, `f32.const`, `f64.const`: 1 particle

### Arithmetic Operations
- **i32**: `add`, `sub`, `and`, `or`, `xor`, `shl`, `shr_s`, `shr_u`: 1 particle
- **i32**: `mul`: 3 particles, `div_s`, `div_u`, `rem_s`, `rem_u`: 5 particles
- **i32**: `clz`, `ctz`: 10 particles, `popcnt`: 8 particles
- **i32**: `rotl`, `rotr`: 2 particles

- **i64**: `add`, `sub`, `and`, `or`, `xor`, `shl`, `shr_s`, `shr_u`: 1 particle
- **i64**: `mul`: 5 particles, `div_s`, `div_u`, `rem_s`, `rem_u`: 7 particles
- **i64**: `clz`, `ctz`: 12 particles, `popcnt`: 10 particles
- **i64**: `rotl`, `rotr`: 2 particles

### Floating Point Operations
- **f32/f64**: `add`, `sub`, `mul`, `min`, `max`: 3 particles
- **f32/f64**: `div`, `sqrt`, `ceil`, `floor`, `trunc`, `nearest`: 5 particles
- **f32/f64**: `abs`, `neg`: 2 particles

### Conversions
- `i32.wrap_i64`: 1 particle
- `i64.extend_i32_s`, `i64.extend_i32_u`: 2 particles
- Type conversions involving floats: 3 particles
- Truncations: 5 particles
- Reinterpretations: 2 particles

### Miscellaneous
- `drop`: 1 particle
- `select`: 3 particles

## Implementation Details

### Core Components

1. **GasCosts Map**: Comprehensive mapping of instruction mnemonics to gas costs
2. **GasCostAnalysis Struct**: Analysis results with statistics
3. **FormatWithCustomSorting**: Flexible output formatting
4. **CLI Integration**: Command-line interface for easy access

### API Usage

```go
// Create disassembler
disassembler := wat.NewDisassembler(wasmBytes)

// Perform gas analysis
analysis, err := disassembler.AnalyzeGasCosts()
if err != nil {
    return err
}

// Format with custom sorting
output := analysis.FormatWithCustomSorting("average-cost")
fmt.Println(output)

// Access specific data
totalCost := analysis.TotalGasCost
avgCost := analysis.AverageGasCostByInstruction["i32.add"]
count := analysis.InstructionCounts["call"]
```

## Use Cases

### 1. Gas Optimization
Identify the most expensive operations in your contract:
```bash
erst disassembler gas-analysis contract.wasm --sort-by average-cost
```

### 2. Performance Profiling
Understand instruction frequency and patterns:
```bash
erst disassembler gas-analysis contract.wasm --sort-by count
```

### 3. Cost Analysis
Find operations that consume the most gas overall:
```bash
erst disassembler gas-analysis contract.wasm --sort-by total-cost
```

### 4. Targeted Optimization
Analyze specific instruction types:
```bash
erst disassembler gas-analysis contract.wasm --instruction-type call
```

## Testing

The implementation includes comprehensive tests:

- Unit tests for gas cost mapping
- Integration tests for analysis functionality
- CLI command tests with various options
- WASM module validation tests

Run tests:
```bash
go test ./internal/wat -v
go test ./cmd -v
```

## Examples

### Sample Contract Analysis

For a simple addition contract:
```wat
(module
  (func $add (param i32 i32) (result i32)
    local.get 0
    local.get 1
    i32.add)
  (export "add" (func $add)))
```

Analysis output:
```
Gas Cost Analysis
================
Total Instructions: 3
Total Gas Cost: 5 particles (0.0005 gas units)
Average Gas Cost: 1 particles (0.0001 gas units)

Top Instructions by Total Gas Cost:
Instruction           Count    Total Cost    Avg Cost     % of Total
----------           -----    -----------    ---------    ----------
local.get            2        4             2             80.00%
i32.add             1        1             1             20.00%
```

## Future Enhancements

Potential improvements for future versions:

1. **Gas Optimization Suggestions**: Automated recommendations for gas savings
2. **Comparison Mode**: Compare gas usage between different contract versions
3. **Visualization**: Charts and graphs for gas consumption patterns
4. **Custom Gas Models**: Support for user-defined gas cost models
5. **Integration with Debug Mode**: Gas analysis in transaction debugging

## Contributing

When contributing to the gas analysis feature:

1. Update gas costs based on empirical data
2. Add tests for new instruction types
3. Maintain backward compatibility
4. Update documentation for new features

## Related Issues

- #1221: [WASM] Map Gas Costs to WASM Instructions
- Gas optimization and performance tracking
- Contract analysis tools
