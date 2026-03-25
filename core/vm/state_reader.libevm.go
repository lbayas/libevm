// Copyright 2025 the libevm authors.
//
// The libevm additions to go-ethereum are free software: you can redistribute
// them and/or modify them under the terms of the GNU Lesser General Public License
// as published by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.

package vm

import (
	"github.com/ava-labs/libevm/common"
	"github.com/ava-labs/libevm/libevm"
	"github.com/ava-labs/libevm/libevm/stateconf"
	"github.com/holiman/uint256"
)

var _ libevm.StateReader = stateDBLibevmReader{}

// AsLibevmStateReader returns s as a [libevm.StateReader], or wraps it when the
// concrete type does not implement that interface (e.g. some test doubles).
func AsLibevmStateReader(s StateDB) libevm.StateReader {
	if r, ok := s.(libevm.StateReader); ok {
		return r
	}
	return stateDBLibevmReader{s}
}

type stateDBLibevmReader struct {
	s StateDB
}

func (r stateDBLibevmReader) GetBalance(addr common.Address) *uint256.Int {
	return r.s.GetBalance(addr)
}

func (r stateDBLibevmReader) GetNonce(addr common.Address) uint64 {
	return r.s.GetNonce(addr)
}

func (r stateDBLibevmReader) GetCodeHash(addr common.Address) common.Hash {
	return r.s.GetCodeHash(addr)
}

func (r stateDBLibevmReader) GetCode(addr common.Address) []byte {
	return r.s.GetCode(addr)
}

func (r stateDBLibevmReader) GetCodeSize(addr common.Address) int {
	return r.s.GetCodeSize(addr)
}

func (r stateDBLibevmReader) GetRefund() uint64 {
	return r.s.GetRefund()
}

func (r stateDBLibevmReader) GetCommittedState(addr common.Address, h common.Hash, opts ...stateconf.StateDBStateOption) common.Hash {
	type withCommitted interface {
		GetCommittedState(common.Address, common.Hash, ...stateconf.StateDBStateOption) common.Hash
	}
	if x, ok := r.s.(withCommitted); ok {
		return x.GetCommittedState(addr, h, opts...)
	}
	_, c := r.s.GetStateAndCommittedState(addr, h)
	return c
}

func (r stateDBLibevmReader) GetState(addr common.Address, h common.Hash, opts ...stateconf.StateDBStateOption) common.Hash {
	type withOpts interface {
		GetState(common.Address, common.Hash, ...stateconf.StateDBStateOption) common.Hash
	}
	if x, ok := r.s.(withOpts); ok {
		return x.GetState(addr, h, opts...)
	}
	return r.s.GetState(addr, h)
}

func (r stateDBLibevmReader) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	return r.s.GetTransientState(addr, key)
}

func (r stateDBLibevmReader) HasSelfDestructed(addr common.Address) bool {
	return r.s.HasSelfDestructed(addr)
}

func (r stateDBLibevmReader) Exist(addr common.Address) bool {
	return r.s.Exist(addr)
}

func (r stateDBLibevmReader) Empty(addr common.Address) bool {
	return r.s.Empty(addr)
}

func (r stateDBLibevmReader) AddressInAccessList(addr common.Address) bool {
	return r.s.AddressInAccessList(addr)
}

func (r stateDBLibevmReader) SlotInAccessList(addr common.Address, slot common.Hash) (bool, bool) {
	return r.s.SlotInAccessList(addr, slot)
}
