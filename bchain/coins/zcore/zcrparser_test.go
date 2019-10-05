// +build unittest

package zcore

import (
	"blockbook/bchain"
	"blockbook/bchain/coins/btc"
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/martinboehm/btcutil/chaincfg"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type testBlock struct {
	size int
	time int64
	txs  []string
}

var testParseBlockTxs = map[int]testBlock{
	// Simple POS block
	20000: {
		size: 867,
		time: 1569345779,
		txs: []string{
			"d0febcb8eee13b5305c3ee205717fa8daa2667379283524fb8fbdae11f66d9d3",
			"a7b1f9801d820a450fb4971db5eb7b498172788817abeb67f6f2846677d31df7",
                        "053a37155f58dea638b136cbc30a4d23343afcebb94764a9b7e3b211a3f03073",
		},
	},
}

func TestMain(m *testing.M) {
	c := m.Run()
	chaincfg.ResetParams()
	os.Exit(c)
}

func helperLoadBlock(t *testing.T, height int) []byte {
	name := fmt.Sprintf("block_dump.%d", height)
	path := filepath.Join("testdata", name)

	d, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	d = bytes.TrimSpace(d)

	b := make([]byte, hex.DecodedLen(len(d)))
	_, err = hex.Decode(b, d)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

func TestParseBlock(t *testing.T) {
	p := NewZCoreParser(GetChainParams("main"), &btc.Configuration{})

	for height, tb := range testParseBlockTxs {
		b := helperLoadBlock(t, height)

		blk, err := p.ParseBlock(b)
		if err != nil {
			t.Errorf("ParseBlock() error %v", err)
		}

		if blk.Size != tb.size {
			t.Errorf("ParseBlock() block size: got %d, want %d", blk.Size, tb.size)
		}

		if blk.Time != tb.time {
			t.Errorf("ParseBlock() block time: got %d, want %d", blk.Time, tb.time)
		}
                fmt.Println(blk)
		if len(blk.Txs) != len(tb.txs) {
			t.Errorf("ParseBlock() number of transactions: got %d, want %d", len(blk.Txs), len(tb.txs))
		}

		for ti, tx := range tb.txs {
			if blk.Txs[ti].Txid != tx {
				t.Errorf("ParseBlock() transaction %d: got %s, want %s", ti, blk.Txs[ti].Txid, tx)
			}
		}
	}
}

func Test_GetAddrDescFromAddress_Mainnet(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "P2PKH1",
			args:    args{address: "P9hRjWq6tMqhroxswc2f5jp2ND2py8YEnu"},
			want:    "76a9140c26ca7967e6fe946f00bf81bcd3b86f43538edf88ac",
			wantErr: true,
		},
	}
	parser := NewZCoreParser(GetChainParams("main"), &btc.Configuration{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.GetAddrDescFromAddress(tt.args.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddrDescFromAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			h := hex.EncodeToString(got)
			if !reflect.DeepEqual(h, tt.want) {
				t.Errorf("GetAddrDescFromAddress() = %v, want %v", h, tt.want)
			}
		})
	}
}

func Test_GetAddressesFromAddrDesc(t *testing.T) {
	type args struct {
		script string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		want2   bool
		wantErr bool
	}{
		{
			name:    "P2PKH1",
			args:    args{script: "76a9140c26ca7967e6fe946f00bf81bcd3b86f43538edf88ac"},
			want:    []string{"P9hRjWq6tMqhroxswc2f5jp2ND2py8YEnu"},
			want2:   true,
			wantErr: true,
		},
	}

	parser := NewZCoreParser(GetChainParams("main"), &btc.Configuration{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := hex.DecodeString(tt.args.script)
			got, got2, err := parser.GetAddressesFromAddrDesc(b)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddressesFromAddrDesc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAddressesFromAddrDesc() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("GetAddressesFromAddrDesc() = %v, want %v", got2, tt.want2)
			}
		})
	}
}

var (
	testTx1       bchain.Tx
	testTxPacked1 = "0100000002be682a41b6af8c4749cb615040e0175bf5946c48f5b65916f4d113a5be4d14ae000000006a47304402203a18c935588165cbabd9574d6fea3f571e2bf0c215de491f29f938a39db0d5b20220689e8d189b5a08ef6bce5ef373a5fbc61ba92e6493343bc2bbe2b2089d818089012103dbb61883c633d95acca198b60362e40d88415ad539723bdf1b752a3ed9ee3e95ffffffff9c9dab1b8bf81c04b4082a0c9609cea1e57f437a6bb2af9f0947c1accf477328020000006a47304402205bd872a6ae17b908f42c940a0732855a4245ade7d118ea2b95b3322650a8a97b0220016a8f5c2f8d36975420dfd5d3d6f168d0c116e15ff86199cc5685bd78dd7c7f0121035574844d62aa130e5a5a4d96a6494deb049778c0bba25908efb7df07b3386eccffffffff020065cd1d000000001976a91456eee85de46eca1e029701ae23dd0c8b993d59ef88ac6cac3b00000000001976a914b12ca9f42d24b05e1df9f0801f4c48e20e8c943b88ac00000000"
)

func init() {
	testTx1 = bchain.Tx{
		Hex:       "0100000002be682a41b6af8c4749cb615040e0175bf5946c48f5b65916f4d113a5be4d14ae000000006a47304402203a18c935588165cbabd9574d6fea3f571e2bf0c215de491f29f938a39db0d5b20220689e8d189b5a08ef6bce5ef373a5fbc61ba92e6493343bc2bbe2b2089d818089012103dbb61883c633d95acca198b60362e40d88415ad539723bdf1b752a3ed9ee3e95ffffffff9c9dab1b8bf81c04b4082a0c9609cea1e57f437a6bb2af9f0947c1accf477328020000006a47304402205bd872a6ae17b908f42c940a0732855a4245ade7d118ea2b95b3322650a8a97b0220016a8f5c2f8d36975420dfd5d3d6f168d0c116e15ff86199cc5685bd78dd7c7f0121035574844d62aa130e5a5a4d96a6494deb049778c0bba25908efb7df07b3386eccffffffff020065cd1d000000001976a91456eee85de46eca1e029701ae23dd0c8b993d59ef88ac6cac3b00000000001976a914b12ca9f42d24b05e1df9f0801f4c48e20e8c943b88ac00000000",
		Blocktime: 1569345779,
		Txid:      "053a37155f58dea638b136cbc30a4d23343afcebb94764a9b7e3b211a3f03073",
		LockTime:  0,
		Version:   1,
		Vin: []bchain.Vin{
			{
				ScriptSig: bchain.ScriptSig{
					Hex: "47304402203a18c935588165cbabd9574d6fea3f571e2bf0c215de491f29f938a39db0d5b20220689e8d189b5a08ef6bce5ef373a5fbc61ba92e6493343bc2bbe2b2089d818089012103dbb61883c633d95acca198b60362e40d88415ad539723bdf1b752a3ed9ee3e95",
				},
				Txid:     "ae144dbea513d1f41659b6f5486c94f55b17e0405061cb49478cafb6412a68be",
				Vout:     0,
				Sequence: 4294967295,
			},
                        {
                                ScriptSig: bchain.ScriptSig{
                                        Hex: "47304402205bd872a6ae17b908f42c940a0732855a4245ade7d118ea2b95b3322650a8a97b0220016a8f5c2f8d36975420dfd5d3d6f168d0c116e15ff86199cc5685bd78dd7c7f0121035574844d62aa130e5a5a4d96a6494deb049778c0bba25908efb7df07b3386ecc",
                                },
                                Txid:     "287347cfacc147099fafb26b7a437fe5a1ce09960c2a08b4041cf88b1bab9d9c",
                                Vout:     2,
                                Sequence: 4294967295,
                        },
		},
		Vout: []bchain.Vout{
			{
				ValueSat: *big.NewInt(500000000),
				N:        0,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "76a91456eee85de46eca1e029701ae23dd0c8b993d59ef88ac",
					Addresses: []string{
						"zGvK8Wns9vhdXYLa3cEPEJqokyWqJqh3bL",
					},
				},
			},
			{
				ValueSat: *big.NewInt(3910764),
				N:        1,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "76a914b12ca9f42d24b05e1df9f0801f4c48e20e8c943b88ac",
					Addresses: []string{
						"zR9Tvg7T1RjrgE3xfiCGrNBXbLgefdRdH3",
					},
				},
			},
		},
	}
}

func Test_PackTx(t *testing.T) {
	type args struct {
		tx        bchain.Tx
		height    uint32
		blockTime int64
		parser    *ZCoreParser
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "zcore-1",
			args: args{
				tx:        testTx1,
				height:    20000,
				blockTime: 1569345779,
				parser:    NewZCoreParser(GetChainParams("main"), &btc.Configuration{}),
			},
			want:    testTxPacked1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.args.parser.PackTx(&tt.args.tx, tt.args.height, tt.args.blockTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("packTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			h := hex.EncodeToString(got)
			if !reflect.DeepEqual(h, tt.want) {
				t.Errorf("packTx() = %v, want %v", h, tt.want)
			}
		})
	}
}

func Test_UnpackTx(t *testing.T) {
	type args struct {
		packedTx string
		parser   *ZCoreParser
	}
	tests := []struct {
		name    string
		args    args
		want    *bchain.Tx
		want1   uint32
		wantErr bool
	}{
		{
			name: "zcore-1",
			args: args{
				packedTx: testTxPacked1,
				parser:   NewZCoreParser(GetChainParams("main"), &btc.Configuration{}),
			},
			want:    &testTx1,
			want1:   20000,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := hex.DecodeString(tt.args.packedTx)
			got, got1, err := tt.args.parser.UnpackTx(b)
			if (err != nil) != tt.wantErr {
				t.Errorf("unpackTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unpackTx() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("unpackTx() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
