package graph

import "github.com/comerc/budva43/app/dto/gql/dto"

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type facadeGQL interface {
	GetStatus() (*dto.Status, error)
}

type Resolver struct {
	Facade facadeGQL
}
