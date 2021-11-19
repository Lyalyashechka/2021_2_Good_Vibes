package usecase

import (
	"github.com/go-park-mail-ru/2021_2_Good_Vibes/internal/app/search"
)

type UseCase struct {
	repositorySearch search.Repository
}

func NewSearchUseCase(repositorySearch search.Repository) *UseCase {
	return &UseCase{
		repositorySearch: repositorySearch,
	}
}
