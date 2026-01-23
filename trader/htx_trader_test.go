package trader

import (
	"encoding/json"
	"testing"

	"github.com/HuobiRDCenter/huobi_futures_Golang/config"
	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/linearswap"
	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/linearswap/restful"
	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/reqbuilder"
)

// 必须是合约专用的 API Key（在 HTX 官网 API 管理 → 合约权限）
const (
	accessKey = "ba925fe2-ca7a210e-a6c9a3ff-dbye2sf5t7"
	secretKey = "a8ded5d8-b02f4dab-f8aefcac-1084e"
)

func TestAccountBalance(t *testing.T) {
	// 初始化 AccountClient（linear swap USDT-M）
	client := new(restful.AccountClient).Init(accessKey, secretKey, "") // 第三个参数 host 留空默认 api.htx.com 或兼容
	url := client.PUrlBuilder.Build(linearswap.GET_METHOD, "/v5/account/balance", nil)
	getResp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		t.Logf("http get error: %s", getErr)
	}
	result := HTXAccountBalanceResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		t.Logf("convert json error: %s", jsonErr)
	}
	t.Logf("getResp: %v", result)
}

func TestPositions(t *testing.T) {
	// 初始化 AccountClient（linear swap USDT-M）
	client := new(restful.AccountClient).Init(accessKey, secretKey, "") // 第三个参数 host 留空默认 api.htx.com 或兼容
	// ulr
	url := client.PUrlBuilder.Build(linearswap.GET_METHOD, "/v5/trade/position/opens", nil)
	getResp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		t.Logf("http get error: %s", getErr)
	}
	result := HtxTradePositionOpensResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		t.Logf("convert json error: %s", getResp)
	}
	t.Logf("getResp: %v", getResp)
}

func TestOrdersHistory(t *testing.T) {
	// 初始化 AccountClient（linear swap USDT-M）
	client := new(restful.AccountClient).Init(accessKey, secretKey, "")
	// ulr
	url := client.PUrlBuilder.Build(linearswap.GET_METHOD, "/v5/trade/order/details", nil)
	getResp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		t.Logf("http get error: %s", getErr)
	}
	result := HtxTradeOrdersResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		t.Logf("convert json error: %s", jsonErr)
	}
	t.Logf("getResp: %v", getResp)
}

func TestCancelOrder(t *testing.T) {
	client := new(restful.OrderClient).Init(accessKey, secretKey, config.Host)
	// url
	url := client.PUrlBuilder.Build(linearswap.POST_METHOD, "/v5/trade/cancel_all_orders", nil)

	getResp, getErr := reqbuilder.HttpPost(url, "{\"contract_code\": \"BTC-USDT\"}")
	if getErr != nil {
		t.Logf("http get error: %s", getErr)
	}
	t.Logf("getResp: %v", getResp)
}
