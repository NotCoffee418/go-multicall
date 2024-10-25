package multicall

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const defaultBatchSize = 100
const defaultTimeout = 30 * time.Second

// CallRequest is a packed call to be executed with multicall.
// ABI Definition: Call3
//
// Parameters:
//   - Target: The address of the contract to call
//   - AllowFailure: Whether the call is allowed to fail
//   - CallData: packed calldata to the target contract with abi.Pack
type CallRequest struct {
	Target       common.Address
	AllowFailure bool
	CallData     []byte
}

// CallResult is the packedresult of a packed call executed with multicall.
// ABI Definition: Call3Value
//
// Parameters:
//   - Data: Packed result of the call. Unpacked with abi.Unpack
//   - Success: Whether the call was successful
type CallResult struct {
	Success bool
	Data    []byte
}

// MulticallConfig is a reusable multicall config.
//
// Parameters:
//   - Client: The Ethereum client to use
//   - MulticallAddress: The address of the multicall contract for the chain.
//     You can find the deployment address for your chain at https://www.multicall3.com/deployments
//   - BatchSize: The maximum number of calls to include in a single multicall
//     Generally it's cheaper and faster to pack as many calls as possible in a single multicall.
//     Default: 100
//   - Timeout: The timeout for the multicall.
//     Default: 30 seconds
type MulticallConfig struct {
	Client           *ethclient.Client
	MulticallAddress common.Address
	BatchSize        int
	Timeout          time.Duration
}

// Create a new reusable multicall config
//
// Parameters:
//   - client: The Ethereum client to use
//   - multicallAddress: The address of the multicall contract for the chain.
//     You can find the deployment address for your chain at https://www.multicall3.com/deployments
//   - batchSize: The maximum number of calls to include in a single multicall
//     Generally it's cheaper and faster to pack as many calls as possible in a single multicall.
//     Default: 100
func NewMulticallConfig(
	client *ethclient.Client,
	multicallAddress common.Address,
	batchSize int,
	timeout time.Duration,
) *MulticallConfig {
	return &MulticallConfig{
		Client:           client,
		MulticallAddress: multicallAddress,
		BatchSize:        batchSize,
		Timeout:          timeout,
	}
}

// Create a new reusable multicall config
//
// Parameters:
//   - client: The Ethereum client to use
//   - multicallAddress: The address of the multicall contract for the chain.
//     You can find the deployment address for your chain at https://www.multicall3.com/deployments
func NewDefaultMulticallConfig(client *ethclient.Client, multicallAddress common.Address) *MulticallConfig {
	return NewMulticallConfig(client, multicallAddress, defaultBatchSize, defaultTimeout)
}

func NewCallRequest(
	target common.Address,
	packedCallData []byte,
	allowFailure bool,
) CallRequest {
	return CallRequest{target, allowFailure, packedCallData}
}
