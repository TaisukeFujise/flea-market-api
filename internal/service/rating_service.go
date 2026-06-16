package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type RatingRepository interface {
	Create(ctx context.Context, rating domain.RatingCreate) error
}

type RatingService struct {
	ratingRepo RatingRepository
	orderRepo  OrderFinder
}

func NewRatingService(r RatingRepository, o OrderFinder) *RatingService {
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
		return apperror.ErrBadRequest.New("order must be completed to submit feedback")
	}
	return s.ratingRepo.Create(ctx, domain.RatingCreate{
		OrderID: orderID,
		RaterID: uid,
		RateeID: order.SellerID,
		Score:   score,
	})
}
