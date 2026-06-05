package domain

type Category struct {
	ID       string
	ParentID *string
	Name     string
}
