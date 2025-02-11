// Copyright 2021 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

//go:build integration
// +build integration

package modules

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/ChainSafe/gossamer/dot/state"
	"github.com/ChainSafe/gossamer/dot/types"
	"github.com/ChainSafe/gossamer/internal/log"
	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/lib/crypto/ed25519"
	"github.com/ChainSafe/gossamer/lib/crypto/sr25519"
	"github.com/ChainSafe/gossamer/lib/keystore"
	"github.com/ChainSafe/gossamer/lib/runtime"
	"github.com/ChainSafe/gossamer/lib/runtime/wasmer"
	"github.com/ChainSafe/gossamer/lib/transaction"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

// https://github.com/paritytech/substrate/blob/5420de3face1349a97eb954ae71c5b0b940c31de/core/transaction-pool/src/tests.rs#L95
var testExt = common.MustHexToBytes("0x410284ffd43593c715fdd31c61141abd04a99fd6822c8558854ccde39a5684e7a56da27d01f8e" +
	"fbe48487e57a22abf7e3acd491b7f3528a33a111b1298601554863d27eb129eaa4e718e1365414ff3d028b62bebc651194c6b5001e5c2839b98" +
	"2757e08a8c0000000600ff8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a480b00c465f14670")

// invalid transaction (above tx, with last byte changed)
//nolint
var testInvalidExt = []byte{1, 212, 53, 147, 199, 21, 253, 211, 28, 97, 20, 26, 189, 4, 169, 159, 214, 130, 44, 133, 88, 133, 76, 205, 227, 154, 86, 132, 231, 165, 109, 162, 125, 142, 175, 4, 21, 22, 135, 115, 99, 38, 201, 254, 161, 126, 37, 252, 82, 135, 97, 54, 147, 201, 18, 144, 156, 178, 38, 170, 71, 148, 242, 106, 72, 69, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 216, 5, 113, 87, 87, 40, 221, 120, 247, 252, 137, 201, 74, 231, 222, 101, 85, 108, 102, 39, 31, 190, 210, 14, 215, 124, 19, 160, 180, 203, 54, 110, 167, 163, 149, 45, 12, 108, 80, 221, 65, 238, 57, 237, 199, 16, 10, 33, 185, 8, 244, 184, 243, 139, 5, 87, 252, 245, 24, 225, 37, 154, 163, 143}

func TestMain(m *testing.M) {
	wasmFilePaths, err := runtime.GenerateRuntimeWasmFile()
	if err != nil {
		log.Errorf("failed to generate runtime wasm file: %s", err)
		os.Exit(1)
	}

	// Start all tests
	code := m.Run()

	runtime.RemoveFiles(wasmFilePaths)
	os.Exit(code)
}

func TestAuthorModule_Pending(t *testing.T) {
	ctrl := gomock.NewController(t)
	telemetryMock := NewMockClient(ctrl)
	telemetryMock.
		EXPECT().
		SendMessage(gomock.Any()).
		AnyTimes()

	txQueue := state.NewTransactionState(telemetryMock)
	auth := NewAuthorModule(log.New(log.SetWriter(io.Discard)), nil, txQueue)

	res := new(PendingExtrinsicsResponse)
	err := auth.PendingExtrinsics(nil, nil, res)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(*res, PendingExtrinsicsResponse([]string{})) {
		t.Errorf("Fail: expected: %+v got: %+v\n", *res, PendingExtrinsicsResponse([]string{}))
	}

	vtx := &transaction.ValidTransaction{
		Extrinsic: types.NewExtrinsic(testExt),
		Validity:  new(transaction.Validity),
	}

	_, err = txQueue.Push(vtx)
	require.NoError(t, err)

	err = auth.PendingExtrinsics(nil, nil, res)
	if err != nil {
		t.Fatal(err)
	}

	expected := common.BytesToHex(vtx.Extrinsic)
	if !reflect.DeepEqual(*res, PendingExtrinsicsResponse([]string{expected})) {
		t.Errorf("Fail: expected: %+v got: %+v\n", res, PendingExtrinsicsResponse([]string{expected}))
	}
}

func TestAuthorModule_SubmitExtrinsic_Integration(t *testing.T) {
	t.Skip()

	ctrl := gomock.NewController(t)
	telemetryMock := NewMockClient(ctrl)
	telemetryMock.
		EXPECT().
		SendMessage(gomock.Any()).
		AnyTimes()

	// setup auth module
	txQueue := state.NewTransactionState(telemetryMock)

	auth := setupAuthModule(t, txQueue)

	// create and submit extrinsic
	ext := Extrinsic{fmt.Sprintf("0x%x", testExt)}

	res := new(ExtrinsicHashResponse)

	err := auth.SubmitExtrinsic(nil, &ext, res)
	require.Nil(t, err)

	// setup expected results
	val := &transaction.Validity{
		Priority:  69,
		Requires:  [][]byte{},
		Provides:  [][]byte{{146, 157, 61, 99, 63, 98, 30, 242, 128, 49, 150, 90, 140, 165, 187, 249}},
		Longevity: 64,
		Propagate: true,
	}
	expected := &transaction.ValidTransaction{
		Extrinsic: types.NewExtrinsic(testExt),
		Validity:  val,
	}
	expectedHash := ExtrinsicHashResponse("0xb20777f4db60ea55b1aeedde2d7b7aff3efeda736b7e2a840b5713348f766078")

	inQueue := txQueue.Pop()

	// compare results
	require.Equal(t, expected, inQueue)
	require.Equal(t, expectedHash, *res)
}

func TestAuthorModule_SubmitExtrinsic_invalid(t *testing.T) {
	t.Skip()

	ctrl := gomock.NewController(t)
	telemetryMock := NewMockClient(ctrl)
	telemetryMock.
		EXPECT().
		SendMessage(gomock.Any()).
		AnyTimes()

	// setup service
	// setup auth module
	txQueue := state.NewTransactionState(telemetryMock)
	auth := setupAuthModule(t, txQueue)

	// create and submit extrinsic
	ext := Extrinsic{fmt.Sprintf("0x%x", testInvalidExt)}

	res := new(ExtrinsicHashResponse)

	err := auth.SubmitExtrinsic(nil, &ext, res)
	require.EqualError(t, err, runtime.ErrInvalidTransaction.Message)
}

func TestAuthorModule_SubmitExtrinsic_invalid_input(t *testing.T) {
	ctrl := gomock.NewController(t)
	telemetryMock := NewMockClient(ctrl)
	telemetryMock.
		EXPECT().
		SendMessage(gomock.Any()).
		AnyTimes()

	// setup service
	// setup auth module
	txQueue := state.NewTransactionState(telemetryMock)
	auth := setupAuthModule(t, txQueue)

	// create and submit extrinsic
	ext := Extrinsic{fmt.Sprintf("%x", "1")}

	res := new(ExtrinsicHashResponse)
	err := auth.SubmitExtrinsic(nil, &ext, res)
	require.EqualError(t, err, "could not byteify non 0x prefixed string: 31")
}

func TestAuthorModule_SubmitExtrinsic_InQueue(t *testing.T) {
	t.Skip()

	ctrl := gomock.NewController(t)
	telemetryMock := NewMockClient(ctrl)
	telemetryMock.
		EXPECT().
		SendMessage(gomock.Any()).
		AnyTimes()

	// setup auth module
	txQueue := state.NewTransactionState(telemetryMock)

	auth := setupAuthModule(t, txQueue)

	// create and submit extrinsic
	ext := Extrinsic{fmt.Sprintf("0x%x", testExt)}

	res := new(ExtrinsicHashResponse)

	// setup expected results
	val := &transaction.Validity{
		Priority:  69,
		Requires:  [][]byte{},
		Provides:  [][]byte{{146, 157, 61, 99, 63, 98, 30, 242, 128, 49, 150, 90, 140, 165, 187, 249}},
		Longevity: 64,
		Propagate: true,
	}
	expected := &transaction.ValidTransaction{
		Extrinsic: types.NewExtrinsic(testExt),
		Validity:  val,
	}

	_, err := txQueue.Push(expected)
	require.Nil(t, err)

	// this should cause error since transaction is already in txQueue
	err = auth.SubmitExtrinsic(nil, &ext, res)
	require.EqualError(t, err, transaction.ErrTransactionExists.Error())

}

func TestAuthorModule_InsertKey_Valid(t *testing.T) {
	seed := "0xb7e9185065667390d2ad952a5324e8c365c9bf503dcf97c67a5ce861afe97309"
	kp, err := sr25519.NewKeypairFromSeed(common.MustHexToBytes(seed))
	require.NoError(t, err)

	auth := setupAuthModule(t, nil)
	req := &KeyInsertRequest{"babe", seed, kp.Public().Hex()}
	res := &KeyInsertResponse{}
	err = auth.InsertKey(nil, req, res)
	require.Nil(t, err)
	require.Len(t, *res, 0) // zero len result on success
}

func TestAuthorModule_InsertKey_Valid_Gran_Keytype(t *testing.T) {
	seed := "0xb7e9185065667390d2ad952a5324e8c365c9bf503dcf97c67a5ce861afe97309"
	kp, err := ed25519.NewKeypairFromSeed(common.MustHexToBytes(seed))
	require.NoError(t, err)

	auth := setupAuthModule(t, nil)
	req := &KeyInsertRequest{"gran", seed, kp.Public().Hex()}
	res := &KeyInsertResponse{}
	err = auth.InsertKey(nil, req, res)
	require.Nil(t, err)

	require.Len(t, *res, 0) // zero len result on success
}

func TestAuthorModule_InsertKey_InValid(t *testing.T) {
	auth := setupAuthModule(t, nil)
	req := &KeyInsertRequest{
		"babe",
		"0xb7e9185065667390d2ad952a5324e8c365c9bf503dcf97c67a5ce861afe97309",
		"0x0000000000000000000000000000000000000000000000000000000000000000"}
	res := &KeyInsertResponse{}
	err := auth.InsertKey(nil, req, res)
	require.EqualError(t, err, "generated public key does not equal provide public key")
}

func TestAuthorModule_InsertKey_UnknownKeyType(t *testing.T) {
	auth := setupAuthModule(t, nil)
	req := &KeyInsertRequest{"mack",
		"0xb7e9185065667390d2ad952a5324e8c365c9bf503dcf97c67a5ce861afe97309",
		"0x6246ddf254e0b4b4e7dffefc8adf69d212b98ac2b579c362b473fec8c40b4c0a"}
	res := &KeyInsertResponse{}
	err := auth.InsertKey(nil, req, res)
	require.EqualError(t, err, "cannot decode key: invalid key type")

}

func TestAuthorModule_HasKey_Integration(t *testing.T) {
	auth := setupAuthModule(t, nil)
	kr, err := keystore.NewSr25519Keyring()
	require.Nil(t, err)

	var res bool
	req := []string{kr.Alice().Public().Hex(), "acco"}
	err = auth.HasKey(nil, &req, &res)
	require.NoError(t, err)
	require.True(t, res)
}

func TestAuthorModule_HasKey_NotFound(t *testing.T) {
	auth := setupAuthModule(t, nil)
	kr, err := keystore.NewSr25519Keyring()
	require.Nil(t, err)

	var res bool
	req := []string{kr.Bob().Public().Hex(), "babe"}
	err = auth.HasKey(nil, &req, &res)
	require.NoError(t, err)
	require.False(t, res)
}

func TestAuthorModule_HasKey_InvalidKey(t *testing.T) {
	auth := setupAuthModule(t, nil)

	var res bool
	req := []string{"0xaa11", "babe"}
	err := auth.HasKey(nil, &req, &res)
	require.EqualError(t, err, "cannot create public key: input is not 32 bytes")
	require.False(t, res)
}

func TestAuthorModule_HasKey_InvalidKeyType(t *testing.T) {
	auth := setupAuthModule(t, nil)
	kr, err := keystore.NewSr25519Keyring()
	require.Nil(t, err)

	var res bool
	req := []string{kr.Alice().Public().Hex(), "xxxx"}
	err = auth.HasKey(nil, &req, &res)
	require.EqualError(t, err, "invalid keystore name")
	require.False(t, res)
}

func setupAuthModule(t *testing.T, txq *state.TransactionState) *AuthorModule {
	cs := newCoreService(t, nil)
	rt := wasmer.NewTestInstance(t, runtime.NODE_RUNTIME)
	t.Cleanup(func() {
		rt.Stop()
	})
	return NewAuthorModule(log.New(log.SetWriter(io.Discard)), cs, txq)
}
