package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type OrderProductRepository interface {
	GetByID(ctx context.Context, id string, uid *string) (domain.ProductDetail, error)
}

type OrderRepository interface {
	Create(ctx context.Context, buyerID, productID string, price int) (domain.Order, error)
}

type OrderService struct {
	orderRepo   OrderRepository
	productRepo OrderProductRepository
}

func NewOrderService(o OrderRepository, p OrderProductRepository) *OrderService {
	return &OrderService{orderRepo: o, productRepo: p}
}

func (s *OrderService) BuyProduct(ctx context.Context, productID, buyerID string) (domain.Order, error) {
	product, err := s.productRepo.GetByID(ctx, productID, nil)
	if err != nil {
		return domain.Order{}, err
	}

	if product.Status != domain.StatusOnSale {
		return domain.Order{}, apperror.ErrConflict.New("product is already sold out")
	}

	if product.SellerID == buyerID {
		return domain.Order{}, apperror.ErrForbidden.New("cannot purchase your own product")
	}

	return s.orderRepo.Create(ctx, buyerID, productID, product.Price)
}
