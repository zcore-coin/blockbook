// +build unittest

package zcore

import (
	"blockbook/bchain"
	"blockbook/bchain/coins/btc"
	"bytes"
	"encoding/hex"
	"math/big"
	"os"
	"reflect"
	"testing"

	"github.com/martinboehm/btcutil/chaincfg"
)

var (
	testTx1, testTx2 bchain.Tx

	testTxPacked1 = "000088b88bd9c1e338010000000136d54c8ae74f4a6a675f88d2773ef388620ee90d5b1498c6bba19f77474d31c20100000048473044022020e61009263d983c88ff4a72c0a6bc30ff4d2647c7d98f66ec4c402c971b5e07022059aaeb73dcfa84c21956b74007ac75101fcc0c24eb767e202c1f927cc8383cde01ffffffff04000000000000000000002610ab3c00000023210290feb542136d3f0fb2c5a5f397262eb843c8527fed94349d51636969558bb558ac0065cd1d000000001976a91460c809c737cd39e019b092f3232036b0f84f6bf388ac80f0fa02000000001976a914bb1f665d18303a04492b15e5a53b556b88b4830d88ac00000000"
	testTxPacked2 = "00008bc18bd9c7c13e01000000010c3cf38f4020b2010ab5526a032c37993686f7a89fbfc840ce307909658825320100000049483045022100df16f40a896ac60108d8da08a483a6280a22aa19c9ee8566ec2122f3aedddcc702203bddc40d2780ec621618467d6a4acbe8f87d40cf25a0d5ef3cd74404fe671bfb01ffffffff0400000000000000000040c4ebb752000000232102f7a01fbbb2edbf7c1d460cd7c417cb79670ffc424e891de5ab87c9d730b7d32dac0065cd1d000000001976a914fbbbd120de101e92e76cb682063556b9d5a2988388ac80f0fa02000000001976a914bb1f665d18303a04492b15e5a53b556b88b4830d88ac00000000"
)

func init() {
	testTx1 = bchain.Tx{
		Hex:       "010000000136d54c8ae74f4a6a675f88d2773ef388620ee90d5b1498c6bba19f77474d31c20100000048473044022020e61009263d983c88ff4a72c0a6bc30ff4d2647c7d98f66ec4c402c971b5e07022059aaeb73dcfa84c21956b74007ac75101fcc0c24eb767e202c1f927cc8383cde01ffffffff04000000000000000000002610ab3c00000023210290feb542136d3f0fb2c5a5f397262eb843c8527fed94349d51636969558bb558ac0065cd1d000000001976a91460c809c737cd39e019b092f3232036b0f84f6bf388ac80f0fa02000000001976a914bb1f665d18303a04492b15e5a53b556b88b4830d88ac00000000",
		Blocktime: 1570257116,
		Txid:      "eeb64ce4df9df27dca13a9feac4b63d64ebeead9a01cd21146a8ae208f5d59e4",
		LockTime:  0,
                Version: 1,
		Vin: []bchain.Vin{
			{
				ScriptSig: bchain.ScriptSig{
					Hex: "473044022020e61009263d983c88ff4a72c0a6bc30ff4d2647c7d98f66ec4c402c971b5e07022059aaeb73dcfa84c21956b74007ac75101fcc0c24eb767e202c1f927cc8383cde01",
				},
				Txid:     "c2314d47779fa1bbc698145b0de90e6288f33e77d2885f676a4a4fe78a4cd536",
				Vout:     1,
				Sequence: 4294967295,
			},
		},
		Vout: []bchain.Vout{
			{
				ValueSat: *big.NewInt(0),
				N:        0,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "",
					Addresses: []string{},
				},
			},
			{
				ValueSat: *big.NewInt(260568000000),
				N:        1,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "210290feb542136d3f0fb2c5a5f397262eb843c8527fed94349d51636969558bb558ac",
					Addresses: []string{
						"zBX5j16Km6B5ZCHrjmoHWbrGAMTizUJtxr",
					},
				},
			},
			{
				ValueSat: *big.NewInt(500000000),
				N:        2,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "76a91460c809c737cd39e019b092f3232036b0f84f6bf388ac",
					Addresses: []string{
						"zHpPKjVgC5SyVfMdouGaAhrjQCZ6R2ZD4K",
					},
				},
			},
			{
				ValueSat: *big.NewInt(50000000),
				N:        3,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "76a914bb1f665d18303a04492b15e5a53b556b88b4830d88ac",
					Addresses: []string{
						"zS44nzYNkZUWfV1TVVgUqJTeHqSjuPjbsi",
					},
				},
                        },
		},
	}

	testTx2 = bchain.Tx{
		Hex:       "01000000010c3cf38f4020b2010ab5526a032c37993686f7a89fbfc840ce307909658825320100000049483045022100df16f40a896ac60108d8da08a483a6280a22aa19c9ee8566ec2122f3aedddcc702203bddc40d2780ec621618467d6a4acbe8f87d40cf25a0d5ef3cd74404fe671bfb01ffffffff0400000000000000000040c4ebb752000000232102f7a01fbbb2edbf7c1d460cd7c417cb79670ffc424e891de5ab87c9d730b7d32dac0065cd1d000000001976a914fbbbd120de101e92e76cb682063556b9d5a2988388ac80f0fa02000000001976a914bb1f665d18303a04492b15e5a53b556b88b4830d88ac00000000",
		Blocktime: 1570304095,
		Txid:      "cc5067a3ca36ea308a31d0dd979761ff59c42241596690da6c66a4aff3ca2cc0",
		LockTime:  0,
                Version: 1,
		Vin: []bchain.Vin{
			{
				ScriptSig: bchain.ScriptSig{
					Hex: "483045022100df16f40a896ac60108d8da08a483a6280a22aa19c9ee8566ec2122f3aedddcc702203bddc40d2780ec621618467d6a4acbe8f87d40cf25a0d5ef3cd74404fe671bfb01",
				},
				Txid:     "32258865097930ce40c8bf9fa8f7863699372c036a52b50a01b220408ff33c0c",
				Vout:     1,
				Sequence: 4294967295,
			},
		},
		Vout: []bchain.Vout{
			{
				ValueSat: *big.NewInt(0),
				N:        0,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "",
					Addresses: []string{},
				},
			},
			{
				ValueSat: *big.NewInt(355273000000),
				N:        1,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "2102f7a01fbbb2edbf7c1d460cd7c417cb79670ffc424e891de5ab87c9d730b7d32dac",
					Addresses: []string{
						"zXA9TABXDmuBYDsY1q4Nn51ciYQy15NCKY",
					},
				},
			},
			{
				ValueSat: *big.NewInt(500000000),
				N:        2,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "76a914fbbbd120de101e92e76cb682063556b9d5a2988388ac",
					Addresses: []string{
						"zXwhRKMJM6NJrd3NH7tFoG4GWa6aFrhA4b",
					},
				},
			},
			{
				ValueSat: *big.NewInt(50000000),
				N:        3,
				ScriptPubKey: bchain.ScriptPubKey{
					Hex: "76a914bb1f665d18303a04492b15e5a53b556b88b4830d88ac",
					Addresses: []string{
						"zS44nzYNkZUWfV1TVVgUqJTeHqSjuPjbsi",
					},
				},
			},
		},
	}
}

func TestMain(m *testing.M) {
	c := m.Run()
	chaincfg.ResetParams()
	os.Exit(c)
}

func TestGetAddrDesc(t *testing.T) {
	type args struct {
		tx     bchain.Tx
		parser *ZCoreParser
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "zcore-1",
			args: args{
				tx:     testTx1,
				parser: NewZCoreParser(GetChainParams("main"), &btc.Configuration{}),
			},
		},
		{
			name: "zcore-2",
			args: args{
				tx:     testTx2,
				parser: NewZCoreParser(GetChainParams("main"), &btc.Configuration{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for n, vout := range tt.args.tx.Vout {
				got1, err := tt.args.parser.GetAddrDescFromVout(&vout)
				if err != nil {
					t.Errorf("getAddrDescFromVout() error = %v, vout = %d", err, n)
					return
				}
                                if len(vout.ScriptPubKey.Addresses) == 0 {
                                	continue
                                }
				got2, err := tt.args.parser.GetAddrDescFromAddress(vout.ScriptPubKey.Addresses[0])
				if err != nil {
					t.Errorf("getAddrDescFromAddress() error = %v, vout = %d", err, n)
					return
				}
				if !bytes.Equal(got1, got2) {
					t.Errorf("Address descriptors mismatch: got1 = %v, got2 = %v", got1, got2)
				}
			}
		})
	}
}

func TestPackTx(t *testing.T) {
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
				height:    35000,
				blockTime: 1570257116,
				parser:    NewZCoreParser(GetChainParams("main"), &btc.Configuration{}),
			},
			want:    testTxPacked1,
			wantErr: false,
		},
		{
			name: "zcore-2",
			args: args{
				tx:        testTx2,
				height:    35777,
				blockTime: 1570304095,
				parser:    NewZCoreParser(GetChainParams("main"), &btc.Configuration{}),
			},
			want:    testTxPacked2,
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

func TestUnpackTx(t *testing.T) {
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
			want1:   35000,
			wantErr: false,
		},
		{
			name: "zcore-2",
			args: args{
				packedTx: testTxPacked2,
				parser:   NewZCoreParser(GetChainParams("main"), &btc.Configuration{}),
			},
			want:    &testTx2,
			want1:   35777,
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
