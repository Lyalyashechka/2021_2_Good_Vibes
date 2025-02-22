package usecase

import (
	"github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/models"
	"github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/microservice/orders"
)

type UseCase struct {
	repositoryOrder orders.Repository
}

func NewOrderUseCase(repositoryOrder orders.Repository) *UseCase {
	return &UseCase{
		repositoryOrder: repositoryOrder,
	}
}

func (uc *UseCase) PutOrder(order models.Order) (int, float64, error) {
	productPrices, err := uc.repositoryOrder.SelectPrices(order.Products)

	if err != nil {
		return 0, 0, err
	}

	productPricesMap := make(map[int]float64, len(productPrices))
	for _, productPrice := range productPrices {
		productPricesMap[productPrice.Id] = productPrice.Price
	}

	var cost float64

	for _, product := range order.Products {
		cost += float64(product.Number) * productPricesMap[product.ProductId]
	}

	order.Cost = cost

	orderId, err := uc.repositoryOrder.PutOrder(order)
	if err != nil {
		return 0, 0, err
	}

	return orderId, cost, nil
}

func (uc *UseCase) GetAllOrders(user int) ([]models.Order, error) {
	return uc.repositoryOrder.GetAllOrders(user)
}
