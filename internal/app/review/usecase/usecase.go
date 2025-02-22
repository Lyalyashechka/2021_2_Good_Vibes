package usecase

import (
	"errors"
	customErrors "github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/errors"
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
	oldReview, err := uc.repositoryReview.GetReviewByUserAndProduct(review.UserId, review.ProductId)
	if err != nil {
		return err
	}

	if oldReview.UserId != 0 {
		return errors.New(customErrors.REVIEW_EXISTS_DESCR)
	}

	var totalRating, ratingsCount int

	productRatings, err := uc.repositoryReview.GetAllRatingsOfProduct(review.ProductId)
	if err != nil {
		return err
	}
	for _, rating := range productRatings {
		totalRating += rating.Rating * rating.Count
		ratingsCount += rating.Count
	}

	productRating := float64(totalRating+review.Rating) / float64(ratingsCount+1)

	err = uc.repositoryReview.AddReview(review, productRating)
	if err != nil {
		return err
	}

	return nil
}

func (uc *UseCase) UpdateReview(review models.Review) error {
	oldReview, err := uc.repositoryReview.GetReviewByUserAndProduct(review.UserId, review.ProductId)
	if err != nil {
		return err
	}

	if oldReview.UserId == 0 {
		return errors.New(customErrors.NO_REVIEW_DESCR)
	}

	var totalRating, ratingsCount int

	productRatings, err := uc.repositoryReview.GetAllRatingsOfProduct(review.ProductId)
	if err != nil {
		return err
	}
	for _, rating := range productRatings {
		totalRating += rating.Rating * rating.Count
		ratingsCount += rating.Count
	}

	productRating := float64(totalRating-oldReview.Rating+review.Rating) / float64(ratingsCount)

	err = uc.repositoryReview.UpdateReview(review, productRating)
	if err != nil {
		return err
	}

	return nil
}

func (uc *UseCase) DeleteReview(userId int, productId int) error {
	oldReview, err := uc.repositoryReview.GetReviewByUserAndProduct(userId, productId)
	if err != nil {
		return err
	}

	if oldReview.UserId == 0 {
		return errors.New(customErrors.NO_REVIEW_DESCR)
	}

	var totalRating, ratingsCount int

	productRatings, err := uc.repositoryReview.GetAllRatingsOfProduct(productId)
	if err != nil {
		return err
	}
	for _, rating := range productRatings {
		totalRating += rating.Rating * rating.Count
		ratingsCount += rating.Count
	}

	productRating := float64(0)
	if ratingsCount > 1 {
		productRating = float64(totalRating-oldReview.Rating) / float64(ratingsCount-1)
	}
	err = uc.repositoryReview.DeleteReview(userId, productId, productRating)
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

func (uc *UseCase) GetReviewsByUser(userName string) ([]models.Review, error) {
	reviews, err := uc.repositoryReview.GetReviewsByUser(userName)
	if err != nil {
		return nil, err
	}

	return reviews, nil
}
