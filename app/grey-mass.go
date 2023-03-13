package app

import (
	"encoding/json"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type APIClient struct {
	Host string `json:"host"`
}

// GreyMass data

type GMParams struct {
	Signer *GMSigner `json:"signer"`
	// PackedTransaction *GMPackedTx `json:"packedTransaction"`
	Transaction *eos.Transaction `json:"transaction"`
}

type GMSigner struct {
	Actor      string `json:"actor"`
	Permission string `json:"permission"`
}

type GMPackedTx struct {
	Compression           uint8           `json:"compression"`
	PackedContextFreeData string          `json:"packed_context_free_data"`
	PackedTrx             string          `json:"packed_trx"`
	Signatures            []ecc.Signature `json:"signatures"`
}

type GMResponse struct {
	Code    uint32         `json:"code"`
	Message string         `json:"message"`
	Data    GMResponseData `json:"data"`
}

type GMResponseData struct {
	Request    []interface{}   `json:"request"`
	Signatures []ecc.Signature `json:"signatures"`
	Version    string          `json:"version"`
}

// Send  request
func (c *APIClient) Send(url, body string, method string) (data []byte, err error) {
	
	if !strings.HasPrefix(url, "http") {
		
		url = fmt.Sprintf("%s%s", c.Host, url)
	}
	
	// make request
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		
		return
	}
	
	req.Header.Add("content-type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.55 Safari/537.36")
	req.Header.Add("sec-ch-ua-platform", "macOS")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-site")
	
	// do send
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		
		return
	}
	
	// read response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		
		return
	}
	
	data = respBody
	return
}

// RequestTxByGM request grey-mass resource api
func (c *APIClient) RequestTxByGM(singer eos.PermissionLevel, tx *eos.Transaction) (freeTx *eos.SignedTransaction,
	signatures []ecc.Signature, err error) {
	
	params := &GMParams{
		Signer: &GMSigner{
			// Actor:      action.Authorization[0].Actor.String(),
			// Permission: action.Authorization[0].Permission.String(),
			Actor:      singer.Actor.String(),
			Permission: singer.Permission.String(),
		},
		Transaction: tx,
	}
	
	data, err := json.Marshal(params)
	if err != nil {
		
		return
	}
	
	gmAPI := "/v1/resource_provider/request_transaction"
	resp, err := c.Send(gmAPI, string(data), http.MethodPost)
	if err != nil {
		
		return
	}
	
	var respData GMResponse
	if err = json.Unmarshal(resp, &respData); err != nil {
		
		return
	}
	
	if respData.Code != 200 {
		
		err = errors.New(respData.Message)
		return
	}
	
	reqTxData, err := json.Marshal(respData.Data.Request[1])
	if err != nil {
		
		return
	}
	
	var signedFreeTx eos.SignedTransaction
	if err = json.Unmarshal(reqTxData, &signedFreeTx); err != nil {
		
		return
	}
	
	freeTx = &signedFreeTx
	signatures = respData.Data.Signatures
	return
}
