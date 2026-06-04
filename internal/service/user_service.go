package service

import "context"

type UserRepository interface{}

type FirebaseClient interface {
	DeleteUser(ctx context.Context, uid string) error
}

type UserService struct {
	repo UserRepository
	fb   FirebaseClient
}

func NewUserService(r UserRepository, fb FirebaseClient) *UserService {
	return &UserService{repo: r, fb: fb}
}
