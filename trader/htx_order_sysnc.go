package trader

import (
	"encoding/json"
	"fmt"
	"nofx/logger"
	"nofx/market"
	"nofx/store"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/linearswap"
	"github.com/HuobiRDCenter/huobi_futures_Golang/sdk/reqbuilder"
	"github.com/shopspring/decimal"
)

type HtxTradeOrdersResponse struct {
	Code    int          `json:"code"`
	Data    []TradeOrder `json:"data,omitempty"`
	Message string       `json:"message,omitempty"`
	Ts      int64        `json:"ts"`
}
type TradeOrder struct {
	Id             string `json:"id"`
	ContractCode   string `json:"contract_code"`
	OrderId        string `json:"order_id"`
	TradeId        string `json:"trade_id"`
	Side           string `json:"side"`
	PositionSide   string `json:"position_side"`
	OrderType      string `json:"order_type"`
	MarginMode     string `json:"margin_mode"`
	Type           string `json:"type"`
	Role           string `json:"role"`
	TradePrice     string `json:"trade_price"`
	TradeVolume    string `json:"trade_volume"`
	TradeTurnover  string `json:"trade_turnover"`
	CreatedTime    string `json:"created_time"`
	UpdatedTime    string `json:"updated_time"`
	OrderStatus    string `json:"order_status"`
	FeeCurrency    string `json:"fee_currency"`
	TradeFee       string `json:"trade_fee"`
	DeductionPrice string `json:"deduction_price"`
	Profit         string `json:"profit"`
	ContractType   string `json:"contract_type"`
}

func (t *HtxTrader) StartOrderSync(traderID string, exchangeID string, exchangeType string, store *store.Store, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			trader, err := store.Trader().GetByID(traderID)
			if err != nil {
				logger.Errorf("交易員[%s]檢查出錯 --> 跳過訂單同步 --> %v", traderID, err)
				continue
			}
			if trader != nil && trader.IsRunning {
				if err := t.SyncOrdersFromHTX(traderID, exchangeID, exchangeType, store); err != nil {
					logger.Infof("⚠️  HTX order sync failed: %v", err)
				}
			} else {
				logger.Warnf("交易員[%s]不存在或者已經停止交易 --> 暫停訂單同步", traderID)
			}
		}
	}()
	logger.Infof("🔄 HTX order sync started (interval: %v)", interval)
}

// GetTrades retrieves trade/execution records from HTX
func (t *HtxTrader) GetTrades() ([]TradeOrder, error) {
	url := t.client.PUrlBuilder.Build(linearswap.GET_METHOD, "/v5/trade/order/details", nil)
	getResp, getErr := reqbuilder.HttpGet(url)
	if getErr != nil {
		return nil, fmt.Errorf("HTX API error: %s", getErr)
	}
	result := HtxTradeOrdersResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, fmt.Errorf("HTX API error: %s", jsonErr)
	}
	return result.Data, nil
}

func (t *HtxTrader) SyncOrdersFromHTX(traderID string, exchangeID string, exchangeType string, st *store.Store) any {
	if st == nil {
		return fmt.Errorf("store is nil")
	}
	// Get recent trades (last 24 hours)
	startTime := time.Now().Add(-24 * time.Hour)
	logger.Infof("🔄 Syncing HTX trades from: %s", startTime.Format(time.RFC3339))
	// Use GetTrades method to fetch trade records
	trades, err := t.GetTrades()
	if err != nil {
		return fmt.Errorf("failed to get trades: %w", err)
	}
	logger.Infof("📥 Received %d trades from HTX", len(trades))
	// Sort trades by time ASC (oldest first) for proper position building
	sort.Slice(trades, func(i, j int) bool {
		iexecTimeMs, _ := strconv.ParseInt(trades[i].CreatedTime, 10, 64)
		iexecTime := time.UnixMilli(iexecTimeMs).UTC()
		jexecTimeMs, _ := strconv.ParseInt(trades[j].CreatedTime, 10, 64)
		jexecTime := time.UnixMilli(jexecTimeMs).UTC()
		return iexecTime.UnixMilli() < jexecTime.UnixMilli()
	})
	// Process trades one by one (no transaction to avoid deadlock)
	orderStore := st.Order()
	positionStore := st.Position()
	posBuilder := store.NewPositionBuilder(positionStore)
	syncedCount := 0
	for _, trade := range trades {
		// Check if trade already exists (use exchangeID which is UUID, not exchange type)
		existing, err := orderStore.GetOrderByExchangeID(exchangeID, trade.OrderId)
		if err == nil && existing != nil {
			continue // Order already exists, skip
		}
		// Normalize symbol
		symbol := market.Normalize(trade.ContractCode)
		// Determine position side from order action
		positionSide := "LONG"
		if strings.Contains(trade.PositionSide, "short") {
			positionSide = "SHORT"
		}
		// Normalize side for storage
		side := strings.ToUpper(trade.Side)
		// Create order record - use UTC time in milliseconds to avoid timezone issues
		execTimeMsInt, _ := strconv.ParseInt(trade.CreatedTime, 10, 64)
		execTimeMs := time.UnixMilli(execTimeMsInt).UTC().UnixMilli()
		execPrice, _ := strconv.ParseFloat(trade.TradePrice, 64)
		quantity, _ := strconv.ParseFloat(trade.TradeVolume, 64)
		profit, _ := strconv.ParseFloat(trade.Profit, 64)
		orderRecord := &store.TraderOrder{
			TraderID:        traderID,
			ExchangeID:      exchangeID,    // UUID
			ExchangeType:    exchangeType,  // Exchange type
			ExchangeOrderID: trade.OrderId, // Use ExecID as unique identifier
			Symbol:          symbol,
			Side:            side,
			PositionSide:    trade.PositionSide, // HTX uses one-way position mode
			Type:            trade.OrderType,
			OrderAction:     trade.PositionSide,
			Quantity:        quantity,
			Price:           execPrice,
			Status:          "FILLED",
			FilledQuantity:  quantity,
			AvgFillPrice:    execPrice,
			Commission:      profit,
			FilledAt:        execTimeMs,
			CreatedAt:       execTimeMs,
			UpdatedAt:       execTimeMs,
		}
		// Insert order record
		if err := orderStore.CreateOrder(orderRecord); err != nil {
			logger.Infof("  ⚠️ Failed to sync trade %s: %v", trade.OrderId, err)
			continue
		}
		turnover, _ := strconv.ParseFloat(trade.TradeTurnover, 64)
		feeDC, _ := decimal.NewFromString(trade.TradeFee)
		profitDC, _ := decimal.NewFromString(trade.Profit)
		turnoverDC, _ := decimal.NewFromString(trade.TradeTurnover)
		closedPnL := profitDC.Sub(feeDC).Div(turnoverDC)
		isMaker := strings.EqualFold(trade.Role, "maker")
		// Create fill record - use UTC time
		fillRecord := &store.TraderFill{
			TraderID:        traderID,
			ExchangeID:      exchangeID,   // UUID
			ExchangeType:    exchangeType, // Exchange type
			OrderID:         orderRecord.ID,
			ExchangeOrderID: trade.OrderId,
			ExchangeTradeID: trade.TradeId,
			Symbol:          symbol,
			Side:            side,
			Price:           execPrice,
			Quantity:        quantity,
			QuoteQuantity:   turnover,
			Commission:      profit,
			CommissionAsset: "USDT",
			RealizedPnL:     closedPnL.InexactFloat64(),
			IsMaker:         isMaker,
			CreatedAt:       execTimeMs,
		}
		if err := orderStore.CreateFill(fillRecord); err != nil {
			logger.Infof("  ⚠️ Failed to sync fill for trade %s: %v", trade.TradeId, err)
		}
		// Create/update position record using PositionBuilder
		if err := posBuilder.ProcessTrade(
			traderID, exchangeID, exchangeType,
			symbol, positionSide, trade.PositionSide,
			quantity, execPrice, feeDC.InexactFloat64(), closedPnL.InexactFloat64(),
			execTimeMs, trade.TradeId,
		); err != nil {
			logger.Infof("  ⚠️ Failed to sync position for trade %s: %v", trade.TradeId, err)
		} else {
			logger.Infof("  📍 Position updated for trade: %s (action: %s, qty: %.6f)", trade.TradeId, trade.PositionSide, quantity)
		}
		syncedCount++
		logger.Infof("  ✅ Synced trade: %s %s %s qty=%.6f price=%.6f pnl=%.2f fee=%.6f action=%s",
			trade.TradeId, symbol, side, quantity, execPrice, closedPnL.InexactFloat64(), feeDC.InexactFloat64(), trade.PositionSide)
	}

	logger.Infof("✅ HTX order sync completed: %d new trades synced", syncedCount)
	return nil
}
