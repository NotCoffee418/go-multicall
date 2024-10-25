package multicall

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

//go:embed multicall_abi.json
var multicallABIData []byte

var (
	multicallABI         abi.ABI
	initMulticallAbiOnce sync.Once
)

// Load multicall abi
func init() {
	initMulticallAbiOnce.Do(func() {
		if err := json.Unmarshal(multicallABIData, &multicallABI); err != nil {
			log.Fatalf("Failed to unmarshal Multicall ABI: %v. Please report this issue, this should never happen.", err)
		}
		multicallABIData = nil // no longer needed, free tiny amount of memory
	})
}

// MulticallRaw allows you to multicall directly with abi packed calls.
// Allows various functions on various contracts.
// Recommended function for most control and best performance.
//
// Parameters:
//   - cfg: The multicall config to use
//   - calls: The calls to execute.
//
// Returns:
//   - []CallResult: Packed results and success status of the calls
//   - error: An error if the multicall fails
func MulticallRaw(
	cfg *MulticallConfig,
	calls []CallRequest,
) ([]CallResult, error) {
	results := make([]CallResult, len(calls))

	// Create a context with a deadline
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel() // Ensure resources are cleaned up

	for i := 0; i < len(calls); i += cfg.BatchSize {
		batch := calls[i:min(i+cfg.BatchSize, len(calls))]
		callData, err := multicallABI.Pack("aggregate3", batch)
		if err != nil {
			return nil, fmt.Errorf("failed to pack multicall %d: %w", i, err)
		}

		callMsg := ethereum.CallMsg{
			To:   &cfg.MulticallAddress,
			Data: callData,
		}

		// Execute the multicall with the context
		result, err := cfg.Client.CallContract(ctx, callMsg, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to execute multicall %d: %w", i, err)
		}

		// Unpack the result
		var response []CallResult

		// Unpacking the response from the contract
		err = multicallABI.UnpackIntoInterface(&response, "aggregate3", result)
		if err != nil {
			return nil, fmt.Errorf("failed to unpack multicall %d: %w", i, err)
		}

		// Store the results as CallResultRaw
		for j := range response {
			results[i+j] = response[j]
		}
	}

	return results, nil
}
