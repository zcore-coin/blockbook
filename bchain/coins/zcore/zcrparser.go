package zcore

import (
        "blockbook/bchain"
	"blockbook/bchain/coins/btc"
        "bytes"
        "fmt"
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
	return &ZCoreParser{BitcoinParser: btc.NewBitcoinParser(params, c)}
}

// ParseBlock parses raw block to our Block struct
func (p *ZCoreParser) ParseBlock(b []byte) (*bchain.Block, error) {
        w := wire.MsgBlock{}
        r := bytes.NewReader(b)

        if err := w.Deserialize(r); err != nil {
                return nil, err
        }
        fmt.Printf("Transactions:",w)
        txs := make([]bchain.Tx, len(w.Transactions))
        for ti, t := range w.Transactions {
                txs[ti] = p.TxFromMsgTx(t, false)
        }

        return &bchain.Block{
                BlockHeader: bchain.BlockHeader{
                        Size: len(b),
                        Time: w.Header.Timestamp.Unix(),
                },
                Txs: txs,
        }, nil
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
