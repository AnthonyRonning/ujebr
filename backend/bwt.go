package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type bwtClient struct {
	bwtHost string
}

func NewBwtClient(bwtUrl string, bwtPort int) (*bwtClient, error) {
	bwtHost := bwtUrl + ":" + strconv.Itoa(bwtPort) + "/"

	bwtClient := &bwtClient{
		bwtHost: bwtHost,
	}

	// test
	_, err := bwtClient.GetTransactions()
	if err != nil {
		return nil, err
	}

	return bwtClient, nil
}

func (b *bwtClient) GetTransactions() ([]*WalletTransaction, error) {
	resp, err := b.getRequest("txs")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// http response error handling
	if resp.StatusCode != 200 {
		return nil, errors.New("received an error from bwt: " + resp.Status)
	}

	// response
	var txsResp []*WalletTransaction
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&txsResp)
	if err != nil {
		return nil, err
	}

	// fmt.Println("Received transactions")
	// fmt.Println(txsResp)

	return txsResp, nil
}

func (b *bwtClient) GetTransactionRaw(txid string) (*RawTransaction, error) {
	resp, err := b.getRequest(fmt.Sprintf("tx/%s/verbose", txid))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// http response error handling
	if resp.StatusCode != 200 {
		return nil, errors.New("received an error from bwt: " + resp.Status)
	}
	// response
	var txResp RawTransaction
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&txResp)
	if err != nil {
		return nil, err
	}

	// fmt.Println("Received raw transaction")
	// fmt.Println(txResp)

	return &txResp, nil
}

func (b *bwtClient) getRequest(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", b.bwtHost+endpoint, nil)
	if err != nil {
		return nil, err
	}

	return b.makeRequest(req)
}

func (b *bwtClient) makeRequest(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type WalletTransaction struct {
	Txid                  string     `json:"txid"`
	BlockHeight           int        `json:"block_height"`
	Funding               []Funding  `json:"funding"`
	Spending              []Spending `json:"spending"`
	BalanceChange         int        `json:"balance_change"`
	OwnFeerate            float64    `json:"own_feerate,omitempty"`
	EffectiveFeerate      float64    `json:"effective_feerate,omitempty"`
	Bip125Replaceable     bool       `json:"bip125_replaceable,omitempty"`
	HasUnconfirmedParents bool       `json:"has_unconfirmed_parents,omitempty"`
}

type Funding struct {
	Vout         int         `json:"vout"`
	Address      string      `json:"address"`
	Scripthash   string      `json:"scripthash"`
	Origin       string      `json:"origin"`
	Desc         string      `json:"desc"`
	Bip32Origins []string    `json:"bip32_origins"`
	Amount       int64       `json:"amount"`
	SpentBy      interface{} `json:"spent_by"`
}

type Spending struct {
	Vin          int      `json:"vin"`
	Address      string   `json:"address"`
	Scripthash   string   `json:"scripthash"`
	Origin       string   `json:"origin"`
	Desc         string   `json:"desc"`
	Bip32Origins []string `json:"bip32_origins"`
	Amount       int64    `json:"amount"`
	Prevout      string   `json:"prevout"`
}

type RawTransaction struct {
	Blockhash     string `json:"blockhash"`
	Blocktime     int    `json:"blocktime"`
	Confirmations int    `json:"confirmations"`
	Hash          string `json:"hash"`
	Hex           string `json:"hex"`
	InActiveChain bool   `json:"in_active_chain"`
	Locktime      int    `json:"locktime"`
	Size          int    `json:"size"`
	Time          int    `json:"time"`
	Txid          string `json:"txid"`
	Version       int    `json:"version"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	Vsize         int    `json:"vsize"`
	Weight        int    `json:"weight"`
}

type Vout struct {
	N            int `json:"n"`
	ScriptPubKey struct {
		Addresses []string `json:"addresses"`
		Asm       string   `json:"asm"`
		Hex       string   `json:"hex"`
		ReqSigs   int      `json:"reqSigs"`
		Type      string   `json:"type"`
	} `json:"scriptPubKey"`
	Value float64 `json:"value"`
}

type Vin struct {
	ScriptSig struct {
		Asm string `json:"asm"`
		Hex string `json:"hex"`
	} `json:"scriptSig"`
	Sequence    int64    `json:"sequence"`
	Txid        string   `json:"txid"`
	Txinwitness []string `json:"txinwitness"`
	Vout        int      `json:"vout"`
}
