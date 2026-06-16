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
	ListByUserID(ctx context.Context, userID string, f domain.OrderFilter) ([]domain.OrderListItem, int, error)
	FindByID(ctx context.Context, id string) (domain.OrderDetail, error)
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

func (s *OrderService) ListOrders(ctx context.Context, userID string, f domain.OrderFilter) ([]domain.OrderListItem, int, error) {
	return s.orderRepo.ListByUserID(ctx, userID, f)
}

func (s *OrderService) GetOrder(ctx context.Context, id, uid string) (domain.OrderDetail, error) {
	order, err := s.orderRepo.FindByID(ctx, id)
	if err != nil {
		return domain.OrderDetail{}, err
	}
	if order.BuyerID != uid && order.SellerID != uid {
		return domain.OrderDetail{}, apperror.ErrForbidden.New("forbidden")
	}
	return order, nil
}
