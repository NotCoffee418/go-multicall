package multicall

import (
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestMulticallRaw_Erc20TokenNames(t *testing.T) {
	type tokenValidation struct {
		address      common.Address
		expectedName string
	}

	partialERC20ABI := `[
		{
			"constant": true,
			"inputs": [],
			"name": "name",
			"outputs": [
				{
					"name": "",
					"type": "string"
				}
			],
			"type": "function"
		}
	]`

	// Load partial ERC20 ABI
	abi, err := abi.JSON(strings.NewReader(partialERC20ABI))
	if err != nil {
		t.Fatalf("failed to parse ABI: %v", err)
	}

	// Get ETH client with public RPC endpoint
	client, err := ethclient.Dial("https://polygon-rpc.com/")
	if err != nil {
		t.Fatalf("failed to dial client: %v", err)
	}

	// Get multicall config
	cfg := NewMulticallConfig(
		client,
		common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
		2, // batch size 2 to test batching
		30*time.Second,
	)

	// Token addresses to request the token name for
	tokenValidations := []tokenValidation{
		{common.HexToAddress("0x7ceB23fD6bC0adD59E62ac25578270cFf1b9f619"), "Wrapped Ether"},
		{common.HexToAddress("0xc2132D05D31c914a87C6611C10748AEb04B58e8F"), "(PoS) Tether USD"},
		{common.HexToAddress("0x3BA4c387f786bFEE076A58914F5Bd38d668B42c3"), "BNB (PoS)"},
		{common.HexToAddress("0xd93f7E271cB87c23AaA73edC008A79646d1F9912"), "Wrapped SOL"},
		{common.HexToAddress("0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359"), "USD Coin"},
	}

	// Pack multicall requests
	requests := make([]CallRequest, len(tokenValidations))
	for i, tv := range tokenValidations {
		packedCallData, err := abi.Pack("name")
		if err != nil {
			t.Fatalf("failed to pack call data: %v", err)
		}
		requests[i] = NewCallRequest(tv.address, packedCallData, true)
	}

	// Execute multicall
	results, err := MulticallRaw(cfg, requests)
	if err != nil {
		t.Fatalf("failed to execute multicall: %v", err)
	}

	// Validate results
	for i, got := range results {
		// Sanity checks
		if got.Data == nil {
			t.Fatalf("call %d returned nil result", i)
		}

		// Unpack the result
		gotTokenName, err := abi.Unpack("name", got.Data)
		if err != nil {
			t.Fatalf("call %d returned unexpected result: %v", i, err)
		}

		// Validate the result
		if gotTokenName[0] != tokenValidations[i].expectedName {
			t.Fatalf("call %d returned unexpected result: `%v`", i, gotTokenName[0])
		}
	}
}

func TestMulticallRaw_AllowFailure(t *testing.T) {
	// Setup
	client, err := ethclient.Dial("https://polygon-rpc.com/")
	if err != nil {
		t.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	cfg := NewMulticallConfig(
		client,
		common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
		2, // batch size
		30*time.Second,
	)

	// Create two calls: one that will succeed (balanceOf) and one that will fail (non-existent function)
	calls := []CallRequest{
		{
			Target:       common.HexToAddress("0x7ceB23fD6bC0adD59E62ac25578270cFf1b9f619"), // WETH address
			AllowFailure: false,
			CallData:     common.FromHex("0x70a08231000000000000000000000000000000000000000000000000000000000000000000"), // balanceOf(address(0))
		},
		{
			Target:       common.HexToAddress("0x7ceB23fD6bC0adD59E62ac25578270cFf1b9f619"), // WETH address again
			AllowFailure: true,
			CallData:     common.FromHex("0xdeadbeef"), // Non-existent function
		},
	}

	// Execute multicall
	results, err := MulticallRaw(cfg, calls)
	if err != nil {
		t.Fatalf("MulticallRaw failed: %v", err)
	}

	// Check results
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// First call should succeed
	if !results[0].Success {
		t.Errorf("Expected first call to succeed, but it failed")
	}
	if len(results[0].Data) == 0 {
		t.Errorf("Expected non-empty return data for first call")
	}

	// Second call should fail, but not cause the entire multicall to fail
	if results[1].Success {
		t.Errorf("Expected second call to fail, but it succeeded")
	}
	if len(results[1].Data) != 0 {
		t.Errorf("Expected empty return data for failed call, got %x", results[1].Data)
	}
}
