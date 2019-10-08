package zcore

import (
	"blockbook/bchain"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"blockbook/bchain/coins/btc"
	"github.com/decred/dcrd/dcrjson"
	"github.com/golang/glog"
	"github.com/juju/errors"
)

// ZCoreRPC is an interface to JSON-RPC bitcoind service.
type ZCoreRPC struct {
	*btc.BitcoinRPC
	mtx         sync.Mutex
	client      http.Client
	rpcURL      string
	rpcUser     string
	bestBlock   uint32
	rpcPassword string
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type GenericCmd struct {
	ID     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params,omitempty"`
}

type GetBestBlockHashResult struct {
	Error  Error    `json:"error"`
	Result string   `json:"result"`
}

type GetBestBlockResult struct {
	Error  Error `json:"error"`
	Result struct {
		Hash   string `json:"hash"`
		Height uint32 `json:"height"`
	} `json:"result"`
}

type GetBlockResult struct {
	Error  Error `json:"error"`
	Result struct {
		Hash          string      `json:"hash"`
		Confirmations int64       `json:"confirmations"`
		Size          int32       `json:"size"`
		Height        uint32      `json:"height"`
		Version       json.Number `json:"version"`
		MerkleRoot    string      `json:"merkleroot"`
                AccCheckpoint string      `json:"acc_checkpoint"`
		Tx            []string    `json:"tx,omitempty"`
		Time          int64       `json:"time"`
		Nonce         json.Number `json:"nonce"`
		Bits          string      `json:"bits"`
		Difficulty    float64     `json:"difficulty"`
		ChainWork     string      `json:"chainwork"`
		PreviousHash  string      `json:"previousblockhash"`
		NextHash      string      `json:"nextblockhash,omitempty"`
	} `json:"result"`
}

type GetBlockHeaderResult struct {
	Error  Error `json:"error"`
	Result struct {
		Hash          string      `json:"hash"`
		Confirmations int64       `json:"confirmations"`
                Height        uint32      `json:"height"`
		Version       json.Number `json:"version"`
		MerkleRoot    string      `json:"merkleroot"`
		Bits          string      `json:"bits"`
		Time          int64       `json:"time"`
		Nonce         uint32      `json:"nonce"`
		Difficulty    float64     `json:"difficulty"`
		ChainWork     string      `json:"chainwork"`
                AccCheckpoint string      `json:"acc_checkpoint"`
		PreviousHash  string      `json:"previousblockhash,omitempty"`
		NextHash      string      `json:"nextblockhash,omitempty"`
	} `json:"result"`
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

type Vin struct {
	Coinbase    string     `json:"coinbase"`
	Txid        string     `json:"txid"`
	Vout        uint32     `json:"vout"`
	Sequence    uint32     `json:"sequence"`
	ScriptSig   *ScriptSig `json:"scriptsig"`
}

type ScriptPubKeyResult struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex,omitempty"`
	ReqSigs   int32    `json:"reqSigs,omitempty"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses,omitempty"`
}

type Vout struct {
	Value        float64            `json:"value"`
	N            uint32             `json:"n"`
	ScriptPubKey ScriptPubKeyResult `json:"scriptPubKey"`
}

type RawTx struct {
	Hex           string `json:"hex"`
	Txid          string `json:"txid"`
	Version       int32  `json:"version"`
	LockTime      uint32 `json:"locktime"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	BlockHash     string `json:"blockhash"`
	Confirmations int64  `json:"confirmations,omitempty"`
	Time          int64  `json:"time,omitempty"`
	Blocktime     int64  `json:"blocktime,omitempty"`
}

type MempoolTxsResult struct {
	Error  Error    `json:"error"`
	Result []string `json:"result"`
}

type GetTransactionResult struct {
	Error  Error `json:"error"`
	Result struct {
		RawTx
	} `json:"result"`
}

type SendRawTransactionResult struct {
	Error  Error  `json:"error"`
	Result string `json:"result"`
}


type GetBlockHashResult struct {
	Error  Error  `json:"error"`
	Result string `json:"result"`
}


type DecodeRawTransactionResult struct {
	Error  Error `json:"error"`
	Result struct {
		Txid     string `json:"txid"`
		Version  int32  `json:"version"`
		Locktime uint32 `json:"locktime"`
		Vin      []Vin  `json:"vin"`
		Vout     []Vout `json:"vout"`
	} `json:"result"`
}


type GetInfoChainResult struct {
	Error  Error `json:"error"`
	Result struct {
		Version         int32   `json:"version"`
		ProtocolVersion int32   `json:"protocolversion"`
		Blocks          int64   `json:"blocks"`
		TimeOffset      int64   `json:"timeoffset"`
		Connections     int32   `json:"connections"`
		Proxy           string  `json:"proxy"`
		Difficulty      float64 `json:"difficulty"`
		TestNet         bool    `json:"testnet"`
		RelayFee        float64 `json:"relayfee"`
		Errors          string  `json:"errors"`
	}
}


type GetBlockChainInfoResult struct {
	Error  Error `json:"error"`
	Result struct {
		Chain                string  `json:"chain"`
		Blocks               int64   `json:"blocks"`
		Headers              int64   `json:"headers"`
		BestBlockHash        string  `json:"bestblockhash"`
		Difficulty           float64  `json:"difficulty"`
		VerificationProgress float64 `json:"verificationprogress"`
		ChainWork            string  `json:"chainwork"`
		InitialBlockDownload bool    `json:"initialblockdownload"`
		MaxBlockSize         int64   `json:"maxblocksize"`
	} `json:"result"`
}

// NewZCoreRPC returns new ZCoreRPC instance.
func NewZCoreRPC(config json.RawMessage, pushHandler func(bchain.NotificationType)) (bchain.BlockChain, error) {
	b, err := btc.NewBitcoinRPC(config, pushHandler)
	if err != nil {
		return nil, err
	}

	var c btc.Configuration
	if err = json.Unmarshal(config, &c); err != nil {
		return nil, errors.Annotate(err, "Invalid configuration file")
	}

	transport := &http.Transport{
		Dial:                (&net.Dialer{KeepAlive: 600 * time.Second}).Dial,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100, // necessary to not to deplete ports
	}

	s := &ZCoreRPC{
		BitcoinRPC:  b.(*btc.BitcoinRPC),
		client:      http.Client{Timeout: time.Duration(c.RPCTimeout) * time.Second, Transport: transport},
		rpcURL:      c.RPCURL,
		rpcUser:     c.RPCUser,
		rpcPassword: c.RPCPass,
	}

	s.BitcoinRPC.RPCMarshaler = btc.JSONMarshalerV1{}
	s.BitcoinRPC.ChainConfig.SupportsEstimateSmartFee = false

	return s, nil
}

// Initialize initializes ZCoreRPC instance.
func (b *ZCoreRPC) Initialize() error {
	ci, err := b.GetChainInfo()
	if err != nil {
		return err
	}
	chainName := ci.Chain

	params := GetChainParams(chainName)

	// always create parser
	b.BitcoinRPC.Parser = NewZCoreParser(params, b.ChainConfig)

	// parameters for getInfo request
	if params.Net == MainnetMagic {
		b.BitcoinRPC.Testnet = false
		b.BitcoinRPC.Network = "livenet"
	} else {
		b.BitcoinRPC.Testnet = true
		b.BitcoinRPC.Network = "testnet"
	}

	glog.Info("rpc: block chain ", params.Name)

	return nil
}

func (d *ZCoreRPC) GetChainInfo() (*bchain.ChainInfo, error) {
	blockchainInfoRequest := GenericCmd{
		ID:     1,
		Method: "getblockchaininfo",
	}

	var blockchainInfoResult GetBlockChainInfoResult
	if err := d.Call(blockchainInfoRequest, &blockchainInfoResult); err != nil {
		return nil, err
	}

	if blockchainInfoResult.Error.Message != "" {
		return nil, mapToStandardErr("Error fetching blockchain info: %s", blockchainInfoResult.Error)
	}

	infoChainRequest := GenericCmd{
		ID:     2,
		Method: "getinfo",
	}

	var infoChainResult GetInfoChainResult
	if err := d.Call(infoChainRequest, &infoChainResult); err != nil {
		return nil, err
	}

	if infoChainResult.Error.Message != "" {
		return nil, mapToStandardErr("Error fetching network info: %s", infoChainResult.Error)
	}

	chainInfo := &bchain.ChainInfo{
		Chain:           blockchainInfoResult.Result.Chain,
		Blocks:          int(blockchainInfoResult.Result.Blocks),
		Headers:         int(blockchainInfoResult.Result.Headers),
		Bestblockhash:   blockchainInfoResult.Result.BestBlockHash,
		Difficulty:      strconv.Itoa(int(blockchainInfoResult.Result.Difficulty)),
		Version:         strconv.Itoa(int(infoChainResult.Result.Version)),
		Subversion:      "",
		ProtocolVersion: strconv.Itoa(int(infoChainResult.Result.ProtocolVersion)),
		Timeoffset:      float64(infoChainResult.Result.TimeOffset),
		Warnings:        "",
	}
	return chainInfo, nil
}


func (d *ZCoreRPC) getBlockHashByHeight(height uint32) (*GetBlockHashResult, error) {
	blockHashRequest := GenericCmd{
		ID:     1,
		Method: "getblockhash",
		Params: []interface{}{height},
	}

	var blockHashResult GetBlockHashResult
	if err := d.Call(blockHashRequest, &blockHashResult); err != nil {
		return nil, err
	}

	if blockHashResult.Error.Message != "" {
		return nil, mapToStandardErr("Error fetching block hash: %s", blockHashResult.Error)
	}

	return &blockHashResult, nil
}

// GetBlockHeader returns the block header of the block the provided block hash.
func (d *ZCoreRPC) GetBlockHeader(hash string) (*bchain.BlockHeader, error) {
	blockHeaderRequest := GenericCmd{
		ID:     1,
		Method: "getblockheader",
		Params: []interface{}{hash},
	}

	var blockHeader GetBlockHeaderResult
	if err := d.Call(blockHeaderRequest, &blockHeader); err != nil {
		return nil, err
	}

	if blockHeader.Error.Message != "" {
		return nil, mapToStandardErr("Error fetching block info: %s", blockHeader.Error)
	}

	header := &bchain.BlockHeader{
		Hash:          blockHeader.Result.Hash,
		Prev:          blockHeader.Result.PreviousHash,
		Next:          blockHeader.Result.NextHash,
		Height:        blockHeader.Result.Height,
		Confirmations: int(blockHeader.Result.Confirmations),
		Time:          blockHeader.Result.Time,
	}

	return header, nil
}

// getChainBestBlock returns the best block according to dcrd chain. This block
// has no atleast one confirming block.
func (d *ZCoreRPC) getChainBestBlock() (*GetBestBlockResult, error) {
	bestBlockHashResult, err := d.getBestBlockHash()
	if err != nil {
		return nil, err
	}

	blockResult, err := d.getBlock(bestBlockHashResult.Result)
	if err != nil {
		return nil, err
	}
        var bestBlockResult GetBestBlockResult
        bestBlockResult.Result.Hash = blockResult.Result.Hash
        bestBlockResult.Result.Height = blockResult.Result.Height
	return &bestBlockResult, nil
}


// getBestBlock returns details for the block mined immediately before the
// official dcrd chain's bestblock i.e. it has a minimum of 1 confirmation.
// The chain's best block is not returned as its block validity is not guarranteed.
func (d *ZCoreRPC) getBestBlock() (*GetBestBlockResult, error) {
	bestBlockResult, err := d.getChainBestBlock()
	if err != nil {
		return nil, err
	}

	// remove the block with less than 1 confirming block
	bestBlockResult.Result.Height--
	validBlockHash, err := d.getBlockHashByHeight(bestBlockResult.Result.Height)
	if err != nil {
		return nil, err
	}

	bestBlockResult.Result.Hash = validBlockHash.Result

	return bestBlockResult, nil
}

// GetBlock returns the block retreived using the provided block hash by default
// or using the block height if an empty hash string was provided. If the
// requested block has less than 2 confirmation bchain.ErrBlockNotFound error
// is returned. This rule is in places to guarrantee that only validated block
// details (txs) are saved to the db. Access to the bestBlock height is threadsafe.
func (d *ZCoreRPC) GetBlock(hash string, height uint32) (*bchain.Block, error) {
	// Confirm if the block at provided height has at least 2 confirming blocks.
	d.mtx.Lock()
	var bestBlockHeight = d.bestBlock
	if height > bestBlockHeight {
		bestBlock, err := d.getBestBlock()
		if err != nil || height > bestBlock.Result.Height {
			// If an error occured or the current height doesn't have a minimum
			// of two confirming blocks (greater than best block), quit.
			d.mtx.Unlock()
			return nil, bchain.ErrBlockNotFound
		}

		d.bestBlock = bestBlock.Result.Height
		bestBlockHeight = bestBlock.Result.Height
	}
	d.mtx.Unlock() // Releases the lock soonest possible

	if hash == "" {
		getHashResult, err := d.getBlockHashByHeight(height)
		if err != nil {
			return nil, err
		}
		hash = getHashResult.Result
	}

	block, err := d.getBlock(hash)
	if err != nil {
		return nil, err
	}

	header := bchain.BlockHeader{
		Hash:          block.Result.Hash,
		Prev:          block.Result.PreviousHash,
		Next:          block.Result.NextHash,
		Height:        block.Result.Height,
		Confirmations: int(block.Result.Confirmations),
		Time:          block.Result.Time,
	}

	bchainBlock := &bchain.Block{BlockHeader: header}

	for _, txID := range block.Result.Tx {
		if block.Result.Height == 0 {
			continue
		}

		tx, err := d.GetTransaction(txID)
		if err != nil {
			return nil, err
		}

		bchainBlock.Txs = append(bchainBlock.Txs, *tx)
	}

	return bchainBlock, nil
}


func (d *ZCoreRPC) getBlock(hash string) (*GetBlockResult, error) {
	blockRequest := GenericCmd{
		ID:     1,
		Method: "getblock",
		Params: []interface{}{hash},
	}

	var block GetBlockResult
	if err := d.Call(blockRequest, &block); err != nil {
		return nil, err
	}

	if block.Error.Message != "" {
		return nil, mapToStandardErr("Error fetching block info: %s", block.Error)
	}

	return &block, nil
}

func (d *ZCoreRPC) getBestBlockHash() (*GetBestBlockHashResult, error) {
	bestBlockRequest := GenericCmd{
		ID:     1,
		Method: "getbestblockhash",
	}

	var bestBlockHash GetBestBlockHashResult
	if err := d.Call(bestBlockRequest, &bestBlockHash); err != nil {
		return nil, err
	}

	if bestBlockHash.Error.Message != "" {
		return nil, mapToStandardErr("Error fetching block info: %s", bestBlockHash.Error)
	}

	return &bestBlockHash, nil
}


func (d *ZCoreRPC) decodeRawTransaction(txHex string) (*bchain.Tx, error) {
	decodeRawTxRequest := GenericCmd{
		ID:     1,
		Method: "decoderawtransaction",
		Params: []interface{}{txHex},
	}

	var decodeRawTxResult DecodeRawTransactionResult
	if err := d.Call(decodeRawTxRequest, &decodeRawTxResult); err != nil {
		return nil, err
	}

	if decodeRawTxResult.Error.Message != "" {
		return nil, mapToStandardErr("Error decoding raw tx: %s", decodeRawTxResult.Error)
	}

	tx := &bchain.Tx{
		Hex:      txHex,
		Txid:     decodeRawTxResult.Result.Txid,
		Version:  decodeRawTxResult.Result.Version,
		LockTime: decodeRawTxResult.Result.Locktime,
	}

	return tx, nil
}


func (d *ZCoreRPC) GetBlockInfo(hash string) (*bchain.BlockInfo, error) {
	block, err := d.getBlock(hash)
	if err != nil {
		return nil, err
	}

	header := bchain.BlockHeader{
		Hash:          block.Result.Hash,
		Prev:          block.Result.PreviousHash,
		Next:          block.Result.NextHash,
		Height:        block.Result.Height,
		Confirmations: int(block.Result.Confirmations),
		Size:          int(block.Result.Size),
		Time:          int64(block.Result.Time),
	}

	bInfo := &bchain.BlockInfo{
		BlockHeader: header,
		MerkleRoot:  block.Result.MerkleRoot,
		Version:     block.Result.Version,
		Nonce:       block.Result.Nonce,
		Bits:        block.Result.Bits,
		Difficulty:  json.Number(strconv.FormatFloat(block.Result.Difficulty, 'e', -1, 64)),
		Txids:       block.Result.Tx,
	}

	return bInfo, nil
}

// GetTransaction returns a transaction by the transaction ID
func (d *ZCoreRPC) GetTransaction(txid string) (*bchain.Tx, error) {
	r, err := d.getRawTransaction(txid)
	if err != nil {
		return nil, err
	}

	tx, err := d.Parser.ParseTxFromJson(r)
	if err != nil {
		return nil, errors.Annotatef(err, "txid %v", txid)
	}

	return tx, nil
}

// getRawTransaction returns json as returned by backend, with all coin specific data
func (d *ZCoreRPC) getRawTransaction(txid string) (json.RawMessage, error) {
	if txid == "" {
		return nil, bchain.ErrTxidMissing
	}

	verbose := 1
	getTxRequest := GenericCmd{
		ID:     1,
		Method: "getrawtransaction",
		Params: []interface{}{txid, &verbose},
	}

	var getTxResult GetTransactionResult
	if err := d.Call(getTxRequest, &getTxResult); err != nil {
		return nil, err
	}

	if getTxResult.Error.Message != "" {
		return nil, mapToStandardErr("Error fetching transaction: %s", getTxResult.Error)
	}

	bytes, err := json.Marshal(getTxResult.Result)
	if err != nil {
		return nil, errors.Annotatef(err, "txid %v", txid)
	}

	return json.RawMessage(bytes), nil
}

func (d *ZCoreRPC) SendRawTransaction(tx string) (string, error) {
	sendRawTxRequest := &GenericCmd{
		ID:     1,
		Method: "sendrawtransaction",
		Params: []interface{}{tx},
	}

	var sendRawTxResult SendRawTransactionResult
	err := d.Call(sendRawTxRequest, &sendRawTxResult)
	if err != nil {
		return "", err
	}

	if sendRawTxResult.Error.Message != "" {
		return "", mapToStandardErr("error sending raw transaction: %s", sendRawTxResult.Error)
	}

	return sendRawTxResult.Result, nil
}


// Call calls Backend RPC interface, using RPCMarshaler interface to marshall the request
func (d *ZCoreRPC) Call(req interface{}, res interface{}) error {
	httpData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", d.rpcURL, bytes.NewBuffer(httpData))
	if err != nil {
		return err
	}
	httpReq.SetBasicAuth(d.rpcUser, d.rpcPassword)
	httpRes, err := d.client.Do(httpReq)
	// in some cases the httpRes can contain data even if it returns error
	// see http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/
	if httpRes != nil {
		defer httpRes.Body.Close()
	}
	if err != nil {
		return err
	}

	// if server returns HTTP error code it might not return json with response
	// handle both cases
	if httpRes.StatusCode != 200 {
		if err = safeDecodeResponse(httpRes.Body, &res); err != nil {
			return errors.Errorf("%v %v", httpRes.Status, err)
		}
		return nil
	}
	return safeDecodeResponse(httpRes.Body, &res)
}


func safeDecodeResponse(body io.ReadCloser, res *interface{}) (err error) {
	var data []byte
	defer func() {
		if r := recover(); r != nil {
			glog.Error("unmarshal json recovered from panic: ", r, "; data: ", string(data))
			debug.PrintStack()
			if len(data) > 0 && len(data) < 2048 {
				err = errors.Errorf("Error: %v", string(data))
			} else {
				err = errors.New("Internal error")
			}
		}
	}()
	data, err = ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	error := json.Unmarshal(data, res)
	return error
}

// mapToStandardErr map the dcrd API Message errors to the standard error messages
// supported by trezor. Dcrd errors to be mapped are listed here:
// https://github.com/decred/dcrd/blob/2f5e47371263b996bb99e8dc3484f659309bd83a/dcrjson/jsonerr.go
func mapToStandardErr(customPrefix string, err Error) error {
	switch {
	case strings.Contains(err.Message, dcrjson.ErrBlockNotFound.Message) || // Block not found
		strings.Contains(err.Message, dcrjson.ErrOutOfRange.Message) || // Block number out of range
		strings.Contains(err.Message, dcrjson.ErrBestBlockHash.Message): // Error getting best block hash
		return bchain.ErrBlockNotFound
	case strings.Contains(err.Message, dcrjson.ErrNoTxInfo.Message): // No information available about transaction
		return bchain.ErrTxNotFound
	case strings.Contains(err.Message, dcrjson.ErrInvalidTxVout.Message): // Output index number (vout) does not exist for transaction
		return bchain.ErrTxidMissing
	default:
		return fmt.Errorf(customPrefix, err.Message)
	}
}
