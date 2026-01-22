package trader

import (
	"encoding/json"
	"fmt"
	"nofx/logger"
	"strconv"
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

// OpenLong Open long position
func (t *HtxTrader) OpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	return nil, nil
}

// OpenShort Open short position
func (t *HtxTrader) OpenShort(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	return nil, nil
}

// CloseLong Close long position (quantity=0 means close all)
func (t *HtxTrader) CloseLong(symbol string, quantity float64) (map[string]interface{}, error) {
	return nil, nil
}

// CloseShort Close short position (quantity=0 means close all)
func (t *HtxTrader) CloseShort(symbol string, quantity float64) (map[string]interface{}, error) {
	return nil, nil
}

// SetLeverage Set leverage
func (t *HtxTrader) SetLeverage(symbol string, leverage int) error {
	return nil
}

// SetMarginMode Set position mode (true=cross margin, false=isolated margin)
func (t *HtxTrader) SetMarginMode(symbol string, isCrossMargin bool) error {
	return nil
}

// GetMarketPrice Get market price
func (t *HtxTrader) GetMarketPrice(symbol string) (float64, error) {
	return 0, nil
}

// SetStopLoss Set stop-loss order
func (t *HtxTrader) SetStopLoss(symbol string, positionSide string, quantity, stopPrice float64) error {
	return nil
}

// SetTakeProfit Set take-profit order
func (t *HtxTrader) SetTakeProfit(symbol string, positionSide string, quantity, takeProfitPrice float64) error {
	return nil
}

// CancelStopLossOrders Cancel only stop-loss orders (BUG fix: don't delete take-profit when adjusting stop-loss)
func (t *HtxTrader) CancelStopLossOrders(symbol string) error {
	return nil
}

// CancelTakeProfitOrders Cancel only take-profit orders (BUG fix: don't delete stop-loss when adjusting take-profit)
func (t *HtxTrader) CancelTakeProfitOrders(symbol string) error {
	return nil
}

// CancelAllOrders Cancel all pending orders for this symbol
func (t *HtxTrader) CancelAllOrders(symbol string) error {
	return nil
}

// CancelStopOrders Cancel stop-loss/take-profit orders for this symbol (for adjusting stop-loss/take-profit positions)
func (t *HtxTrader) CancelStopOrders(symbol string) error {
	return nil
}

// FormatQuantity Format quantity to correct precision
func (t *HtxTrader) FormatQuantity(symbol string, quantity float64) (string, error) {
	return "", nil
}

// GetOrderStatus Get order status
// Returns: status(FILLED/NEW/CANCELED), avgPrice, executedQty, commission
func (t *HtxTrader) GetOrderStatus(symbol string, orderID string) (map[string]interface{}, error) {
	return nil, nil
}

// GetClosedPnL Get closed position PnL records from exchange
// startTime: start time for query (usually last sync time)
// limit: max number of records to return
// Returns accurate exit price, fees, and close reason for positions closed externally
func (t *HtxTrader) GetClosedPnL(startTime time.Time, limit int) ([]ClosedPnLRecord, error) {
	return nil, nil
}

// GetOpenOrders Get open/pending orders from exchange
// Returns stop-loss, take-profit, and limit orders that haven't been filled
func (t *HtxTrader) GetOpenOrders(symbol string) ([]OpenOrder, error) {
	return nil, nil
}
