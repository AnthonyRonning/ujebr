package backend

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

type recovery struct {
	bClient   *rpcclient.Client
	bwtClient *bwtClient
}

type RecoveryCfg struct {
	BitcoindHost string
	BitcoindPort int
	BitcoindUser string
	BitcoindPass string
	BwtUrl       string
	BwtPort      int
}

func NewRecovery(cfg *RecoveryCfg) (*recovery, error) {
	// Connect to local bitcoin core RPC server using HTTP POST mode.
	// TODO: Core access not needed yet.
	/*
		connCfg := &rpcclient.ConnConfig{
			Host:         cfg.BitcoindHost + ":" + strconv.Itoa(cfg.BitcoindPort),
			User:         cfg.BitcoindUser,
			Pass:         cfg.BitcoindPass,
			HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
			DisableTLS:   true, // Bitcoin core does not provide TLS by default
		}

		client, err := rpcclient.New(connCfg, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer client.Shutdown()

		// Get the current block count.
		blockCount, err := client.GetBlockCount()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Block count: %d", blockCount)
	*/

	// Connect to bwt
	bwtClient, err := NewBwtClient(cfg.BwtUrl, cfg.BwtPort)
	if err != nil {
		return nil, err
	}

	return &recovery{
		//bClient:   client,
		bwtClient: bwtClient,
	}, nil
}

// Recover takes in a recovery address and optional seed and will attempt
// to query BWT in order to find transactions that are recoverable.
// If seed is not configured then the dry run will be enabled. Only unsigned
// transactions will be created.
// If seed is enabled then the it will create the signed transaction(s).
func (r *recovery) Recover(addr string, seed string) ([]string, error) {
	// Get all transactions from bwt
	txs, err := r.bwtClient.GetTransactions()
	if err != nil {
		return nil, err
	}

	// Look through each one to find replaceable transactions
	replaceableTxs := make([]*WalletTransaction, 0)
	for _, tx := range txs {
		if tx.Bip125Replaceable {
			replaceableTxs = append(replaceableTxs, tx)
		}
	}

	if len(replaceableTxs) == 0 {
		return nil, errors.New("no transactions replaceable")
	}

	// Go through each replaceable transaction and create new transaction from it
	unsignedTransactions := make([]string, 0)
	for _, replaceableTx := range replaceableTxs {
		unsignedTransaction, err := r.NewRecoveryTxFromReplaceable(addr, replaceableTx)
		if err != nil {
			return nil, err
		}

		// fmt.Println("Created unsigned transaction")
		var txBuff bytes.Buffer
		err = unsignedTransaction.Serialize(&txBuff)
		if err != nil {
			return nil, err
		}

		unsignedTransactions = append(unsignedTransactions, hex.EncodeToString(txBuff.Bytes()))
	}

	return unsignedTransactions, nil
}

func (r *recovery) NewRecoveryTxFromReplaceable(destination string, replaceable *WalletTransaction) (*wire.MsgTx, error) {
	// Go through all inputs we own and get the raw transaction for them
	prevInputTxs, err := r.GetPreviousInputInfo(replaceable)
	if err != nil {
		return nil, err
	}

	// extracting destination address as []byte from function argument (destination string)
	destinationAddr, err := btcutil.DecodeAddress(destination, &chaincfg.TestNet3Params)
	if err != nil {
		return nil, err
	}

	destinationAddrByte, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
		return nil, err
	}

	// creating a new bitcoin transaction, different sections of the tx, including
	// input list (contain UTXOs) and outputlist (contain destination address and usually our address)
	// in next steps, sections will be field and pass to sign
	redeemTx, err := NewTx()
	if err != nil {
		return nil, err
	}

	totalAmount := int64(0)
	for _, prevInputTx := range prevInputTxs {
		utxoHash, err := chainhash.NewHashFromStr(prevInputTx.TxId)
		if err != nil {
			return nil, err
		}

		// the second argument is vout or Tx-index, which is the index
		// of spending UTXO in the transaction that Txid referred to
		// in this case is 0, but can vary different numbers
		outPoint := wire.NewOutPoint(utxoHash, prevInputTx.Index)

		// making the input, and adding it to transaction
		txIn := wire.NewTxIn(outPoint, nil, nil)

		// Set non-replaceable sequence
		txIn.Sequence = wire.MaxTxInSequenceNum

		redeemTx.AddTxIn(txIn)

		// add total amount of each prev input
		totalAmount += prevInputTx.Amount
	}

	// Compute amount & fee
	// TODO: calculate it better to account for sig added at the end
	baseSizeWithoutTxOut := redeemTx.SerializeSize()

	// adding the destination address and the amount to
	// the transaction as output
	dummyTxOut := wire.NewTxOut(totalAmount, destinationAddrByte)
	txOutSize := dummyTxOut.SerializeSize()

	projectedSize := baseSizeWithoutTxOut + txOutSize
	higherSatPerVByteFee := 1 + int64(replaceable.OwnFeerate) // just needs to be 1 sat bigger than what it is replacing
	totalFee := higherSatPerVByteFee * int64(projectedSize)
	totalAmount -= totalFee

	// adding the destination address and the real amount to the transaction as output
	redeemTxOut := wire.NewTxOut(totalAmount, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	return redeemTx, nil
}

type PrevInputInfo struct {
	TxId     string
	PkScript string
	Index    uint32
	Amount   int64
}

func (r *recovery) GetPreviousInputInfo(tx *WalletTransaction) ([]*PrevInputInfo, error) {
	// Go through each spending utxo and get more info of each
	previousInputs := make([]*PrevInputInfo, 0)
	for _, spend := range tx.Spending {
		// parse txid:index
		prevOutParsed := strings.Split(spend.Prevout, ":")
		if len(prevOutParsed) != 2 {
			return nil, fmt.Errorf("unexpected prevout: %s", spend.Prevout)
		}
		txid := prevOutParsed[0]
		index, err := strconv.ParseUint(prevOutParsed[1], 32, 10)
		if err != nil {
			return nil, err
		}

		previousInputInfo := &PrevInputInfo{
			TxId:   txid,
			Index:  uint32(index),
			Amount: spend.Amount,
		}

		// Lookup the pkscript
		rawTx, err := r.bwtClient.GetTransactionRaw(txid)
		if err != nil {
			return nil, err
		}
		if len(rawTx.Vout) < int(index) {
			return nil, errors.New("not enough outputs expected for transaction")
		}

		previousInputInfo.PkScript = rawTx.Vout[int(index)].ScriptPubKey.Hex

		previousInputs = append(previousInputs, previousInputInfo)
	}

	return previousInputs, nil
}

func SignTx(privKey string, pkScript string, redeemTx *wire.MsgTx) (string, error) {

	wif, err := btcutil.DecodeWIF(privKey)
	if err != nil {
		return "", err
	}

	sourcePKScript, err := hex.DecodeString(pkScript)
	if err != nil {
		return "", nil
	}

	// since there is only one input in our transaction
	// we use 0 as second argument, if the transaction
	// has more args, should pass related index
	signature, err := txscript.SignatureScript(redeemTx, 0, sourcePKScript, txscript.SigHashAll, wif.PrivKey, false)
	if err != nil {
		return "", nil
	}

	// since there is only one input, and want to add
	// signature to it use 0 as index
	redeemTx.TxIn[0].SignatureScript = signature

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)

	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

	return hexSignedTx, nil
}
