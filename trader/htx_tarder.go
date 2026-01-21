package trader

import (
	"encoding/json"
	"fmt"
	"nofx/logger"
	"sync"
	"time"

	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/linearswap"
	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/linearswap/restful"
	responseorder "github.com/HuobiRDCenter/huobi_futures_Golang/sdk/linearswap/restful/response/order"
	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/reqbuilder"
)

type HtxTradePositionOpensResponse struct {
	Code int `json:"code"`

	Data []struct {
		ContractCode      string  `json:"contract_code"`
		PositionSide      string  `json:"position_side"`
		MarginMode        string  `json:"margin_mode"`
		CostOpen          string  `json:"cost_open"`
		Volume            string  `json:"volume"`
		Available         string  `json:"available"`
		LeverRate         string  `json:"lever_rate"`
		AdlRiskPercent    string  `json:"adl_risk_percent"`
		LiquidationPrice  string  `json:"liquidation_price"`
		InitialMargin     string  `json:"initial_margin"`
		MaintenanceMargin string  `json:"maintenance_margin"`
		ProfitUnreal      string  `json:"profit_unreal"`
		ProfitRate        string  `json:"profit_rate"`
		MarginRate        string  `json:"margin_rate"`
		MarginCurrency    string  `json:"margin_currency"`
		PositionMode      string  `json:"position_mode"`
		Last              float64 `json:"last"`
		ContractType      string  `json:"contract_type"`
		CreatedTime       string  `json:"created_time"`
		UpdatedTime       string  `json:"updated_time"`
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
		CreatedTime           string `json:"created_time"`
		UpdatedTime           string `json:"updated_time"`
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
			CreatedTime           string `json:"created_time"`
			UpdatedTime           string `json:"updated_time"`
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
	cachedBalance     map[string]interface{}
	balanceCacheTime  time.Time
	balanceCacheMutex sync.RWMutex
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
func (t *HtxTrader) GetBalance() (map[string]interface{}, error) {
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
	balance := map[string]interface{}{
		"totalEquity":           result.Data.Equity,
		"totalWalletBalance":    result.Data.Equity,
		"availableBalance":      result.Data.AvailableMargin,
		"totalUnrealizedProfit": result.Data.ProfitUnreal,
		"balance":               result.Data.Equity, // Compatible with other exchange formats
	}

	// Update cache
	t.balanceCacheMutex.Lock()
	t.cachedBalance = balance
	t.balanceCacheTime = time.Now()
	t.balanceCacheMutex.Unlock()

	return balance, nil
}

// GetPositions Get all positions
func (t *HtxTrader) GetPositions() ([]map[string]interface{}, error) {
	data := make(chan responseorder.GetTradePositionOpensResponse)
	t.client.GetTradePositionOpensAsync(data, "")

	return nil, nil
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
