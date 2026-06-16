package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type DamageReportRepository interface {
	Create(ctx context.Context, input domain.DamageReportCreate, uid string) error
	ValidateImageForProduct(ctx context.Context, imageID, productID string) error
}

type DamageReportService struct {
	orderRepo  OrderFinder
	reportRepo DamageReportRepository
}

func NewDamageReportService(r DamageReportRepository, o OrderFinder) *DamageReportService {
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
		return apperror.ErrBadRequest.New("order must be completed to report damage")
	}
	if err := s.reportRepo.ValidateImageForProduct(ctx, input.ImageID, order.Product.ID); err != nil {
		return err
	}
	input.ProductID = order.Product.ID
	return s.reportRepo.Create(ctx, input, uid)
}
