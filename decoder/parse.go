package decoder

import (
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type TransferData struct {
	Sender common.Address
	To     common.Address
	Value  *big.Int
}

func TransferParse(topics []common.Hash, dataHex string) (*TransferData, error) {
	dataBytes, _ := hex.DecodeString(dataHex)
	// 构建一个类型
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}
	// 组装参数
	arguments := abi.Arguments{
		{
			Name:    "value",
			Type:    uint256Type,
			Indexed: false,
		},
	}
	unpacked, err := arguments.Unpack(dataBytes)
	if err != nil {
		return nil, err
	}
	value := unpacked[0].(*big.Int)

	return &TransferData{
		Sender: common.HexToAddress(topics[1].Hex()),
		To:     common.HexToAddress(topics[2].Hex()),
		Value:  value,
	}, nil
}
