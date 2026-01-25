package trader

import (
	"encoding/json"
	"testing"

	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/linearswap"
	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/linearswap/restful"
	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/reqbuilder"
)

// 必须是合约专用的 API Key（在 HTX 官网 API 管理 → 合约权限）
const (
	accessKey = "ba925fe2-ca7a210e-a6c9a3ff-dbye2sf5t7"
	secretKey = "a8ded5d8-b02f4dab-f8aefcac-1084e"
)

func TestAccountInfo(t *testing.T) {
	// 多资产模式不让用
	ac := new(restful.AccountClient).Init(accessKey, secretKey, "")
	apiPath := "/linear-swap-api/v3/unified_account_info"
	url := ac.PUrlBuilder.Build(linearswap.GET_METHOD, apiPath, nil)
	resp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		t.Error(getErr)
	}
	t.Log(resp)
}

func TestAccountBalance(t *testing.T) {
	ac := new(restful.AccountClient).Init(accessKey, secretKey, "")
	apiPath := "/linear-swap-api/v1/swap_balance_valuation"
	url := ac.PUrlBuilder.Build(linearswap.POST_METHOD, apiPath, nil)
	data := map[string]any{
		"valuation_asset": "USDT",
	}
	jsonBytes, _ := json.Marshal(data)
	resp, getErr := reqbuilder.HttpPost(url, string(jsonBytes))
	if getErr != nil {
		t.Error(getErr)
	}
	t.Log(resp)
}
func TestOpenOrders(t *testing.T) {
	ac := new(restful.AccountClient).Init(accessKey, secretKey, "")
	apiPath := "/v5/trade/order/opens"
	url := ac.PUrlBuilder.Build(linearswap.GET_METHOD, apiPath, nil)
	resp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		t.Error(getErr)
	}
	t.Log(resp)
}
func TestGetPositions(t *testing.T) {
	ac := new(restful.AccountClient).Init(accessKey, secretKey, "")
	apiPath := "/v5/trade/position/opens"
	url := ac.PUrlBuilder.Build(linearswap.GET_METHOD, apiPath, nil)
	resp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		t.Error(getErr)
	}
	t.Log(resp)
}
func TestGetClosedPnL(t *testing.T) {
	ac := new(restful.AccountClient).Init(accessKey, secretKey, "")
	apiPath := "/v5/trade/order/details"
	url := ac.PUrlBuilder.Build(linearswap.GET_METHOD, apiPath, nil)
	resp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		t.Error(getErr)
	}
	t.Log(resp)
	var data map[string]interface{}
	err := json.Unmarshal([]byte(resp), &data)
	if err != nil {
		t.Error("解析失败:", err)
	}
	if data["code"] == 200 {
		details := data["data"]
		for _, detail := range details.([]interface{}) {
			contractCode := detail.(map[string]interface{})["contract_code"].(string)
			marginMode := detail.(map[string]interface{})["margin_mode"].(string)
			historyApiPath := "/v5/trade/order/history?contract_code=" + contractCode + "&margin_mode=" + marginMode
			historyUrl := ac.PUrlBuilder.Build(linearswap.GET_METHOD, historyApiPath, nil)
			historyResp, _ := reqbuilder.HttpGet(historyUrl)
			t.Log(historyResp)
		}
	}
	// todo 根据返回的合约标志，遍历查询
}
