package api

import (
	"fafnir/portfolio-service/internal/db/generated"
	"fmt"

	portfoliopb "fafnir/shared/pb/portfolio"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertAccountTypeToDB(t portfoliopb.AccountType) generated.AccountType {
	switch t {
	case portfoliopb.AccountType_ACCOUNT_TYPE_SAVINGS:
		return generated.AccountTypeSavings
	case portfoliopb.AccountType_ACCOUNT_TYPE_INVESTMENT:
		return generated.AccountTypeInvestment
	case portfoliopb.AccountType_ACCOUNT_TYPE_CHEQUING:
		return generated.AccountTypeChequing
	default:
		return generated.AccountTypeInvestment
	}
}

func convertCurrencyTypeToDB(t portfoliopb.CurrencyType) generated.CurrencyType {
	switch t {
	case portfoliopb.CurrencyType_CURRENCY_TYPE_USD:
		return generated.CurrencyTypeUSD
	case portfoliopb.CurrencyType_CURRENCY_TYPE_CAD:
		return generated.CurrencyTypeCAD
	default:
		return generated.CurrencyTypeUSD
	}
}

func convertAccountToProto(a generated.Account) *portfoliopb.Account {
	bal, _ := a.Balance.Float64Value()
	return &portfoliopb.Account{
		Id:            a.ID.String(),
		UserId:        a.UserID.String(),
		AccountNumber: a.AccountNumber,
		Type:          convertAccountTypeFromDB(a.AccountType),
		Currency:      convertCurrencyTypeFromDB(a.Currency),
		Balance:       bal.Float64,
		CreatedAt:     convertTime(a.CreatedAt),
		UpdatedAt:     convertTime(a.UpdatedAt),
	}
}

func convertAccountTypeFromDB(t generated.AccountType) portfoliopb.AccountType {
	switch t {
	case generated.AccountTypeSavings:
		return portfoliopb.AccountType_ACCOUNT_TYPE_SAVINGS
	case generated.AccountTypeInvestment:
		return portfoliopb.AccountType_ACCOUNT_TYPE_INVESTMENT
	case generated.AccountTypeChequing:
		return portfoliopb.AccountType_ACCOUNT_TYPE_CHEQUING
	default:
		return portfoliopb.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	}
}

func convertCurrencyTypeFromDB(t generated.CurrencyType) portfoliopb.CurrencyType {
	switch t {
	case generated.CurrencyTypeUSD:
		return portfoliopb.CurrencyType_CURRENCY_TYPE_USD
	case generated.CurrencyTypeCAD:
		return portfoliopb.CurrencyType_CURRENCY_TYPE_CAD
	default:
		return portfoliopb.CurrencyType_CURRENCY_TYPE_UNSPECIFIED
	}
}

func convertHoldingToProto(h generated.Holding) *portfoliopb.Holding {
	qty, _ := h.Quantity.Float64Value()
	avg, _ := h.AvgCost.Float64Value()

	return &portfoliopb.Holding{
		Id:        h.ID.String(),
		AccountId: h.AccountID.String(),
		Symbol:    h.Symbol,
		Quantity:  qty.Float64,
		AvgCost:   avg.Float64,
		CreatedAt: convertTime(h.CreatedAt),
		UpdatedAt: convertTime(h.UpdatedAt),
	}
}

func floatToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	s := fmt.Sprintf("%f", f)
	if err := n.Scan(s); err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return n
}

func floatToNumericNullIfZero(f float64) pgtype.Numeric {
	if f == 0 {
		return pgtype.Numeric{Valid: false}
	}
	return floatToNumeric(f)
}

func convertTime(t pgtype.Timestamptz) *timestamppb.Timestamp {
	if !t.Valid {
		return nil
	}
	return timestamppb.New(t.Time)
}
