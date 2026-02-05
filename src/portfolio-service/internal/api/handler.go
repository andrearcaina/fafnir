package api

import (
	"context"
	"errors"
	"fmt"
	"log"

	"fafnir/portfolio-service/internal/db"
	"fafnir/portfolio-service/internal/db/generated"
	basepb "fafnir/shared/pb/base"
	orderpb "fafnir/shared/pb/order"
	portfoliopb "fafnir/shared/pb/portfolio"
	natsC "fafnir/shared/pkg/nats"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

type PortfolioHandler struct {
	db   *db.Database
	nats *natsC.NatsClient
	portfoliopb.UnimplementedPortfolioServiceServer
}

func NewPortfolioHandler(db *db.Database, nats *natsC.NatsClient) *PortfolioHandler {
	return &PortfolioHandler{
		db:   db,
		nats: nats,
	}
}

func (h *PortfolioHandler) RegisterSubscribeHandlers() {
	_, err := h.nats.QueueSubscribe("orders.filled", "portfolio-service", "portfolio-service-durable", h.handleOrderFilled)

	if err != nil {
		log.Printf("Failed to subscribe to orders.filled: %v\n", err)
	}
}

func (h *PortfolioHandler) handleOrderFilled(msg *nats.Msg) {
	var event orderpb.OrderFilledEvent
	if err := proto.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("Failed to unmarshal OrderFilledEvent: %v", err)
		return
	}

	log.Printf("Processing OrderFilledEvent for order %s", event.OrderId)

	userId, err := uuid.Parse(event.UserId)
	if err != nil {
		log.Printf("Invalid user ID in event: %v", err)
		return
	}

	// calculate total cost/proceeds
	// use SettlementAmount if available (for multi-currency support)
	var totalSettlementValue float64
	var avgCostBasis float64

	if event.SettlementAmount > 0 {
		totalSettlementValue = event.SettlementAmount
		avgCostBasis = event.SettlementAmount / event.FillQuantity
	} else {
		// fallback for legacy events or same-currency
		totalSettlementValue = event.FillQuantity * event.FillPrice
		avgCostBasis = event.FillPrice
	}

	err = h.db.ExecMultiTx(context.Background(), func(q *generated.Queries) error {
		// first get the investment account for the user
		accounts, err := q.GetAccountByUserId(context.Background(), userId)
		if err != nil {
			return fmt.Errorf("failed to get accounts: %w", err)
		}

		// find the investment account
		// if they have multiple, we just take the first one for now
		var investmentAcc *generated.Account
		for _, acc := range accounts {
			if acc.AccountType == generated.AccountTypeInvestment {
				investmentAcc = &acc
				break
			}
		}

		// if no investment account found, return error
		if investmentAcc == nil {
			return errors.New("no investment account found for user")
		}

		// then, settle the order
		switch event.Side {
		case orderpb.OrderSide_ORDER_SIDE_BUY:
			// buy order, deduct funds
			_, err := q.UpdateAccountBalance(context.Background(), generated.UpdateAccountBalanceParams{
				ID:      investmentAcc.ID,
				Balance: floatToNumeric(-totalSettlementValue),
			})
			if err != nil {
				return fmt.Errorf("failed to deduct funds: %w", err)
			}

			// add/update holdings (for buy, quantity increases, avg cost updates)
			_, err = q.UpsertHolding(context.Background(), generated.UpsertHoldingParams{
				AccountID: investmentAcc.ID,
				Symbol:    event.Symbol,
				Quantity:  floatToNumeric(event.FillQuantity),
				AvgCost:   floatToNumeric(avgCostBasis),
			})
			if err != nil {
				return fmt.Errorf("failed to upsert holding (buy): %w", err)
			}

		case orderpb.OrderSide_ORDER_SIDE_SELL:
			// otherwise, sell order, add funds
			_, err := q.UpdateAccountBalance(context.Background(), generated.UpdateAccountBalanceParams{
				ID:      investmentAcc.ID,
				Balance: floatToNumeric(totalSettlementValue),
			})
			if err != nil {
				return fmt.Errorf("failed to add funds: %w", err)
			}

			// decrease holdings (for sell, quantity decreases, avg cost remains same)
			_, err = q.DecreaseHolding(context.Background(), generated.DecreaseHoldingParams{
				AccountID: investmentAcc.ID,
				Symbol:    event.Symbol,
				Quantity:  floatToNumeric(event.FillQuantity),
			})
			if err != nil {
				return fmt.Errorf("failed to decrease holding (sell): %w", err)
			}
		default:
			log.Printf("Order side unspecified/unknown for order %s. Skipping settlement.", event.OrderId)
			return errors.New("order side unspecified/unknown")
		}

		// audit log
		var txType generated.TransactionType
		var desc string
		if event.Side == orderpb.OrderSide_ORDER_SIDE_BUY {
			txType = generated.TransactionTypeBuy
			desc = fmt.Sprintf("Bought %f shares of %s", event.FillQuantity, event.Symbol)
		} else {
			txType = generated.TransactionTypeSell
			desc = fmt.Sprintf("Sold %f shares of %s", event.FillQuantity, event.Symbol)
		}

		var refID = uuid.MustParse(event.OrderId)
		_, err = q.InsertAuditLog(context.Background(), generated.InsertAuditLogParams{
			AccountID:       investmentAcc.ID,
			TransactionType: txType,
			Amount:          floatToNumeric(totalSettlementValue),
			Description:     desc,
			ReferenceID:     &refID,
		})
		if err != nil {
			return fmt.Errorf("failed to insert audit log: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Printf("Settlement failed for order %s: %v", event.OrderId, err)
		// we ACK even on failure because we don't have a retry/DLQ mechanism yet
		// and we want to avoid infinite redelivery loops that drain money
		_ = msg.Ack()
	} else {
		log.Printf("Settlement successful for order %s", event.OrderId)
		_ = msg.Ack()
	}
}

func (h *PortfolioHandler) CreateAccount(ctx context.Context, req *portfoliopb.CreateAccountRequest) (*portfoliopb.CreateAccountResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return &portfoliopb.CreateAccountResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	// validate (usually frontend will have buttons to select which is possible, but just in case)
	if req.Type == portfoliopb.AccountType_ACCOUNT_TYPE_UNSPECIFIED {
		return &portfoliopb.CreateAccountResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, errors.New("account type is unspecified")
	}

	if req.Currency == portfoliopb.CurrencyType_CURRENCY_TYPE_UNSPECIFIED {
		return &portfoliopb.CreateAccountResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, errors.New("currency type is unspecified")
	}

	params := generated.InsertAccountParams{
		UserID:        userId,
		AccountNumber: uuid.New().String()[0:12], // just random 12 digits for now
		AccountType:   convertAccountTypeToDB(req.Type),
		Currency:      convertCurrencyTypeToDB(req.Currency),
		Balance:       floatToNumeric(500.00), // default 500 (just a sim)
	}

	var account generated.Account

	err = h.db.ExecMultiTx(ctx, func(q *generated.Queries) error {
		var err error
		account, err = q.InsertAccount(ctx, params)
		if err != nil {
			return err
		}

		// log initial deposit
		_, err = q.InsertAuditLog(ctx, generated.InsertAuditLogParams{
			AccountID:       account.ID,
			TransactionType: generated.TransactionTypeDeposit,
			Amount:          params.Balance,
			Description:     "Initial Deposit",
			ReferenceID:     nil,
		})
		return err
	})

	if err != nil {
		log.Printf("Failed to insert account: %v", err)
		return &portfoliopb.CreateAccountResponse{
			Code: basepb.ErrorCode_INTERNAL,
		}, err
	}

	return &portfoliopb.CreateAccountResponse{
		Code:    basepb.ErrorCode_OK,
		Account: convertAccountToProto(account),
	}, nil
}

func (h *PortfolioHandler) DeleteAccount(ctx context.Context, req *portfoliopb.DeleteAccountRequest) (*portfoliopb.DeleteAccountResponse, error) {
	accountId, err := uuid.Parse(req.AccountId)
	if err != nil {
		return &portfoliopb.DeleteAccountResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	if err := h.db.GetQueries().DeleteAccount(ctx, accountId); err != nil {
		return &portfoliopb.DeleteAccountResponse{
			Code: basepb.ErrorCode_INTERNAL,
		}, err
	}

	return &portfoliopb.DeleteAccountResponse{
		Code: basepb.ErrorCode_OK,
	}, nil
}

func (h *PortfolioHandler) GetPortfolioSummary(ctx context.Context, req *portfoliopb.GetPortfolioSummaryRequest) (*portfoliopb.GetPortfolioSummaryResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return &portfoliopb.GetPortfolioSummaryResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	accounts, err := h.db.GetQueries().GetAccountByUserId(ctx, userId)
	if err != nil {
		return &portfoliopb.GetPortfolioSummaryResponse{
			Code: basepb.ErrorCode_INTERNAL,
		}, err
	}

	var protoAccounts []*portfoliopb.Account
	var totalBal float64

	for _, acc := range accounts {
		protoAccounts = append(protoAccounts, convertAccountToProto(acc))
		bal, _ := acc.Balance.Float64Value()
		totalBal += bal.Float64
	}

	return &portfoliopb.GetPortfolioSummaryResponse{
		Code:         basepb.ErrorCode_OK,
		Accounts:     protoAccounts,
		TotalBalance: totalBal,
	}, nil
}

func (h *PortfolioHandler) GetHoldings(ctx context.Context, req *portfoliopb.GetHoldingsRequest) (*portfoliopb.GetHoldingsResponse, error) {
	accountId, err := uuid.Parse(req.AccountId)
	if err != nil {
		return &portfoliopb.GetHoldingsResponse{Code: basepb.ErrorCode_INVALID_ARGUMENT}, err
	}

	holdings, err := h.db.GetQueries().GetHoldingsByAccountId(ctx, accountId)
	if err != nil {
		return &portfoliopb.GetHoldingsResponse{Code: basepb.ErrorCode_INTERNAL}, err
	}

	var protoHoldings []*portfoliopb.Holding
	for _, holding := range holdings {
		protoHoldings = append(protoHoldings, convertHoldingToProto(holding))
	}

	return &portfoliopb.GetHoldingsResponse{
		Code:     basepb.ErrorCode_OK,
		Holdings: protoHoldings,
	}, nil
}

func (h *PortfolioHandler) GetHolding(ctx context.Context, req *portfoliopb.GetHoldingRequest) (*portfoliopb.GetHoldingResponse, error) {
	accountId, err := uuid.Parse(req.AccountId)
	if err != nil {
		return &portfoliopb.GetHoldingResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	params := generated.GetHoldingByAccountIdAndSymbolParams{
		AccountID: accountId,
		Symbol:    req.Symbol,
	}

	holding, err := h.db.GetQueries().GetHoldingByAccountIdAndSymbol(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &portfoliopb.GetHoldingResponse{
				Code: basepb.ErrorCode_NOT_FOUND,
			}, nil
		}
		return &portfoliopb.GetHoldingResponse{
			Code: basepb.ErrorCode_INTERNAL,
		}, err
	}

	return &portfoliopb.GetHoldingResponse{
		Code:    basepb.ErrorCode_OK,
		Holding: convertHoldingToProto(holding),
	}, nil
}

func (h *PortfolioHandler) GetWatchlist(ctx context.Context, req *portfoliopb.GetWatchlistRequest) (*portfoliopb.GetWatchlistResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return &portfoliopb.GetWatchlistResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	rows, err := h.db.GetQueries().GetWatchlist(ctx, userId)
	if err != nil {
		return &portfoliopb.GetWatchlistResponse{
			Code: basepb.ErrorCode_INTERNAL,
		}, err
	}

	items := make([]*portfoliopb.WatchlistItem, len(rows))
	for i, r := range rows {
		items[i] = &portfoliopb.WatchlistItem{
			Symbol:  r.Symbol,
			AddedAt: convertTime(r.CreatedAt),
		}
	}

	return &portfoliopb.GetWatchlistResponse{
		Code:  basepb.ErrorCode_OK,
		Items: items,
	}, nil
}

func (h *PortfolioHandler) AddToWatchlist(ctx context.Context, req *portfoliopb.AddToWatchlistRequest) (*portfoliopb.AddToWatchlistResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return &portfoliopb.AddToWatchlistResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	params := generated.AddToWatchlistParams{
		UserID: userId,
		Symbol: req.Symbol,
	}

	if err := h.db.GetQueries().AddToWatchlist(ctx, params); err != nil {
		return &portfoliopb.AddToWatchlistResponse{
			Code: basepb.ErrorCode_INTERNAL,
		}, err
	}

	return &portfoliopb.AddToWatchlistResponse{
		Code: basepb.ErrorCode_OK,
	}, nil
}

func (h *PortfolioHandler) RemoveFromWatchlist(ctx context.Context, req *portfoliopb.RemoveFromWatchlistRequest) (*portfoliopb.RemoveFromWatchlistResponse, error) {
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		return &portfoliopb.RemoveFromWatchlistResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	params := generated.RemoveFromWatchlistParams{
		UserID: userId,
		Symbol: req.Symbol,
	}

	if err := h.db.GetQueries().RemoveFromWatchlist(ctx, params); err != nil {
		return &portfoliopb.RemoveFromWatchlistResponse{
			Code: basepb.ErrorCode_INTERNAL,
		}, err
	}

	return &portfoliopb.RemoveFromWatchlistResponse{
		Code: basepb.ErrorCode_OK,
	}, nil
}

func (h *PortfolioHandler) GetTransactions(ctx context.Context, req *portfoliopb.GetTransactionsRequest) (*portfoliopb.GetTransactionsResponse, error) {
	accountId, err := uuid.Parse(req.AccountId)
	if err != nil {
		return &portfoliopb.GetTransactionsResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	// first check if account id exists
	if _, err := h.db.GetQueries().GetAccountById(ctx, accountId); err != nil {
		return &portfoliopb.GetTransactionsResponse{
			Code: basepb.ErrorCode_NOT_FOUND,
		}, nil
	}

	// then get transactions
	txs, err := h.db.GetQueries().GetTransactionsByAccountId(ctx, accountId)
	if err != nil {
		return &portfoliopb.GetTransactionsResponse{
			Code: basepb.ErrorCode_INTERNAL,
		}, err
	}

	var protoTxs []*portfoliopb.Transaction
	for _, tx := range txs {
		protoTx := &portfoliopb.Transaction{
			Id:          tx.ID.String(),
			AccountId:   tx.AccountID.String(),
			Type:        convertTransactionTypeToProto(tx.TransactionType),
			Amount:      numericToFloat(tx.Amount),
			Description: tx.Description,
			CreatedAt:   convertTime(tx.CreatedAt),
		}
		if tx.ReferenceID != nil {
			protoTx.ReferenceId = tx.ReferenceID.String()
		}
		protoTxs = append(protoTxs, protoTx)
	}

	return &portfoliopb.GetTransactionsResponse{
		Code:         basepb.ErrorCode_OK,
		Transactions: protoTxs,
	}, nil
}

func (h *PortfolioHandler) Deposit(ctx context.Context, req *portfoliopb.DepositRequest) (*portfoliopb.DepositResponse, error) {
	accountId, err := uuid.Parse(req.AccountId)
	if err != nil {
		return &portfoliopb.DepositResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	if req.Amount <= 0 {
		return &portfoliopb.DepositResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, errors.New("deposit amount must be positive")
	}

	var newBalance float64

	err = h.db.ExecMultiTx(ctx, func(q *generated.Queries) error {
		// verify account exists
		acc, err := q.GetAccountById(ctx, accountId)
		if err != nil {
			return err
		}

		// simple currency check (reject mismatch)
		dbCurrency := convertCurrencyTypeToProto(acc.Currency)
		if req.Currency != portfoliopb.CurrencyType_CURRENCY_TYPE_UNSPECIFIED && req.Currency != dbCurrency {
			return fmt.Errorf("currency mismatch: account is %s, deposit is %s", dbCurrency, req.Currency)
		}

		// update balance
		updatedAcc, err := q.UpdateAccountBalance(ctx, generated.UpdateAccountBalanceParams{
			ID:      accountId,
			Balance: floatToNumeric(req.Amount),
		})
		if err != nil {
			return err
		}

		bal, _ := updatedAcc.Balance.Float64Value()
		newBalance = bal.Float64

		// insert audit log
		_, err = q.InsertAuditLog(ctx, generated.InsertAuditLogParams{
			AccountID:       accountId,
			TransactionType: generated.TransactionTypeDeposit,
			Amount:          floatToNumeric(req.Amount),
			Description:     "Manual Deposit",
		})
		return err
	})

	if err != nil {
		return &portfoliopb.DepositResponse{Code: basepb.ErrorCode_INTERNAL}, err
	}

	return &portfoliopb.DepositResponse{
		Code:       basepb.ErrorCode_OK,
		NewBalance: newBalance,
	}, nil
}

func (h *PortfolioHandler) Transfer(ctx context.Context, req *portfoliopb.TransferRequest) (*portfoliopb.TransferResponse, error) {
	fromId, err := uuid.Parse(req.FromAccountId)
	if err != nil {
		return &portfoliopb.TransferResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}
	toId, err := uuid.Parse(req.ToAccountId)
	if err != nil {
		return &portfoliopb.TransferResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, err
	}

	if req.Amount <= 0 {
		return &portfoliopb.TransferResponse{
			Code: basepb.ErrorCode_INVALID_ARGUMENT,
		}, errors.New("transfer amount must be positive")
	}

	err = h.db.ExecMultiTx(ctx, func(q *generated.Queries) error {
		// get both accounts
		fromAcc, err := q.GetAccountById(ctx, fromId)
		if err != nil {
			return fmt.Errorf("from_account not found: %w", err)
		}
		toAcc, err := q.GetAccountById(ctx, toId)
		if err != nil {
			return fmt.Errorf("to_account not found: %w", err)
		}

		// check currencies (must match for now)
		fromCurr := convertCurrencyTypeToProto(fromAcc.Currency)
		toCurr := convertCurrencyTypeToProto(toAcc.Currency)

		if req.Currency != portfoliopb.CurrencyType_CURRENCY_TYPE_UNSPECIFIED {
			if req.Currency != fromCurr {
				return fmt.Errorf("currency mismatch: from_account is %s, req is %s", fromCurr, req.Currency)
			}
		}

		if fromCurr != toCurr {
			return fmt.Errorf("cross-currency transfer not supported yet (%s -> %s)", fromCurr, toCurr)
		}

		// check balance
		currentBal, _ := fromAcc.Balance.Float64Value()
		if currentBal.Float64 < req.Amount {
			return errors.New("insufficient funds")
		}

		// deduct from source
		_, err = q.UpdateAccountBalance(ctx, generated.UpdateAccountBalanceParams{
			ID:      fromId,
			Balance: floatToNumeric(-req.Amount),
		})
		if err != nil {
			return err
		}

		// add to destination
		_, err = q.UpdateAccountBalance(ctx, generated.UpdateAccountBalanceParams{
			ID:      toId,
			Balance: floatToNumeric(req.Amount),
		})
		if err != nil {
			return err
		}

		// insert audit logs
		// outgoing
		_, err = q.InsertAuditLog(ctx, generated.InsertAuditLogParams{
			AccountID:       fromId,
			TransactionType: generated.TransactionTypeTransferOut,
			Amount:          floatToNumeric(req.Amount), // positive amount
			Description:     fmt.Sprintf("Transfer to %s", toAcc.AccountNumber),
			ReferenceID:     &toId, // referencing other account ID
		})
		if err != nil {
			return err
		}

		// incoming
		_, err = q.InsertAuditLog(ctx, generated.InsertAuditLogParams{
			AccountID:       toId,
			TransactionType: generated.TransactionTypeTransferIn,
			Amount:          floatToNumeric(req.Amount), // positive amount
			Description:     fmt.Sprintf("Transfer from %s", fromAcc.AccountNumber),
			ReferenceID:     &fromId, // referencing other account ID
		})
		return err
	})

	if err != nil {
		if err.Error() == "insufficient funds" {
			return &portfoliopb.TransferResponse{
				Code: basepb.ErrorCode_INTERNAL,
			}, err
		}
		return &portfoliopb.TransferResponse{
			Code: basepb.ErrorCode_INTERNAL,
		}, err
	}

	return &portfoliopb.TransferResponse{
		Code: basepb.ErrorCode_OK,
	}, nil
}
