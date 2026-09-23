package txmgr

import (
	"context"
	"errors"
	"math/big"
)

type GasPriceEstimatorFn func(ctx context.Context, backend ETHBackend) (*big.Int, *big.Int, *big.Int, error)

func DefaultGasPriceEstimatorFn(ctx context.Context, backend ETHBackend) (*big.Int, *big.Int, *big.Int, error) {
	tip, err := backend.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	head, err := backend.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	if head.BaseFee == nil {
		return nil, nil, nil, errors.New("txmgr does not support pre-london blocks that do not have a base fee")
	}

	// pDai patch (PulseChain 369 is Shanghai-level: no EIP-4844 blob market).
	// Upstream asks the L1 for eth_blobBaseFee unconditionally and fails the WHOLE
	// fee estimate when that call fails. A pre-Cancun geth answers `null` (which
	// hexutil refuses to decode into a *hexutil.Big) and PulseChain's public RPCs
	// answer -32601 "method does not exist", so no transaction — batch, proposal or
	// challenge — could ever be sent to 369. The header we already hold says whether
	// a blob market exists: ExcessBlobGas is nil before Cancun. Only then is the
	// call skipped; a Cancun L1 keeps upstream behaviour exactly. nil is the value
	// SuggestGasPriceCaps documents for "4844 not yet active".
	var blobBaseFee *big.Int
	if head.ExcessBlobGas != nil {
		blobBaseFee, err = backend.BlobBaseFee(ctx)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return tip, head.BaseFee, blobBaseFee, nil
}
