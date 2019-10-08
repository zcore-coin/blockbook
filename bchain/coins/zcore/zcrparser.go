package zcore

import (
	"math"
	"math/big"
	"encoding/json"
	"blockbook/bchain"
	"blockbook/bchain/coins/btc"
	"github.com/martinboehm/btcd/wire"
	"github.com/martinboehm/btcutil/chaincfg"
)

// magic numbers
const (
	MainnetMagic wire.BitcoinNet = 0xcc645c66
	TestnetMagic wire.BitcoinNet = 0xcb618550
	RegtestMagic wire.BitcoinNet = 0x314527a9
)

// chain parameters
var (
	MainNetParams chaincfg.Params
	TestNetParams chaincfg.Params
	RegtestParams chaincfg.Params
)

func init() {
	MainNetParams = chaincfg.MainNetParams
	MainNetParams.Net = MainnetMagic

	// Address encoding magics
	MainNetParams.PubKeyHashAddrID = []byte{142}
	MainNetParams.ScriptHashAddrID = []byte{0}

	TestNetParams = chaincfg.TestNet3Params
	TestNetParams.Net = TestnetMagic

	// Address encoding magics
	TestNetParams.PubKeyHashAddrID = []byte{139} // base58 prefix: y
	TestNetParams.ScriptHashAddrID = []byte{19}  // base58 prefix: 8 or 9

	RegtestParams = chaincfg.RegressionNetParams
	RegtestParams.Net = RegtestMagic

	// Address encoding magics
	RegtestParams.PubKeyHashAddrID = []byte{139} // base58 prefix: y
	RegtestParams.ScriptHashAddrID = []byte{19}  // base58 prefix: 8 or 9
}

// ZCoreParser handle
type ZCoreParser struct {
	*btc.BitcoinParser
}

// NewZCoreParser returns new ZCoreParser instance
func NewZCoreParser(params *chaincfg.Params, c *btc.Configuration) *ZCoreParser {
	return &ZCoreParser{
		BitcoinParser: btc.NewBitcoinParser(params, c),
	}
}


// GetChainParams contains network parameters for the main ZCore network,
// the regression test ZCore network, the test ZCore network and
// the simulation test ZCore network, in this order
func GetChainParams(chain string) *chaincfg.Params {
	if !chaincfg.IsRegistered(&MainNetParams) {
		err := chaincfg.Register(&MainNetParams)
		if err == nil {
			err = chaincfg.Register(&TestNetParams)
		}
		if err == nil {
			err = chaincfg.Register(&RegtestParams)
		}
		if err != nil {
			panic(err)
		}
	}
	switch chain {
	case "test":
		return &TestNetParams
	case "regtest":
		return &RegtestParams
	default:
		return &MainNetParams
	}
}


func (p *ZCoreParser) ParseTxFromJson(jsonTx json.RawMessage) (*bchain.Tx, error) {
	var getTxResult GetTransactionResult
	if err := json.Unmarshal([]byte(jsonTx), &getTxResult.Result); err != nil {
		return nil, err
	}

	vins := make([]bchain.Vin, len(getTxResult.Result.Vin))
	for index, input := range getTxResult.Result.Vin {
		hexData := bchain.ScriptSig{}
		if input.ScriptSig != nil {
			hexData.Hex = input.ScriptSig.Hex
		}

		vins[index] = bchain.Vin{
			Coinbase:  input.Coinbase,
			Txid:      input.Txid,
			Vout:      input.Vout,
			ScriptSig: hexData,
			Sequence:  input.Sequence,
			// Addresses: []string{},
		}
	}

	vouts := make([]bchain.Vout, len(getTxResult.Result.Vout))
	for index, output := range getTxResult.Result.Vout {
		addr := output.ScriptPubKey.Addresses

		vouts[index] = bchain.Vout{
			ValueSat: *big.NewInt(int64(math.Round(output.Value * 1e8))),
			N:        output.N,
			ScriptPubKey: bchain.ScriptPubKey{
				Hex:       output.ScriptPubKey.Hex,
				Addresses: addr,
			},
		}
	}

	tx := &bchain.Tx{
		Hex:           getTxResult.Result.Hex,
		Txid:          getTxResult.Result.Txid,
		Version:       getTxResult.Result.Version,
		LockTime:      getTxResult.Result.LockTime,
		Vin:           vins,
		Vout:          vouts,
		Confirmations: uint32(getTxResult.Result.Confirmations),
		Time:          getTxResult.Result.Time,
		Blocktime:     getTxResult.Result.Blocktime,
	}

	return tx, nil
}

