# go-multicall

A Go library for efficient Ethereum multicalls using the Multicall3 contract.


## Features

- Efficient batching of multiple contract calls
- Support for different Ethereum networks
- Configurable batch size and timeout
- Error handling and failed call detection
- Allow failure of calls using Aggregate3

## Installation

To install go-multicall, use the following command:

```bash
go get github.com/NotCoffee418/go-multicall
```

## Usage

Here's a basic example of how to use go-multicall to fetch multiple ERC20 token names in a single call:

```go
package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/NotCoffee418/go-multicall"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// Connect to an Ethereum node
	client, err := ethclient.Dial("https://polygon-rpc.com/")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

    // Multicall3 for your your chain
    // Find it here: https://www.multicall3.com/deployments
    multicallAddress := common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11")

	// Create a multicall configuration
	cfg := multicall.NewMulticallConfig(
		client,
		multicallAddress, 
		100,  // Batch size
		30*time.Second, // Timeout
	)

	// Load the target contract abi
	erc20ABI, err := abi.JSON("[]") // Placeholder
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}

	// Define token addresses to query
	tokenAddresses := []common.Address{
		common.HexToAddress("0x7ceB23fD6bC0adD59E62ac25578270cFf1b9f619"), // WETH
		common.HexToAddress("0xc2132D05D31c914a87C6611C10748AEb04B58e8F"), // USDT
		common.HexToAddress("0x3BA4c387f786bFEE076A58914F5Bd38d668B42c3"), // BNB
	}

	// Prepare multicall requests
	requests := make([]multicall.CallRequest, len(tokenAddresses))
	for i, addr := range tokenAddresses {
		callData, err := erc20ABI.Pack("name")
		if err != nil {
			log.Fatalf("Failed to pack call data: %v", err)
		}
		requests[i] = multicall.NewCallRequest(addr, callData, true)
	}

	// Execute multicall
	results, err := multicall.MulticallRaw(cfg, requests)
	if err != nil {
		log.Fatalf("Multicall failed: %v", err)
	}

	// Process results
	for i, result := range results {
		if result.Success {
			name, err := erc20ABI.Unpack("name", result.Data)
			if err != nil {
				log.Printf("Failed to unpack result for %s: %v", tokenAddresses[i].Hex(), err)
				continue
			}
			fmt.Printf("Token at %s has name: %s\n", tokenAddresses[i].Hex(), name[0].(string))
		} else {
			fmt.Printf("Call failed for token at %s\n", tokenAddresses[i].Hex())
		}
	}
}
```

This example demonstrates how to:

1. Connect to an Ethereum node (Polygon in this case)
2. Create a multicall configuration
3. Prepare ERC20 'name' function calls for multiple token addresses
4. Execute the multicall
5. Process and display the results

## Dependencies

- [go-ethereum](https://github.com/ethereum/go-ethereum)

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
