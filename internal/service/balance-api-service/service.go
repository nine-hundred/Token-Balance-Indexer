package balance_api_service

import (
	"context"
	"onbloc/internal/repository/postgresdb"
	"onbloc/internal/response"
)

type Service struct {
	repository *postgresdb.Repository
}

func NewService(repository *postgresdb.Repository) *Service {
	return &Service{repository: repository}
}

func (s Service) GetTokenBalances(ctx context.Context, addr string) (response.BalancesResponse, error) {
	balances, err := s.repository.GetBalancesByAddress(ctx, addr)
	if err != nil {
		return response.BalancesResponse{}, err
	}

	tokenBalances := make([]response.TokenBalance, 0, len(balances))
	for _, balance := range balances {
		tokenBalances = append(tokenBalances, response.TokenBalance{
			TokenPath: balance.TokenPath,
			Amount:    balance.Amount,
		})
	}

	return response.BalancesResponse{
		Balances: tokenBalances,
	}, nil
}

func (s Service) GetAllTokenBalances(ctx context.Context, offset, limit int) (response.BalancesResponse, error) {
	balances, err := s.repository.GetAllBalances(ctx, offset, limit)
	if err != nil {
		return response.BalancesResponse{}, err
	}
	tokenBalances := make([]response.TokenBalance, 0, len(balances))
	for _, balance := range balances {
		tokenBalances = append(tokenBalances, response.TokenBalance{
			TokenPath: balance.TokenPath,
			Amount:    balance.Amount,
		})
	}

	return response.BalancesResponse{
		Balances: tokenBalances,
	}, nil
}

func (s Service) GetTokenPathBalanceByAddress(ctx context.Context, tokenPath, address string) (response.AccountBalancesResponse, error) {
	balances, err := s.repository.GetTokenPathBalanceByAddress(ctx, tokenPath, address)
	if err != nil {
		return response.AccountBalancesResponse{}, err
	}
	accountBalances := make([]response.AccountBalance, 0, len(balances))
	for _, balance := range balances {
		accountBalances = append(accountBalances, response.AccountBalance{
			Address:   balance.Address,
			TokenPath: balance.TokenPath,
			Amount:    balance.Amount,
		})
	}
	return response.AccountBalancesResponse{
		AccountBalances: accountBalances,
	}, nil
}

func (s Service) GetAllTokenPathBalances(ctx context.Context, tokenPath string) (response.AccountBalancesResponse, error) {
	balances, err := s.repository.GetAllTokenPathBalances(ctx, tokenPath)
	if err != nil {
		return response.AccountBalancesResponse{}, err
	}

	accountBalances := make([]response.AccountBalance, 0, len(balances))
	for _, balance := range balances {
		accountBalances = append(accountBalances, response.AccountBalance{
			Address:   balance.Address,
			TokenPath: balance.TokenPath,
			Amount:    balance.Amount,
		})
	}
	return response.AccountBalancesResponse{
		AccountBalances: accountBalances,
	}, nil
}

func (s Service) GetTokenTransferHistoryByAddress(ctx context.Context, address string) (response.TransfersResponse, error) {
	histories, err := s.repository.GetTokenTransferHistoryByAddress(ctx, address)
	if err != nil {
		return response.TransfersResponse{}, nil
	}

	transferHistories := make([]response.Transfer, 0, len(histories))
	for _, history := range histories {
		transferHistories = append(transferHistories, response.Transfer{
			FromAddress: history.From,
			ToAddress:   history.To,
			TokenPath:   history.PkgPath,
			Amount:      history.Amount,
		})
	}
	return response.TransfersResponse{
		Transfers: transferHistories,
	}, nil
}

func (s Service) GetAllTokenTransferHistory(ctx context.Context) (response.TransfersResponse, error) {
	histories, err := s.repository.GetTokenTransferHistories(ctx)
	if err != nil {
		return response.TransfersResponse{}, nil
	}

	transferHistories := make([]response.Transfer, 0, len(histories))
	for _, history := range histories {
		transferHistories = append(transferHistories, response.Transfer{
			FromAddress: history.From,
			ToAddress:   history.To,
			TokenPath:   history.PkgPath,
			Amount:      history.Amount,
		})
	}
	return response.TransfersResponse{
		Transfers: transferHistories,
	}, nil
}
