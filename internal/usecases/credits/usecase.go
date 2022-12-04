package credits

import (
	"context"
	"errors"
	"fmt"

	repoIntf "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/infra/repositories/credits"
	userSvcIntf "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/usecases/users"
	localerrors "github.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

type UseCase struct {
	debitsRepo repoIntf.Repository
	userSvc    userSvcIntf.UserService
}

func NewUseCase(creditsRepo repoIntf.Repository, userService userSvcIntf.UserService) *UseCase {
	return &UseCase{
		userSvc:    userService,
		debitsRepo: creditsRepo,
	}
}

func (u *UseCase) Create(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID, credit valueObjects.Money) error {
	_, err := u.userSvc.GetBatchFromIDs(ctx, []valueObjects.UserID{userID})
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return fmt.Errorf("Credit.Create : user doesn't exists : %w", localerrors.ErrDoesntExist)
		}
		return err
	}

	Credit := entities.NewDebit(channelID, userID, credit)
	err = u.debitsRepo.Create(ctx, Credit)
	if err != nil {
		return fmt.Errorf("credits.Create: %w", err)
	}

	return nil
}

func (u *UseCase) GetAll(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.Credit, error) {
	credits, err := u.debitsRepo.GetAll(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("credits.GetAll: %w", err)
	}

	return credits, nil
}

func (u *UseCase) Get(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID) (entities.Credit, error) {
	Credit, err := u.debitsRepo.Get(ctx, channelID, userID)
	if err != nil {
		return entities.Credit{}, fmt.Errorf("credits.Get: %w", err)
	}

	return Credit, nil
}

func (u *UseCase) Delete(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID) error {
	err := u.debitsRepo.Delete(ctx, channelID, userID)
	if err != nil {
		return fmt.Errorf("credits.Delete: %w", err)
	}

	return err
}

func (u *UseCase) ClearBalances(c context.Context, channelID valueObjects.ChannelID) error {
	return u.debitsRepo.Atomic(c, func(ctx context.Context) error {
		credits, err := u.debitsRepo.GetAll(ctx, channelID)
		if err != nil {
			return fmt.Errorf("credits.ClearBalance: %w", err)
		}

		for _, Credit := range credits {
			Credit.Credit = 0
			err = u.debitsRepo.Update(ctx, Credit)
			if err != nil {
				return fmt.Errorf("credits.ClearBalance: %w", err)
			}
		}
		return nil
	})
}

func (u *UseCase) CreditTransaction(c context.Context, channelID valueObjects.ChannelID, buyer valueObjects.UserID, payers []valueObjects.UserID, credit valueObjects.Money) error {

	err := u.debitsRepo.Atomic(c, func(ctx context.Context) error {
		for _, user := range payers {

			Credit, err := u.debitsRepo.Get(ctx, channelID, user)
			if err != nil {
				return localerrors.UpdateCreditError{
					Err:  fmt.Errorf("credits.CreditTransaction: %w", err),
					User: user,
				}
			}

			Credit.Credit -= credit

			err = u.debitsRepo.Update(ctx, Credit)
			if err != nil {
				return localerrors.UpdateCreditError{
					Err:  fmt.Errorf("credits.CreditTransaction: %w", err),
					User: user,
				}
			}
		}

		Credit, err := u.debitsRepo.Get(ctx, channelID, buyer)
		if err != nil {
			return localerrors.UpdateCreditError{
				Err:  fmt.Errorf("credits.CreditTransaction: %w", err),
				User: buyer,
			}
		}

		Credit.Credit += valueObjects.Money(len(payers)) * credit

		err = u.debitsRepo.Update(ctx, Credit)
		if err != nil {
			return localerrors.UpdateCreditError{
				Err:  fmt.Errorf("credits.CreditTransaction: %w", err),
				User: buyer,
			}
		}

		return nil
	})

	return err
}
