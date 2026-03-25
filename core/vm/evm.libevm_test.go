// Copyright 2024-2025 the libevm authors.
//
// The libevm additions to go-ethereum are free software: you can redistribute
// them and/or modify them under the terms of the GNU Lesser General Public License
// as published by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// The libevm additions are distributed in the hope that they will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser
// General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see
// <http://www.gnu.org/licenses/>.
package vm

import (
	"math/big"
	"testing"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/params"
)

type evmArgOverrider struct {
	newEVMchainID int64

	gotResetHook     bool
	resetTxContextTo TxContext
	resetStateDBTo   StateDB
}

func (o *evmArgOverrider) OverrideNewEVMArgs(args *NewEVMArgs) *NewEVMArgs {
	args.ChainConfig = &params.ChainConfig{ChainID: big.NewInt(o.newEVMchainID)}
	return args
}

func (o *evmArgOverrider) OverrideEVMResetArgs(_ params.Rules, _ *EVMResetArgs) *EVMResetArgs {
	o.gotResetHook = true
	return &EVMResetArgs{
		TxContext: o.resetTxContextTo,
		StateDB:   o.resetStateDBTo,
	}
}

func (o *evmArgOverrider) register(t *testing.T) {
	t.Helper()
	TestOnlyClearRegisteredHooks()
	RegisterHooks(o)
	t.Cleanup(TestOnlyClearRegisteredHooks)
}

func TestOverrideNewEVMArgs(t *testing.T) {
	// The overrideNewEVMArgs function accepts and returns all arguments to
	// NewEVM(), in order. Here we lock in our assumption of that order. If this
	// breaks then all functionality overriding the args MUST be updated.
	var _ func(BlockContext, StateDB, *params.ChainConfig, Config) *EVM = NewEVM

	const chainID = 13579
	hooks := evmArgOverrider{newEVMchainID: chainID}
	hooks.register(t)

	assertChainID := func(t *testing.T, want int64) {
		t.Helper()
		evm := NewEVM(BlockContext{}, nil, nil, Config{})
		got := evm.ChainConfig().ChainID
		require.Equalf(t, big.NewInt(want), got, "%T.ChainConfig().ChainID set by NewEVM() hook", evm)
	}
	assertChainID(t, chainID)

	t.Run("WithTempRegisteredHooks", func(t *testing.T) {
		err := libevm.WithTemporaryExtrasLock(func(lock libevm.ExtrasLock) error {
			override := evmArgOverrider{newEVMchainID: 24680}
			return WithTempRegisteredHooks(lock, &override, func() error {
				assertChainID(t, override.newEVMchainID)
				return nil
			})
		})
		require.NoError(t, err)
		t.Run("after", func(t *testing.T) {
			assertChainID(t, chainID)
		})
	})
}

func TestOverrideEVMResetArgs(t *testing.T) {
	const (
		chainID  = 0xc0ffee
		gasPrice = 1357924680
	)
	hooks := &evmArgOverrider{
		newEVMchainID: chainID,
		resetTxContextTo: TxContext{
			GasPrice: uint256.NewInt(gasPrice),
		},
	}
	hooks.register(t)

	evm := NewEVM(BlockContext{}, nil, &params.ChainConfig{ChainID: big.NewInt(chainID)}, Config{})
	hooks.gotResetHook = false
	evm.SetTxContext(TxContext{})
	assert.Truef(t, hooks.gotResetHook, "SetTxContext should invoke OverrideEVMResetArgs")
	assert.Equalf(t, uint256.NewInt(gasPrice), evm.GasPrice, "%T.GasPrice set by SetTxContext() hook", evm)
}
