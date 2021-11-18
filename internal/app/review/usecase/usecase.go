package usecase

import (
	"github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/models"
	"github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/review"
)

type UseCase struct {
	repositoryReview review.Repository
}

func NewReviewUseCase(repositoryReview review.Repository) *UseCase {
	return &UseCase{
		repositoryReview: repositoryReview,
	}
}


func (uc *UseCase) AddReview(review models.Review) error {
	err := uc.repositoryReview.AddReview(review)
	if err != nil {
		return err
	}

	return nil
}

func (uc *UseCase) GetReviewsByProductId(productId int) ([]models.Review, error) {
	reviews, err := uc.repositoryReview.GetReviewsByProductId(productId)
	if err != nil {
		return nil, err
	}

	return reviews, nil
}

/*
func (uc *UseCase) GetBasket(userId int) ([]models.BasketProduct, error) {
	basketProducts, err := uc.repositoryBasket.GetBasket(userId)
	if err != nil {
		return nil, err
	}

	return basketProducts, nil
}

func (uc *UseCase) DropBasket(userId int) error {
	err := uc.repositoryBasket.DropBasket(userId)
	if err != nil {
		return err
	}

	return nil
}

func (uc *UseCase) DeleteProduct(product models.BasketProduct) error {
	err := uc.repositoryBasket.DeleteProduct(product)
	if err != nil {
		return err
	}

	return nil
}
*/