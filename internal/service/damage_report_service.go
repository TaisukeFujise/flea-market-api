package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type DamageReportOrderRepository interface {
	FindByID(ctx context.Context, id string) (domain.OrderDetail, error)
}

type DamageReportRepository interface {
	Create(ctx context.Context, input domain.DamageReportCreate, uid string) error
}

type DamageReportService struct {
	orderRepo  DamageReportOrderRepository
	reportRepo DamageReportRepository
}

func NewDamageReportService(r DamageReportRepository, o DamageReportOrderRepository) *DamageReportService {
	return &DamageReportService{reportRepo: r, orderRepo: o}
}

func (s *DamageReportService) Create(ctx context.Context, orderID, uid string, input domain.DamageReportCreate) error {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.BuyerID != uid {
		return apperror.ErrForbidden.New("only buyer can report damage")
	}
	if order.Status != domain.OrderStatusCompleted {
		return apperror.ErrForbidden.New("order must be completed to report damage")
	}
	input.ProductID = order.Product.ID
	return s.reportRepo.Create(ctx, input, uid)
}
