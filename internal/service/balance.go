package service

import "seller-pages/internal/models"

type BalanceService struct{}

func NewBalanceService() *BalanceService {
	return &BalanceService{}
}

func (s *BalanceService) GetBalanceInfo() models.BalanceInfo {
	info := models.BalanceInfo{
		ShopID:            "2619f2da-b3cc-490e-81ad-105323448a78",
		Balance:           574229.23,
		Sales:             97234.1,
		Income:            85001,
		ShopRating:        4.87,
		TotalSalesCount:   24336,
		TotalRefundsCount: 442,
		MonthlyRatingGrow: 0.04,
		MonthlySalesGrow:  447,
		SalesChart: models.SaleChartDTO{
			Data: []models.SalePoint{
				{Amount: 97234.1, Period: "Сентябрь"},
				{Amount: 105000, Period: "Август"},
				{Amount: 80000, Period: "Июль"},
				{Amount: 99000, Period: "Июнь"},
				{Amount: 109000, Period: "Май"},
			},
		},
	}
	info.TotalRefundsPercent = float32(info.TotalRefundsCount) / float32(info.TotalSalesCount) * 100
	info.SalesChart.AverageSales = (info.SalesChart.Data[0].Amount + info.SalesChart.Data[1].Amount + info.SalesChart.Data[2].Amount) / 3
	return info
}
