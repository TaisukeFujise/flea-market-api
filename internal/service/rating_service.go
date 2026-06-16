package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type RatingRepository interface {
	Create(ctx context.Context, rating domain.Rating) error
}

type RatingOrderRepository interface {
	FindByID(ctx context.Context, id string) (domain.OrderDetail, error)
}

type RatingService struct {
	ratingRepo RatingRepository
	orderRepo  RatingOrderRepository
}

func NewRatingService(r RatingRepository, o RatingOrderRepository) *RatingService {
	return &RatingService{ratingRepo: r, orderRepo: o}
}

func (s *RatingService) Create(ctx context.Context, orderID, uid string, score int) error {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.BuyerID != uid {
		return apperror.ErrForbidden.New("only buyer can submit feedback")
	}
	if order.Status != domain.OrderStatusCompleted {
		return apperror.ErrForbidden.New("order must be completed to submit feedback")
	}
	return s.ratingRepo.Create(ctx, domain.Rating{
		OrderID: orderID,
		RaterID: uid,
		RateeID: order.SellerID,
		Score:   score,
	})
}
