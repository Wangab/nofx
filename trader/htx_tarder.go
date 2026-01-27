package trader

import (
	"encoding/json"
	"fmt"
	"nofx/logger"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/linearswap"
	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/linearswap/restful"
	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/reqbuilder"
)

type HtxTradePositionOpensResponse struct {
	Code int `json:"code"`

	Data []struct {
		ContractCode      string `json:"contract_code"`
		PositionSide      string `json:"position_side"`
		Direction         string `json:"direction"`
		MarginMode        string `json:"margin_mode"`
		OpenAvgPrice      string `json:"open_avg_price"`
		Volume            string `json:"volume"`
		Available         string `json:"available"`
		LeverRate         int    `json:"lever_rate"`
		AdlRiskPercent    string `json:"adl_risk_percent"`
		LiquidationPrice  string `json:"liquidation_price"`
		Margin            string `json:"margin"`
		InitialMargin     string `json:"initial_margin"`
		MaintenanceMargin string `json:"maintenance_margin"`
		ProfitUnreal      string `json:"profit_unreal"`
		ProfitRate        string `json:"profit_rate"`
		MarginRate        string `json:"margin_rate"`
		MarkPrice         string `json:"mark_price"`
		LastPrice         string `json:"last_price"`
		MarginCurrency    string `json:"margin_currency"`
		ContractType      string `json:"contract_type"`
		CreatedTime       string `json:"created_time"`
		UpdatedTime       string `json:"updated_time"`
	} `json:"data,omitempty"`

	Message string `json:"message,omitempty"`

	Ts int64 `json:"ts"`
}

type HTXAccountBalanceResponse struct {
	Code int `json:"code"`

	Data struct {
		State                 string `json:"state"`
		Equity                string `json:"equity"`
		InitialMargin         string `json:"initial_margin"`
		MaintenanceMargin     string `json:"maintenance_margin"`
		MaintenanceMarginRate string `json:"maintenance_margin_rate"`
		ProfitUnreal          string `json:"profit_unreal"`
		AvailableMargin       string `json:"available_margin"`
		CreatedTime           int64  `json:"created_time"`
		UpdatedTime           int64  `json:"updated_time"`
		Details               []struct {
			Currency              string `json:"currency"`
			Equity                string `json:"equity"`
			IsolatedEquity        string `json:"isolated_equity"`
			Available             string `json:"available"`
			WithdrawAvailable     string `json:"withdraw_available"`
			ProfitUnreal          string `json:"profit_unreal"`
			InitialMargin         string `json:"initial_margin"`
			MaintenanceMargin     string `json:"maintenance_margin"`
			MaintenanceMarginRate string `json:"maintenance_margin_rate"`
			InitialMarginRate     string `json:"initial_margin_rate"`
			CreatedTime           int64  `json:"created_time"`
			UpdatedTime           int64  `json:"updated_time"`
		} `json:"details"`
	} `json:"data,omitempty"`

	Message string `json:"message,omitempty"`

	Ts int64 `json:"ts"`
}
type HTXOrderResponse struct {
	Code int `json:"code"`
	Data struct {
		ClientOrderId string `json:"client_order_id"`
		OrderId       string `json:"order_id"`
	} `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Ts      int64  `json:"ts"`
}

type HTXContractPriceResponse struct {
	Status string `json:"status"`
	Data   []struct {
		Id            string `json:"id"`
		Basis         string `json:"basis"`
		BasisRate     string `json:"basis_rate"`
		ContractPrice string `json:"contract_price"`
		IndexPrice    string `json:"index_price"`
	} `json:"data,omitempty"`
	Ts int64 `json:"ts"`
}

type HtxTrader struct {
	apiKey    string
	secretKey string
	client    *restful.AccountClient
	// Balance cache
	cachedBalance     map[string]any
	balanceCacheTime  time.Time
	balanceCacheMutex sync.RWMutex
	// Position cache
	cachedPositions     []map[string]any
	positionsCacheTime  time.Time
	positionsCacheMutex sync.RWMutex
	// Cache duration (15 seconds)
	cacheDuration time.Duration
	marginMode    string
	leverRate     int
}

func NewHtxTrader(accessKey, secretKey string) *HtxTrader {
	acClient := new(restful.AccountClient).Init(accessKey, secretKey, "")
	trader := &HtxTrader{
		secretKey:     secretKey,
		apiKey:        accessKey,
		client:        acClient,
		cacheDuration: 15 * time.Second,
	}
	logger.Infof("🔵 [HTX] Trader initialized")
	return trader
}
func (t *HtxTrader) coverSymbol(symbol string) string {
	if !strings.Contains(symbol, "-") {
		symbol = strings.ReplaceAll(strings.ToUpper(symbol), "USDT", "-USDT")
	}
	return symbol
}

// GetBalance Get account balance
func (t *HtxTrader) GetBalance() (map[string]any, error) {
	// Check cache
	t.balanceCacheMutex.RLock()
	if t.cachedBalance != nil && time.Since(t.balanceCacheTime) < t.cacheDuration {
		balance := t.cachedBalance
		t.balanceCacheMutex.RUnlock()
		return balance, nil
	}
	t.balanceCacheMutex.RUnlock()

	url := t.client.PUrlBuilder.Build(linearswap.GET_METHOD, "/v5/account/balance", nil)
	getResp, getErr := reqbuilder.HttpGet(url)
	logger.Infof("✅ HTX Trader GetBalance response -> %s", getResp)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get account balance: %w", getErr)
	}
	result := HTXAccountBalanceResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to get account balance: %w", jsonErr)
	}
	if result.Code != 200 {
		return nil, fmt.Errorf("failed to get account balance: %s", result.Message)
	}
	totalEquity, _ := strconv.ParseFloat(result.Data.Equity, 64)
	availableBalance, _ := strconv.ParseFloat(result.Data.AvailableMargin, 64)
	totalUnrealizedProfit, _ := strconv.ParseFloat(result.Data.ProfitUnreal, 64)
	balance := map[string]interface{}{
		"totalEquity":           totalEquity,
		"totalWalletBalance":    totalEquity,
		"availableBalance":      availableBalance,
		"totalUnrealizedProfit": totalUnrealizedProfit,
		"balance":               totalEquity, // Compatible with other exchange formats
	}

	// Update cache
	t.balanceCacheMutex.Lock()
	t.cachedBalance = balance
	t.balanceCacheTime = time.Now()
	t.balanceCacheMutex.Unlock()

	return balance, nil
}

// GetPositions Get all positions
func (t *HtxTrader) GetPositions() ([]map[string]any, error) {
	// Check cache
	t.positionsCacheMutex.RLock()
	if t.cachedPositions != nil && time.Since(t.positionsCacheTime) < t.cacheDuration {
		positions := t.cachedPositions
		t.positionsCacheMutex.RUnlock()
		return positions, nil
	}
	t.positionsCacheMutex.RUnlock()
	// Call API
	url := t.client.PUrlBuilder.Build(linearswap.GET_METHOD, "/v5/trade/position/opens", nil)
	getResp, err := reqbuilder.HttpGet(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get htx positions: %w", err)
	}
	result := HtxTradePositionOpensResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to get htx positions: %w", jsonErr)
	}
	if result.Code != 200 {
		return nil, fmt.Errorf("failed to get htx positions: %s", result.Message)
	}
	list := result.Data

	var positions []map[string]interface{}

	for _, item := range list {
		symbol := item.ContractCode
		side := item.PositionSide
		positionAmt := item.Volume
		entryPrice, _ := strconv.ParseFloat(item.OpenAvgPrice, 64)
		markPrice, _ := strconv.ParseFloat(item.MarkPrice, 64)
		unrealisedProfit, _ := strconv.ParseFloat(item.ProfitUnreal, 64)
		unrealisedPnl, _ := strconv.ParseFloat(item.ProfitRate, 64)
		liqPrice, _ := strconv.ParseFloat(item.LiquidationPrice, 64)
		leverage := item.LeverRate
		createdTime, _ := strconv.ParseInt(item.UpdatedTime, 10, 64)
		updatedTime, _ := strconv.ParseInt(item.UpdatedTime, 10, 64)
		logger.Infof("[HTX] GetPositions symbol=%v, side=%s", symbol, side)
		position := map[string]any{
			"symbol":           symbol,
			"side":             side,
			"positionAmt":      positionAmt,
			"entryPrice":       entryPrice,
			"markPrice":        markPrice,
			"unRealizedProfit": unrealisedProfit,
			"unrealizedPnL":    unrealisedPnl,
			"liquidationPrice": liqPrice,
			"leverage":         leverage,
			"createdTime":      createdTime, // Position open time (ms)
			"updatedTime":      updatedTime, // Position last update time (ms)
		}

		positions = append(positions, position)
	}
	// Update cache
	t.positionsCacheMutex.Lock()
	t.cachedPositions = positions
	t.positionsCacheTime = time.Now()
	t.positionsCacheMutex.Unlock()

	return positions, nil
}
func (t *HtxTrader) OpenLongLimit(symbol string, quantity float64, leverage int, price float64, takeProfit float64, stopLoss float64) (map[string]interface{}, error) {
	return t.OpenLong(symbol, quantity, leverage)
}

func (t *HtxTrader) OpenShortLimit(symbol string, quantity float64, leverage int, price float64, takeProfit float64, stopLoss float64) (map[string]interface{}, error) {
	return t.OpenShort(symbol, quantity, leverage)
}

// OpenLong Open long position
func (t *HtxTrader) OpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// First cancel all pending orders for this symbol (clean up old stop-loss and take-profit orders)
	if err := t.CancelAllOrders(symbol); err != nil {
		logger.Infof("  ⚠ Failed to cancel old pending orders (may not have any): %v", err)
	}
	if err := t.SetLeverage(symbol, leverage); err != nil {
		logger.Infof("  ⚠ Failed to set lever: %v", err)
	}
	symbol = t.coverSymbol(symbol)
	url := t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/v5/trade/order", nil)
	marginMode := "cross"
	if t.marginMode != "" {
		marginMode = t.marginMode
	}
	data := map[string]any{
		"contract_code": symbol,
		"margin_mode":   marginMode,
		"position_side": "long",
		"side":          "buy",
		"type":          "market",
		"volume":        quantity,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("❌ Htx OpenLong failed: %w", err)
	}
	resp, getErr := reqbuilder.HttpPost(url, string(jsonBytes))
	if getErr != nil {
		return nil, fmt.Errorf("❌ Htx OpenLong failed: %w", getErr)
	}
	result := HTXOrderResponse{}
	jsonErr := json.Unmarshal([]byte(resp), &result)
	if jsonErr != nil {
		return nil, fmt.Errorf("❌ failed to open htx long order: %w", jsonErr)
	}
	if result.Code != 200 {
		return nil, fmt.Errorf("❌ failed to open htx long order: %s", result.Message)
	}
	return map[string]any{
		"orderId": result.Data.OrderId,
		"symbol":  symbol,
		"status":  "FILLED",
	}, nil
}

// OpenShort Open short position
func (t *HtxTrader) OpenShort(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// First cancel all pending orders for this symbol (clean up old stop-loss and take-profit orders)
	if err := t.CancelAllOrders(symbol); err != nil {
		logger.Infof("  ⚠ Failed to cancel old pending orders (may not have any): %v", err)
	}
	if err := t.SetLeverage(symbol, leverage); err != nil {
		logger.Infof("  ⚠ Failed to set lever: %v", err)
	}
	symbol = t.coverSymbol(symbol)
	url := t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/v5/trade/order", nil)
	marginMode := "cross"
	if t.marginMode != "" {
		marginMode = t.marginMode
	}
	data := map[string]any{
		"contract_code": symbol,
		"margin_mode":   marginMode,
		"position_side": "short",
		"side":          "sell",
		"type":          "market",
		"volume":        quantity,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("❌ Htx OpenShort failed: %w", err)
	}
	resp, getErr := reqbuilder.HttpPost(url, string(jsonBytes))
	if getErr != nil {
		return nil, fmt.Errorf("❌ Htx OpenShort failed: %w", getErr)
	}
	result := HTXOrderResponse{}
	jsonErr := json.Unmarshal([]byte(resp), &result)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to open htx long order: %w", jsonErr)
	}
	if result.Code != 200 {
		return nil, fmt.Errorf("failed to open htx long order: %s", result.Message)
	}
	return map[string]any{
		"orderId": result.Data.OrderId,
		"symbol":  symbol,
		"status":  "FILLED",
	}, nil
}

// CloseLong Close long position (quantity=0 means close all)
func (t *HtxTrader) CloseLong(symbol string, quantity float64) (map[string]interface{}, error) {
	symbol = t.coverSymbol(symbol)
	url := t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/v5/trade/order", nil)
	marginMode := "cross"
	if t.marginMode != "" {
		marginMode = t.marginMode
	}
	data := map[string]any{
		"contract_code": symbol,
		"margin_mode":   marginMode,
		"position_side": "long",
		"side":          "sell",
		"type":          "market",
		"volume":        quantity,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("❌ Htx CloseLong failed: %w", err)
	}
	resp, getErr := reqbuilder.HttpPost(url, string(jsonBytes))
	if getErr != nil {
		return nil, fmt.Errorf("❌ Htx CloseLong failed: %w", getErr)
	}
	result := HTXOrderResponse{}
	jsonErr := json.Unmarshal([]byte(resp), &result)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to open htx long order: %w", jsonErr)
	}
	if result.Code != 200 {
		return nil, fmt.Errorf("failed to open htx long order: %s", result.Message)
	}
	return map[string]any{
		"orderId": result.Data.OrderId,
		"symbol":  symbol,
		"status":  "FILLED",
	}, nil
}

// CloseShort Close short position (quantity=0 means close all)
func (t *HtxTrader) CloseShort(symbol string, quantity float64) (map[string]interface{}, error) {
	symbol = t.coverSymbol(symbol)
	url := t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/v5/trade/order", nil)
	marginMode := "cross"
	if t.marginMode != "" {
		marginMode = t.marginMode
	}
	data := map[string]any{
		"contract_code": symbol,
		"margin_mode":   marginMode,
		"position_side": "short",
		"side":          "buy",
		"type":          "market",
		"volume":        quantity,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("❌ Htx CloseShort failed: %w", err)
	}
	resp, getErr := reqbuilder.HttpPost(url, string(jsonBytes))
	if getErr != nil {
		return nil, fmt.Errorf("❌ Htx CloseShort failed: %w", err)
	}
	result := HTXOrderResponse{}
	jsonErr := json.Unmarshal([]byte(resp), &result)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to open htx long order: %w", jsonErr)
	}
	if result.Code != 200 {
		return nil, fmt.Errorf("failed to open htx long order: %s", result.Message)
	}
	return map[string]any{
		"orderId": result.Data.OrderId,
		"symbol":  symbol,
		"status":  "FILLED",
	}, nil
}

// SetLeverage Set leverage
func (t *HtxTrader) SetLeverage(symbol string, leverage int) error {
	symbol = t.coverSymbol(symbol)
	url := t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/v5/position/lever", nil)
	marginMode := "cross"
	if t.marginMode != "" {
		marginMode = t.marginMode
	}
	data := map[string]any{
		"contract_code": symbol,
		"margin_mode":   marginMode,
		"lever_rate":    strconv.Itoa(leverage),
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("❌ Htx SetLerverage failed: %w", err)
	}
	_, getErr := reqbuilder.HttpPost(url, string(jsonBytes))
	if getErr != nil {
		return fmt.Errorf("❌ Htx SetLerverage failed: %w", err)
	}
	return nil
}

// SetMarginMode Set position mode (true=cross margin, false=isolated margin)
func (t *HtxTrader) SetMarginMode(symbol string, isCrossMargin bool) error {
	symbol = t.coverSymbol(symbol)
	if isCrossMargin {
		t.marginMode = "cross"
	} else {
		t.marginMode = "isolated"
	}
	url := t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/v5/position/lever", nil)
	lever := 5
	if t.leverRate != 0 {
		lever = t.leverRate
	}
	data := map[string]any{
		"contract_code": symbol,
		"margin_mode":   t.marginMode,
		"lever_rate":    strconv.Itoa(lever),
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("❌ Htx SetMarginMode failed: %w", err)
	}
	_, getErr := reqbuilder.HttpPost(url, string(jsonBytes))
	if getErr != nil {
		return fmt.Errorf("❌ Htx SetMarginMode failed: %w", getErr)
	}
	return nil
}

// GetMarketPrice Get market price
func (t *HtxTrader) GetMarketPrice(symbol string) (float64, error) {
	symbol = t.coverSymbol(symbol)
	url := t.client.PUrlBuilder.Build(linearswap.GET_METHOD, "/index/market/history/linear_swap_basis", nil)
	requestUrl := fmt.Sprintf("%s?contract_code=%s&period=1min&size=1", url, symbol)
	resp, err := reqbuilder.HttpGet(requestUrl)
	if err != nil {
		return 0, fmt.Errorf("failed to get htx GetMarketPrice: %w", err)
	}
	result := HTXContractPriceResponse{}
	jsonErr := json.Unmarshal([]byte(resp), &result)
	if jsonErr != nil {
		return 0, fmt.Errorf("failed to open htx market price: %w", jsonErr)
	}
	if result.Status != "ok" {
		return 0, fmt.Errorf("failed to open htx market price: %v", result)
	}
	if len(result.Data) > 0 {
		price, _ := strconv.ParseFloat(result.Data[0].ContractPrice, 64)
		return price, nil
	}
	return 0, nil
}

// SetStopLoss Set stop-loss order
func (t *HtxTrader) SetStopLoss(symbol string, positionSide string, quantity, stopPrice float64) error {
	symbol = t.coverSymbol(symbol)
	url := t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/linear-swap-api/v1/swap_cross_tpsl_order", nil)
	if t.marginMode == "isolated" {
		url = t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/linear-swap-api/v1/swap_tpsl_order", nil)
	}
	direction := "buy"
	if positionSide == "LONG" {
		direction = "sell"
	}
	data := map[string]any{
		"contract_code":    symbol,
		"direction":        direction,
		"volume":           quantity,
		"sl_trigger_price": stopPrice,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("❌ Htx SetStopLoss failed: %w", err)
	}
	_, getErr := reqbuilder.HttpPost(url, string(jsonBytes))
	if getErr != nil {
		return fmt.Errorf("❌ Htx SetStopLess failed: %w", err)
	}
	return nil
}

// SetTakeProfit Set take-profit order
func (t *HtxTrader) SetTakeProfit(symbol string, positionSide string, quantity, takeProfitPrice float64) error {
	symbol = t.coverSymbol(symbol)
	url := t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/linear-swap-api/v1/swap_cross_tpsl_order", nil)
	if t.marginMode == "isolated" {
		url = t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/linear-swap-api/v1/swap_tpsl_order", nil)
	}
	direction := "buy"
	if positionSide == "LONG" {
		direction = "sell"
	}
	data := map[string]any{
		"contract_code":    symbol,
		"direction":        direction,
		"volume":           quantity,
		"tp_trigger_price": takeProfitPrice,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("❌ Htx SetTakeProfit failed: %w", err)
	}
	_, getErr := reqbuilder.HttpPost(url, string(jsonBytes))
	if getErr != nil {
		return fmt.Errorf("❌ Htx SetTakeProfit failed: %w", getErr)
	}
	return nil
}

// CancelStopLossOrders Cancel only stop-loss orders (BUG fix: don't delete take-profit when adjusting stop-loss)
func (t *HtxTrader) CancelStopLossOrders(symbol string) error {
	htxSymbol := t.coverSymbol(symbol)
	apiPath := "/v5/trade/order/opens?contract_code=" + htxSymbol
	url := t.client.PUrlBuilder.Build(linearswap.GET_METHOD, apiPath, nil)
	resp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		return getErr
	}
	var data struct {
		List []struct {
			Id             string `json:"id"`
			ContractCode   string `json:"contract_code"`
			OrderId        string `json:"order_id"`
			TpTriggerPrice string `json:"tp_trigger_price"`
			SlTriggerPrice string `json:"sl_trigger_price"`
		} `json:"data"`
	}
	err := json.Unmarshal([]byte(resp), &data)
	if err != nil {
		return fmt.Errorf("❌ Htx CancelStopLossOrders failed: %w", err)
	}
	var ids []string
	for _, pos := range data.List {
		if pos.SlTriggerPrice != "" {
			ids = append(ids, pos.OrderId)
		}
	}
	if len(ids) > 0 {
		cancelUrl := t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/v5/trade/cancel_batch_orders", nil)
		dataPost := map[string]any{
			"contract_code": symbol,
			"order_id":      ids,
		}
		jsonBytes, err2 := json.Marshal(dataPost)
		if err2 != nil {
			return fmt.Errorf("❌ Htx CanselStopLoss failed: %w", err2)
		}
		_, e := reqbuilder.HttpPost(cancelUrl, string(jsonBytes))
		if e != nil {
			return e
		}
	}
	return nil
}

// CancelTakeProfitOrders Cancel only take-profit orders (BUG fix: don't delete stop-loss when adjusting take-profit)
func (t *HtxTrader) CancelTakeProfitOrders(symbol string) error {
	htxSymbol := t.coverSymbol(symbol)
	apiPath := "/v5/trade/order/opens?contract_code=" + htxSymbol
	url := t.client.PUrlBuilder.Build(linearswap.GET_METHOD, apiPath, nil)
	resp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		return getErr
	}
	var data struct {
		List []struct {
			Id             string `json:"id"`
			ContractCode   string `json:"contract_code"`
			OrderId        string `json:"order_id"`
			TpTriggerPrice string `json:"tp_trigger_price"`
			SlTriggerPrice string `json:"sl_trigger_price"`
		} `json:"data"`
	}
	err := json.Unmarshal([]byte(resp), &data)
	if err != nil {
		return fmt.Errorf("❌ Htx CancelTakeProfitOrders failed: %w", err)
	}
	var ids []string
	for _, pos := range data.List {
		if pos.TpTriggerPrice != "" {
			ids = append(ids, pos.OrderId)
		}
	}
	if len(ids) > 0 {
		cancelUrl := t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/v5/trade/cancel_batch_orders", nil)
		dataPost := map[string]any{
			"contract_code": symbol,
			"order_id":      ids,
		}
		jsonBytes, err1 := json.Marshal(dataPost)
		if err1 != nil {
			return fmt.Errorf("❌ Htx CancelTakeProfitOrders failed: %w", err1)
		}
		_, e := reqbuilder.HttpPost(cancelUrl, string(jsonBytes))
		if e != nil {
			return fmt.Errorf("❌ Htx CancelTakeProfitOrders failed: %w", e)
		}
	}
	return nil
}

// CancelAllOrders Cancel all pending orders for this symbol
func (t *HtxTrader) CancelAllOrders(symbol string) error {
	symbol = t.coverSymbol(symbol)
	// url
	url := t.client.PUrlBuilder.Build(linearswap.POST_METHOD, "/v5/trade/cancel_all_orders", nil)
	data := map[string]any{
		"contract_code": symbol,
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("❌ Htx CancelAllOrders failed: %w", err)
	}
	_, getErr := reqbuilder.HttpPost(url, string(jsonBytes))
	if getErr != nil {
		return fmt.Errorf("❌ Htx CancelALLOrders failed: %w", getErr)
	}
	return nil
}

// CancelStopOrders Cancel stop-loss/take-profit orders for this symbol (for adjusting stop-loss/take-profit positions)
func (t *HtxTrader) CancelStopOrders(symbol string) error {
	err1 := t.CancelStopLossOrders(symbol)
	if err1 != nil {
		return fmt.Errorf("❌ Htx CancelStopOrders failed: %w", err1)
	}
	err2 := t.CancelTakeProfitOrders(symbol)
	if err2 != nil {
		return fmt.Errorf("❌ Htx CancelStopOrders failed: %w", err2)
	}
	return nil
}

// FormatQuantity Format quantity to correct precision
func (t *HtxTrader) FormatQuantity(symbol string, quantity float64) (string, error) {
	return strconv.FormatFloat(quantity, 'f', -1, 64), nil
}

// GetOrderStatus Get order status
// Returns: status(FILLED/NEW/CANCELED), avgPrice, executedQty, commission
func (t *HtxTrader) GetOrderStatus(symbol string, orderID string) (map[string]interface{}, error) {
	symbol = t.coverSymbol(symbol)
	apiPath := fmt.Sprintf("/v5/trade/order?contract_code=%s&order_id=%s", symbol, orderID)
	url := t.client.PUrlBuilder.Build(linearswap.GET_METHOD, apiPath, nil)
	resp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		return nil, fmt.Errorf("❌ Htx GetOrderStatus failed: %w", getErr)
	}
	var data struct {
		State         string `json:"state"`
		TradeAvgPrice string `json:"trade_avg_price"`
		TradeVolume   string `json:"trade_volume"`
		RealProfit    string `json:"real_profit"`
	}
	err := json.Unmarshal([]byte(resp), &data)
	if err != nil {
		return nil, fmt.Errorf("❌ Htx GetOrderStatus failed: %w", err)
	}
	status := data.State
	unifiedStatus := status
	switch status {
	case "filled":
		unifiedStatus = "FILLED"
	case "new":
		unifiedStatus = "NEW"
	case "canceled", "rejected":
		unifiedStatus = "CANCELED"
	case "partially_filled", "partially_canceled":
		unifiedStatus = "PARTIALLY_FILLED"
	}
	return map[string]interface{}{
		"orderId":     orderID,
		"status":      unifiedStatus,
		"avgPrice":    data.TradeAvgPrice,
		"executedQty": data.TradeVolume,
		"commission":  data.RealProfit,
	}, nil
}

// GetClosedPnL Get closed position PnL records from exchange
// startTime: start time for query (usually last sync time)
// limit: max number of records to return
// Returns accurate exit price, fees, and close reason for positions closed externally
func (t *HtxTrader) GetClosedPnL(startTime time.Time, limit int) ([]ClosedPnLRecord, error) {
	apiPath := fmt.Sprintf("/v5/trade/order/details?start_time=%d&limit=%d", startTime.UnixMilli(), limit)
	url := t.client.PUrlBuilder.Build(linearswap.GET_METHOD, apiPath, nil)
	resp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		return nil, fmt.Errorf("❌ Htx GetClosePNL failed: %w", getErr)
	}
	var data struct {
		List []struct {
			Id             string  `json:"id"`
			ContractCode   string  `json:"contract_code"`
			OrderId        string  `json:"order_id"`
			TradeId        string  `json:"trade_id"`
			Side           string  `json:"side"`
			PositionSide   string  `json:"position_side"`
			OderType       string  `json:"order_type"`
			MarginMode     string  `json:"margin_mode"`
			Type           string  `json:"type"`
			Role           string  `json:"role"`
			TradePrice     string  `json:"trade_price"`
			TradeVolume    string  `json:"trade_volume"`
			TradeTurnover  string  `json:"trade_turnover"`
			CreateTime     float64 `json:"created_time"`
			UpdateTime     float64 `json:"updated_time"`
			OrderSource    string  `json:"order_source"`
			FeeCurrency    string  `json:"fee_currency"`
			TradeFee       string  `json:"trade_fee"`
			DeductionPrice string  `json:"deduction_price"`
			Profit         string  `json:"profit"`
			ContractType   string  `json:"contract_type"`
		} `json:"data"`
	}
	err := json.Unmarshal([]byte(resp), &data)
	if err != nil {
		return nil, fmt.Errorf("❌ Htx GetClosePNL failed: %w", err)
	}
	records := make([]ClosedPnLRecord, 0, len(data.List))
	for _, pos := range data.List {
		historyApiPath := "/v5/trade/order/history?contract_code=" + pos.ContractCode + "&margin_mode=" + pos.MarginMode
		historyUrl := t.client.PUrlBuilder.Build(linearswap.GET_METHOD, historyApiPath, nil)
		historyResp, e := reqbuilder.HttpGet(historyUrl)
		if e != nil {
			return nil, e
		}
		var historyData struct {
			List []struct {
				Id                 string `json:"id"`
				ContractCode       string `json:"contract_code"`
				Side               string `json:"side"`
				PositionSide       string `json:"position_side"`
				PriceMatch         string `json:"price_match"`
				OrderId            string `json:"order_id"`
				ClientOrderId      string `json:"client_order_id"`
				MarginMode         string `json:"margin_mode"`
				Price              string `json:"price"`
				Volume             string `json:"volume"`
				LeverRate          int64  `json:"lever_rate"`
				State              string `json:"state"`
				OrderSource        string `json:"order_source"`
				ReduceOnly         bool   `json:"reduce_only"`
				TimeInForce        string `json:"time_in_force"`
				TpTriggerPrice     string `json:"tp_trigger_price"`
				TpOrderPrice       string `json:"tp_order_price"`
				TpType             string `json:"tp_type"`
				TpTriggerPriceType string `json:"tp_trigger_price_type"`
				SlTriggerPrice     string `json:"sl_trigger_price"`
				SlOrderPrice       string `json:"sl_order_price"`
				SlType             string `json:"sl_type"`
				SlTriggerPriceType string `json:"sl_trigger_price_type"`
				TradeAvgPrice      string `json:"trade_avg_price"`
				TradeVolume        string `json:"trade_volume"`
				TradeTurnover      string `json:"trade_turnover"`
				FeeCurrency        string `json:"fee_currency"`
				Fee                string `json:"fee"`
				PriceProtect       string `json:"price_protect"`
				Profit             string `json:"profit"`
				ContractType       string `json:"contract_type"`
				CancelReason       string `json:"cancel_reason"`
				CreateTime         int64  `json:"created_time"`
				UpdateTime         int64  `json:"updated_time"`
			} `json:"data"`
		}
		err = json.Unmarshal([]byte(historyResp), &historyData)
		if err != nil {
			return nil, fmt.Errorf("❌ Htx GetClosePNL failed: %w", err)
		}
		for _, posd := range historyData.List {
			record := ClosedPnLRecord{
				Symbol: posd.ContractCode,
				Side:   posd.PositionSide,
			}
			record.EntryPrice, _ = strconv.ParseFloat(posd.Price, 64)
			record.ExitPrice, _ = strconv.ParseFloat(posd.TradeAvgPrice, 64)
			record.Quantity, _ = strconv.ParseFloat(posd.TradeVolume, 64)
			record.RealizedPnL, _ = strconv.ParseFloat(posd.Profit, 64)
			record.Fee, _ = strconv.ParseFloat(posd.Fee, 64)
			record.Leverage = int(posd.LeverRate)
			record.EntryTime = time.UnixMilli(posd.CreateTime).UTC()
			record.ExitTime = time.UnixMilli(posd.UpdateTime).UTC()
			record.CloseType = "unknown"
			records = append(records, record)
		}
	}
	return records, nil
}

// GetOpenOrders Get open/pending orders from exchange
// Returns stop-loss, take-profit, and limit orders that haven't been filled
func (t *HtxTrader) GetOpenOrders(symbol string) ([]OpenOrder, error) {
	htxSymbol := t.coverSymbol(symbol)
	apiPath := "/v5/trade/order/opens?contract_code=" + htxSymbol
	url := t.client.PUrlBuilder.Build(linearswap.GET_METHOD, apiPath, nil)
	resp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		return nil, fmt.Errorf("❌ Htx GetOpenOrders failed: %w", getErr)
	}
	var data map[string]interface{}
	err := json.Unmarshal([]byte(resp), &data)
	if err != nil {
		return nil, fmt.Errorf("❌ Htx GetOpenOrders failed: %w", err)
	}
	var orders []OpenOrder
	if data["code"] == 200 {
		details := data["data"]
		for _, detail := range details.([]interface{}) {
			orderId := detail.(map[string]interface{})["order_id"].(string)
			side := detail.(map[string]interface{})["side"].(string)
			positionSide := detail.(map[string]interface{})["position_side"].(string)
			orderType := detail.(map[string]interface{})["type"].(string)

			ord := OpenOrder{
				OrderID:      orderId,
				Symbol:       symbol,
				Side:         side,         // BUY/SELL
				PositionSide: positionSide, // LONG/SHORT
				Type:         orderType,    // LIMIT/STOP_MARKET/TAKE_PROFIT_MARKET
				Status:       strings.ToUpper(detail.(map[string]interface{})["status"].(string)),
			}
			if price, e := strconv.ParseFloat(detail.(map[string]interface{})["price"].(string), 64); e != nil {
				return nil, fmt.Errorf("❌ Htx GetOpenOrders failed: %w", e)
			} else {
				ord.Price = price
			}
			if stopPrice, e := strconv.ParseFloat(detail.(map[string]interface{})["sl_trigger_price"].(string), 64); e != nil {
				return nil, fmt.Errorf("❌ Htx GetOpenOrders failed: %w", e)
			} else {
				ord.StopPrice = stopPrice
			}
			if quantity, e := strconv.ParseFloat(detail.(map[string]interface{})["volume"].(string), 64); e != nil {
				return nil, fmt.Errorf("❌ Htx GetOpenOrders failed: %w", e)
			} else {
				ord.Quantity = quantity
			}
			orders = append(orders, ord)
		}
	}
	return orders, nil
}
