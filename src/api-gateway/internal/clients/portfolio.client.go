package clients

import (
	"context"
	"fafnir/api-gateway/graph/model"
	basepb "fafnir/shared/pb/base"
	pb "fafnir/shared/pb/portfolio"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PortfolioClient struct {
	client pb.PortfolioServiceClient
}

func NewPortfolioClient(url string) *PortfolioClient {
	conn, err := grpc.NewClient(
		url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	return &PortfolioClient{
		client: pb.NewPortfolioServiceClient(conn),
	}
}

func (c *PortfolioClient) CreateAccount(ctx context.Context, userID string, req model.CreateAccountRequest) (model.CreateAccountResponse, error) {
	accType := pb.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	if req.Type == "SAVINGS" {
		accType = pb.AccountType_ACCOUNT_TYPE_SAVINGS
	} else if req.Type == "INVESTMENT" {
		accType = pb.AccountType_ACCOUNT_TYPE_INVESTMENT
	} else if req.Type == "CHEQUING" {
		accType = pb.AccountType_ACCOUNT_TYPE_CHEQUING
	}

	curr := pb.CurrencyType_CURRENCY_TYPE_UNSPECIFIED
	if req.Currency == "USD" {
		curr = pb.CurrencyType_CURRENCY_TYPE_USD
	} else if req.Currency == "CAD" {
		curr = pb.CurrencyType_CURRENCY_TYPE_CAD
	}

	resp, err := c.client.CreateAccount(ctx, &pb.CreateAccountRequest{
		UserId:   userID,
		Type:     accType,
		Currency: curr,
	})
	if err != nil {
		return model.CreateAccountResponse{
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	return model.CreateAccountResponse{
		Data: convertAccountToModel(resp.Account),
		Code: resp.GetCode().String(),
	}, nil
}

func (c *PortfolioClient) GetPortfolioSummary(ctx context.Context, userID string) (model.GetPortfolioSummaryResponse, error) {
	resp, err := c.client.GetPortfolioSummary(ctx, &pb.GetPortfolioSummaryRequest{
		UserId: userID,
	})
	if err != nil {
		return model.GetPortfolioSummaryResponse{
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	var accounts []*model.Account
	for _, acc := range resp.Accounts {
		accounts = append(accounts, convertAccountToModel(acc))
	}

	return model.GetPortfolioSummaryResponse{
		Accounts:     accounts,
		TotalBalance: resp.TotalBalance,
		Code:         resp.GetCode().String(),
	}, nil
}

func (c *PortfolioClient) GetHoldings(ctx context.Context, req model.GetHoldingsRequest) (model.GetHoldingsResponse, error) {
	resp, err := c.client.GetHoldings(ctx, &pb.GetHoldingsRequest{
		AccountId: req.AccountID,
	})
	if err != nil {
		return model.GetHoldingsResponse{
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	var holdings []*model.Holding
	for _, h := range resp.Holdings {
		holdings = append(holdings, convertHoldingToModel(h))
	}

	return model.GetHoldingsResponse{
		Data: holdings,
		Code: resp.GetCode().String(),
	}, nil
}

func (c *PortfolioClient) GetHolding(ctx context.Context, req model.GetHoldingRequest) (model.GetHoldingResponse, error) {
	resp, err := c.client.GetHolding(ctx, &pb.GetHoldingRequest{
		AccountId: req.AccountID,
		Symbol:    req.Symbol,
	})
	if err != nil {
		return model.GetHoldingResponse{
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	return model.GetHoldingResponse{
		Data: convertHoldingToModel(resp.Holding),
		Code: resp.GetCode().String(),
	}, nil
}

func (c *PortfolioClient) GetWatchlist(ctx context.Context, userID string) (model.GetWatchlistResponse, error) {
	resp, err := c.client.GetWatchlist(ctx, &pb.GetWatchlistRequest{
		UserId: userID,
	})
	if err != nil {
		return model.GetWatchlistResponse{
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	var items []*model.WatchlistItem
	for _, item := range resp.Items {
		items = append(items, &model.WatchlistItem{
			Symbol:  item.Symbol,
			AddedAt: item.AddedAt.AsTime().String(),
		})
	}

	return model.GetWatchlistResponse{
		Data: items,
		Code: resp.GetCode().String(),
	}, nil
}

func (c *PortfolioClient) AddToWatchlist(ctx context.Context, userID string, req model.AddToWatchlistRequest) (model.AddToWatchlistResponse, error) {
	resp, err := c.client.AddToWatchlist(ctx, &pb.AddToWatchlistRequest{
		UserId: userID,
		Symbol: req.Symbol,
	})
	if err != nil {
		return model.AddToWatchlistResponse{
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	return model.AddToWatchlistResponse{
		Code: resp.GetCode().String(),
	}, nil
}

func (c *PortfolioClient) RemoveFromWatchlist(ctx context.Context, userID string, req model.RemoveFromWatchlistRequest) (model.RemoveFromWatchlistResponse, error) {
	resp, err := c.client.RemoveFromWatchlist(ctx, &pb.RemoveFromWatchlistRequest{
		UserId: userID,
		Symbol: req.Symbol,
	})
	if err != nil {
		return model.RemoveFromWatchlistResponse{
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	return model.RemoveFromWatchlistResponse{
		Code: resp.GetCode().String(),
	}, nil
}

func (c *PortfolioClient) DeleteAccount(ctx context.Context, accountID string) (bool, error) {
	resp, err := c.client.DeleteAccount(ctx, &pb.DeleteAccountRequest{
		AccountId: accountID,
	})
	if err != nil {
		return false, err
	}

	return resp.Code == basepb.ErrorCode_OK, nil
}

func (c *PortfolioClient) GetTransactions(ctx context.Context, req model.GetTransactionsRequest) (model.GetTransactionsResponse, error) {
	resp, err := c.client.GetTransactions(ctx, &pb.GetTransactionsRequest{
		AccountId: req.AccountID,
	})
	if err != nil {
		return model.GetTransactionsResponse{
			Code: basepb.ErrorCode_INTERNAL.String(),
		}, err
	}

	var txs []*model.Transaction
	for _, t := range resp.Transactions {
		txs = append(txs, &model.Transaction{
			ID:          t.Id,
			AccountID:   t.AccountId,
			Type:        t.Type,
			Amount:      t.Amount,
			Description: t.Description,
			ReferenceID: &t.ReferenceId,
			CreatedAt:   t.CreatedAt.AsTime().String(),
		})
	}

	return model.GetTransactionsResponse{
		Code: resp.Code.String(),
		Data: txs,
	}, nil
}

func convertAccountToModel(acc *pb.Account) *model.Account {
	if acc == nil {
		return nil
	}
	return &model.Account{
		ID:            acc.Id,
		UserID:        acc.UserId,
		AccountNumber: acc.AccountNumber,
		Type:          acc.Type.String(),
		Currency:      acc.Currency.String(),
		Balance:       acc.Balance,
		CreatedAt:     acc.CreatedAt.AsTime().String(),
		UpdatedAt:     acc.UpdatedAt.AsTime().String(),
	}
}

func convertHoldingToModel(h *pb.Holding) *model.Holding {
	if h == nil {
		return nil
	}
	return &model.Holding{
		ID:        h.Id,
		AccountID: h.AccountId,
		Symbol:    h.Symbol,
		Quantity:  h.Quantity,
		AvgCost:   h.AvgCost,
		CreatedAt: h.CreatedAt.AsTime().String(),
		UpdatedAt: h.UpdatedAt.AsTime().String(),
	}
}
