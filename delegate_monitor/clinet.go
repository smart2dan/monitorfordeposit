package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	//"strconv"
	"strings"
	//"time"
	//	"usdtsyncd/db"

	logger "github.com/alecthomas/log4go"
	simplejson "github.com/bitly/go-simplejson"
)

const (
	listblocktransactions = "omni_listblocktransactions"
	gettrade              = "omni_gettrade"
	getblockcount         = "getblockcount"
	listaddressgroupings  = "listaddressgroupings"
)

type WalletClient struct {
	confPath string
	cfg      *Configure
	httpUrl  string
	//url      string
	bExit bool
}

type HttpRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	//	Id      string        `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type HttpResponse struct {
	Result []string `json:"result"`
	Id     string   `json:"id"`
	Error  int      `json:"error"`
}
type NomarlRes struct {
	Result string
	Id     string
	Error  int
}
type NormalIntRes struct {
	Result int
	Id     string
	Error  int
}
type AddressInfo struct {
	addr    string
	balance float64
}
type HttpAddrRes struct {
	Result []interface{} `json:"result"`
	Id     string        `json:"id"`
	Error  int           `json:"error"`
}

type VoteInfoRes struct {
	//Info  []interface{} `json:"result"`
	Id     string `json:"id"`
	Error  int    `json:"error"`
	Result []struct {
		Name     string
		Delegate string
	}
}

type Delegates struct {
	//Info  []interface{} `json:"result"`
	Id     string `json:"id"`
	Error  int    `json:"error"`
	Result []struct {
		Name    string
		Address string
	}
}

type HttpResponseEx struct {
	Result int    `json:"result"`
	Id     string `json:"id"`
	Error  int    `json:"error"`
}

func CreateWalletTask(cfg *Configure) *WalletClient {
	url := fmt.Sprintf("http://%s:%s@%s:%d", cfg.Main.Rpcuser, cfg.Main.Rpcpassword,
		cfg.Main.Rpcip, cfg.Main.Rpcport)
	wallet := &WalletClient{
		confPath: defaultConfig,
		cfg:      cfg,
		httpUrl:  url,
		// url:      cfg.Db.Url,
		bExit: false,
	}
	return wallet
}

func (c *WalletClient) HttpPost(method string, params []interface{}) ([]byte, error) {
	//url := "http://bitcoin:MeVz4mcTYZgvY4fT3bECaf8YkWKYxdhhZ1@47.75.43.198:8332"
	request := &HttpRequest{
		Jsonrpc: "1.0",
		//	Id:      c.cfg.Main.AppCode,
		Method: method,
		Params: params,
	}

	val, err := json.Marshal(request)
	//fmt.Println(string(val))

	if err != nil {
		logger.Error("Marshal json http request failed! err: %s", err)
		return []byte(""), err
	}

	payload := strings.NewReader(string(val))
	//fmt.Println(payload)
	req, err1 := http.NewRequest("POST", c.httpUrl, payload)
	if err1 != nil {
		logger.Error("create request failed! url:%s err: %s", c.httpUrl, err)
		return []byte(""), err1
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("cache-control", "no-cache")

	res, err2 := http.DefaultClient.Do(req)
	//fmt.Println(res)
	if err2 != nil {
		logger.Error("post http request failed! err: %s", err2)
		return []byte(""), err2
	}
	defer res.Body.Close()
	body, err3 := ioutil.ReadAll(res.Body)
	if err3 != nil {
		logger.Error("read http body failed! err: %s", err3)
		return []byte(""), err3
	}
	return body, nil
}

func (c *WalletClient) GetAddrInfo() ([]AddressInfo, error) {
	var params []interface{}
	var httpRes HttpAddrRes

	res, err := c.HttpPost(listaddressgroupings, params)

	if err != nil || strings.Compare(string(res), "") == 0 {
		logger.Error("http request failed!")
		return nil, errors.New("http request failed!")
	}
	err = json.Unmarshal(res, &httpRes)
	if err != nil {
		logger.Error("json Unmarshal failed")
		return nil, err
	}
	data := reflect.ValueOf(httpRes.Result)
	var addresInfo []AddressInfo = make([]AddressInfo, data.Len())
	for i := 0; i < data.Len(); i++ {
		temp1 := reflect.ValueOf(data.Index(i).Interface())
		temp2 := reflect.ValueOf(temp1.Index(0).Interface())
		addresInfo[i].addr = temp2.Index(0).Interface().(string)
		addresInfo[i].balance = temp2.Index(1).Interface().(float64)
		//fmt.Println(addresInfo[i].addr)
		//fmt.Println(addresInfo[i].balance)
	}
	return addresInfo, nil
}

func (c *WalletClient) GetTxInfo(txid string) (string, error) {
	var params []interface{}
	params = append(params, txid)
	res, err := c.HttpPost("gettransactionnew", params)
	if err != nil || strings.Compare(string(res), "") == 0 {
		logger.Error("http request failed!")
		return nil, error.New("http request failed!")
	}

}

func (c *WalletClient) ListDelegates() (map[string]string, error) {
	var params []interface{}
	var HttpRes Delegates
	var AllDelegate map[string]string
	AllDelegate = make(map[string]string)
	res, err := c.HttpPost("listdelegates", params)

	if err != nil || strings.Compare(string(res), "") == 0 {
		logger.Error("http request failed!")
		return nil, errors.New("http request failed!")
	}
	err = json.Unmarshal(res, &HttpRes)
	if err != nil {
		logger.Error("json Unmarshal failed")
		return nil, err
	}

	for _, v := range HttpRes.Result {
		AllDelegate[v.Address] = v.Name
	}
	//fmt.Println(AllDelegate)
	return AllDelegate, nil
}
func (c *WalletClient) getBalance(address string) (int, error) {
	var params []interface{}
	var HttpRes NormalIntRes
	params = append(params, address)
	res, err := c.HttpPost("getaddressbalance", params)
	if err != nil || strings.Compare(string(res), "") == 0 {
		logger.Error("http request failed!")
		return -1, nil
	}
	//fmt.Println(string(res))
	err = json.Unmarshal(res, &HttpRes)
	if err != nil {
		logger.Error("json Unmarshal failed")
		return -1, nil
	}
	//fmt.Println(HttpRes.Result)
	return HttpRes.Result, nil
}
func (c *WalletClient) sendmany()
func (c *WalletClient) getnewAddress() (string, error) {
	var params []interface{}
	var HttpRes NomarlRes
	res, err := c.HttpPost("getnewaddress", params)
	if err != nil || strings.Compare(string(res), "") == 0 {
		logger.Error("http request failed!")
		return nil, errors.New("http request failed!")
	}
	err = json.Unmarshal(res, &HttpRes)
	if err != nil {
		logger.Error("json unmarshal failed!")
		return nil, err
	}
	return HttpRes.Result, nil
}

func (c *WalletClient) GetVoted(address string) ([]string, error) {
	var params []interface{}
	var HttpRes VoteInfoRes
	var Delegates []string
	params = append(params, address)
	res, err := c.HttpPost("listvoteddelegates", params)
	if err != nil || strings.Compare(string(res), "") == 0 {
		logger.Error("http request failed!")
		return nil, errors.New("http request failed!")
	}
	err = json.Unmarshal(res, &HttpRes)
	if err != nil {
		logger.Error("json Unmarshal failed")
		return nil, err
	}

	for _, v := range HttpRes.Result {
		//fmt.Println(v.Name)
		Delegates = append(Delegates, v.Name)
	}
	//fmt.Println(Delegates)
	return Delegates, nil
}
func (c *WalletClient) SendFrom(from string, des string, amount float64) {
	var params []interface{}
	var HttpRes NomarlRes
	params = append(params, from)
	params = append(params, des)

	// amount_s := fmt.Sprintf("%f", amount)
	amount_s := fmt.Sprintf("%.8f", amount)
	params = append(params, amount_s)
	fmt.Println(params)

	res, err := c.HttpPost("sendfromaddress", params)
	if err != nil || strings.Compare(string(res), "") == 0 {
		logger.Error("http request failed!")
		//return "request err", err
	}
	fmt.Println(string(res))
	err = json.Unmarshal(res, &HttpRes)
	if err != nil {
		fmt.Println(from)
		logger.Error("json Umarshal failed")
		//return "json error", err
	}
	//fmt.Println(HttpRes)
	//return HttpRes.Result, nil
}
func (c *WalletClient) VoteforDelegate(address string, name []string, status bool) (string, error) {
	var params []interface{}
	var HttpRes NomarlRes
	var cmd string
	params = append(params, address)
	for _, v := range name {
		params = append(params, v)
	}
	if status == true {
		cmd = "vote"
	} else {
		cmd = "cancelvote"
	}
	fmt.Println(cmd)
	res, err := c.HttpPost(cmd, params)
	if err != nil || strings.Compare(string(res), "") == 0 {
		logger.Error("http request failed!")
		return "request err", err
	}
	//fmt.Println(string(res))
	err = json.Unmarshal(res, &HttpRes)
	if err != nil {
		logger.Error("json Umarshal failed")
		return "json error", err
	}
	//fmt.Println(HttpRes)
	return HttpRes.Result, nil

}

/*
type TransactionInfo struct {
	Amount           string
	Block            int
	Blockhash        string
	Blocktime        int64
	Confirmations    int
	Divisible        bool
	Fee              string
	Ismine           bool
	Positioninblock  int
	Propertyid       int
	Referenceaddress string
	Sendingaddress   string
	Txid             string
	Type             string
	Typeint          int
	Valid            bool
	Version          int
}

type HttpTradeResponse struct {
	Trade TransactionInfo `json:"result"`
	Id    string          `json:"id"`
	Error int             `json:"error"`
}

type HttpAddressResponse struct {
	Set   AddressSet `json:"result"`
	Id    string     `json:"id"`
	Error int        `json:"error"`
}
type AddressSet struct {
	base58   string
	cashaddr string
	bitpay   string
}
func (c *WalletClient) HttpGet(address string) ([]byte, error) {
	url := f/("https://bch.btc.com/tools/bch-addr-convert?address=bitcoincash%%3A%s", address)
	fmt.Println(string(url))
	req, err := http.Get(url)
	if err != nil {
		logger.Error("convert address failed! err: %s", err)
		return []byte(""), err
	}
	val, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return []byte(""), err
	}
	fmt.Println(string(val))
	return []byte(""), err
}
func (c *WalletClient) GetTradeByTxid(txids []string) {
	for _, val := range txids {
		if strings.Compare(val, "") == 0 {
			continue
		}

		var count int64 = 0
		var params []interface{}
		var tradeRes HttpTradeResponse
		params = append(params, val)

	LoopRequest: //如果出现错误就必须重试，直到请成功。
		res, err := c.HttpPost(gettrade, params)
		if err != nil {
			count += 5
			if count == 60 {
				count = 0
				logger.Warn("Http request failed, waiting for retry. url: %s param: %s", c.httpUrl, val)
			}
			time.Sleep(5 * time.Second)
			goto LoopRequest
		}

		err = json.Unmarshal(res, &tradeRes)
		if err != nil {
			logger.Error("json Unmarshal failed!")
			time.Sleep(50 * time.Millisecond)
			continue
		}

		if tradeRes.Trade.Propertyid != 31 {
			continue
		}

		if strings.Compare(tradeRes.Id, c.cfg.Main.AppCode) != 0 {
			logger.Error("recv appid error! discard!")
			continue
		}

		ts := &db.Transaction{
			Amount:           tradeRes.Trade.Amount,
			Block:            tradeRes.Trade.Block,
			Blockhash:        tradeRes.Trade.Blockhash,
			Blocktime:        tradeRes.Trade.Blocktime,
			Confirmations:    tradeRes.Trade.Confirmations,
			Divisible:        tradeRes.Trade.Divisible,
			Fee:              tradeRes.Trade.Fee,
			Ismine:           tradeRes.Trade.Ismine,
			Positioninblock:  tradeRes.Trade.Positioninblock,
			Propertyid:       tradeRes.Trade.Propertyid,
			Referenceaddress: tradeRes.Trade.Referenceaddress,
			Sendingaddress:   tradeRes.Trade.Sendingaddress,
			Txid:             tradeRes.Trade.Txid,
			Type:             tradeRes.Trade.Type,
			Typeint:          tradeRes.Trade.Typeint,
			Valid:            tradeRes.Trade.Valid,
			Version:          tradeRes.Trade.Version,
		}
		//var resp string
		_, err = mgo.AddTransaction(ts)
		if err != nil {
			logger.Error("add transaction info failed! err: %s", err)
			continue
		}
		sync := &db.SyncLastTrans{
			SerId:          db.LastRecordId,
			Sendingaddress: tradeRes.Trade.Sendingaddress,
			Txid:           tradeRes.Trade.Txid,
			Block:          tradeRes.Trade.Block,
		}
		err = mgo.UpdateSyncLastTransInfo(sync)
		if err != nil {
			logger.Error("update sync last trans info failed! err: %s", err)
			_, err = mgo.FindSyncLastTransInfo(db.LastRecordId)
			if err != nil {
				_, err = mgo.AddSyncLastTransInfo(sync)
				if err != nil {
					logger.Info("add sycn last trans info failed! err: %s", err)
				}
			}
		}
		//logger.Info("TxHash: %s", tradeRes.Trade.Txid)
	}
}

func (c *WalletClient) GetBlockCount() (int, error) {
	var params []interface{}
	var httpRes HttpResponseEx

	res, err := c.HttpPost(getblockcount, params)
	if err != nil || strings.Compare(string(res), "") == 0 {
		logger.Error("http request failed!")
		return 0, errors.New("http request failed!")
	}

	err = json.Unmarshal(res, &httpRes)
	if err != nil {
		logger.Error("json Unmarshal failed!")
		return 0, err
	}

	return httpRes.Result, nil
}

func (c *WalletClient) Run() {
	blockHeight := c.cfg.Main.InitBlockHeight
	syncLast, err1 := mgo.FindSyncLastTransInfo(db.LastRecordId)
	if err1 != nil {
		sync := &db.SyncLastTrans{
			SerId:          db.LastRecordId,
			Sendingaddress: "",
			Txid:           "",
			Block:          blockHeight,
		}
		_, err2 := mgo.AddSyncLastTransInfo(sync)
		if err2 != nil {
			logger.Info("add sycn last trans info failed: err: %s", err2)
			return
		}

	} else {
		blockHeight = syncLast.Block
		c.cfg.Main.InitBlockHeight = syncLast.Block
		err := mgo.RemoveTransaction(blockHeight)
		if err == nil {
			logger.Info("Remove last blockhigh transaction data,scan data again! block: %d", blockHeight)
		}
	}

	for c.bExit != true {
		var params []interface{}
		var httpRes HttpResponse

		blockCount, err0 := c.GetBlockCount()
		if err0 != nil {
			time.Sleep(60 * time.Second)
			continue
		}

		logger.Info("block count: %d", blockCount)
		if blockHeight <= blockCount {
			params = append(params, blockHeight)
			blockHeight += 1
			res, err := c.HttpPost(listblocktransactions, params)
			if err != nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			err = json.Unmarshal(res, &httpRes)
			if err != nil {
				logger.Error("json Unmarshal failed!")
				time.Sleep(50 * time.Millisecond)
				continue
			}

			if strings.Compare(httpRes.Id, c.cfg.Main.AppCode) != 0 {
				logger.Error("recv appid error! discard!")
				continue
			}

			sync := &db.SyncLastTrans{
				SerId: db.LastRecordId,
				Block: blockHeight - 1,
			}
			err = mgo.UpdateSyncLastTransInfo(sync)
			if err != nil {
				logger.Error("update sync last trans info failed! err: %s", err)
				_, err = mgo.FindSyncLastTransInfo(db.LastRecordId)
				if err != nil {
					_, err = mgo.AddSyncLastTransInfo(sync)
					if err != nil {
						logger.Info("add sycn last trans info failed! err: %s", err)
					}
				}
			}
			logger.Info("blockHeight: %d", blockHeight-1)

			c.GetTradeByTxid(httpRes.Result)
			time.Sleep(1 * time.Millisecond)
		} else {
			time.Sleep(5 * time.Second)
		}

	}
}
*/
