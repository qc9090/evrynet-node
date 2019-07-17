package tests

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestCreateContractWithProviderAddress(t *testing.T) {
	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)
	sender := common.HexToAddress(senderAddrStr)
	provideraddr := common.HexToAddress(providerAddrStr)
	payLoadBytes, err := hexutil.Decode(payload)
	assert.NoError(t, err)

	ethClient, err := ethclient.Dial(ethRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := ethClient.NonceAt(context.Background(), sender, nil)
	assert.NoError(t, err)
	tx := types.NewContractCreation(nonce, big.NewInt(0), testGasLimit, big.NewInt(testGasPrice), payLoadBytes, &provideraddr)
	tx, err = types.SignTx(tx, types.HomesteadSigner{}, spk)
	assert.NoError(t, err)
	assert.NoError(t, ethClient.SendTransaction(context.Background(), tx))

}

func TestCreateContractWithoutProviderAddress(t *testing.T) {
	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)
	sender := common.HexToAddress(senderAddrStr)
	payLoadBytes, err := hexutil.Decode(payload)
	assert.NoError(t, err)

	ethClient, err := ethclient.Dial(ethRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := ethClient.NonceAt(context.Background(), sender, nil)
	assert.NoError(t, err)
	tx := types.NewContractCreation(nonce, big.NewInt(0), testGasLimit, big.NewInt(testGasPrice), payLoadBytes, nil)
	tx, err = types.SignTx(tx, types.HomesteadSigner{}, spk)
	assert.NoError(t, err)
	assert.NoError(t, ethClient.SendTransaction(context.Background(), tx))

}

func TestCreateContractWithProviderSignature(t *testing.T) {
	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)
	ppk, err := crypto.HexToECDSA(providerPK)
	assert.NoError(t, err)
	sender := common.HexToAddress(senderAddrStr)
	payLoadBytes, err := hexutil.Decode(payload)
	assert.NoError(t, err)

	ethClient, err := ethclient.Dial(ethRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := ethClient.NonceAt(context.Background(), sender, nil)
	assert.NoError(t, err)
	tx := types.NewContractCreation(nonce, big.NewInt(0), testGasLimit, big.NewInt(testGasPrice), payLoadBytes, nil)
	tx, err = types.SignTx(tx, types.HomesteadSigner{}, spk)
	assert.NoError(t, err)
	tx, err = types.ProviderSignTx(tx, types.HomesteadSigner{}, ppk)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, ethClient.SendTransaction(context.Background(), tx))
}

func TestCreateContractWithProviderAddressMustHaveOwnerAddress(t *testing.T) {
	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)
	sender := common.HexToAddress(senderAddrStr)
	provideraddr := common.HexToAddress(providerAddrStr)
	payLoadBytes, err := hexutil.Decode(payload)
	assert.NoError(t, err)

	ethClient, err := ethclient.Dial(ethRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := ethClient.NonceAt(context.Background(), sender, nil)
	assert.NoError(t, err)
	tx := types.NewContractCreation(nonce, big.NewInt(0), testGasLimit, big.NewInt(testGasPrice), payLoadBytes, &sender, &provideraddr)
	tx, err = types.SignTx(tx, types.HomesteadSigner{}, spk)
	assert.NoError(t, err)
	assert.Equal(t, senderAddrStr, tx.Owner().Hex())
}

func TestCreateNormalContractMustHaveNoOwnerAndProviderAddress(t *testing.T) {
	spk, err := crypto.HexToECDSA(senderPK)
	assert.NoError(t, err)
	sender := common.HexToAddress(senderAddrStr)
	payLoadBytes, err := hexutil.Decode(payload)
	assert.NoError(t, err)

	ethClient, err := ethclient.Dial(ethRPCEndpoint)
	assert.NoError(t, err)
	nonce, err := ethClient.NonceAt(context.Background(), sender, nil)
	assert.NoError(t, err)
	tx := types.NewContractCreation(nonce, big.NewInt(0), testGasLimit, big.NewInt(testGasPrice), payLoadBytes)
	tx, err = types.SignTx(tx, types.HomesteadSigner{}, spk)
	assert.NoError(t, err)
	assert.Nil(t, tx.Owner())
	assert.Nil(t, tx.Provider())
}
