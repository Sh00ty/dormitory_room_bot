package credits

import (
	"context"
	"errors"
	"fmt"

	users "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/users"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	localerrors "gitlab.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
)

type creditManager struct {
	debitsRepo Repository
	resolver   Resolver
	userSvc    users.UserService
}

func NewUseCase(creditsRepo Repository, userService users.UserService, resolver Resolver) *creditManager {
	return &creditManager{
		userSvc:    userService,
		resolver:   resolver,
		debitsRepo: creditsRepo,
	}
}

func (u *creditManager) Create(ctx context.Context, channelID valueObjects.ChannelID, userID valueObjects.UserID, credit valueObjects.Money) error {
	_, err := u.userSvc.GetBatchFromIDs(ctx, []valueObjects.UserID{userID})
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return fmt.Errorf("credits.Create  %w", localerrors.ErrDoesntExist)
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

func (u *creditManager) GetAll(ctx context.Context, channelID valueObjects.ChannelID) ([]entities.Credit, error) {
	credits, err := u.debitsRepo.GetAll(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("credits.GetAll: %w", err)
	}

	return credits, nil
}

func (u *creditManager) Checkout(c context.Context, channelID valueObjects.ChannelID) ([]entities.Transaction, error) {
	transacrions := make([]entities.Transaction, 0)
	err := u.debitsRepo.Atomic(c, func(ctx context.Context) error {
		credits, err := u.debitsRepo.GetAll(ctx, channelID)
		if err != nil {
			return fmt.Errorf("credits.Checkout: %w", err)
		}

		if len(credits) == 0 {
			return fmt.Errorf("credits.Checkout: %w", localerrors.ErrDoesntExist)
		}

		for _, Credit := range credits {
			Credit.Credit = 0
			if err = u.debitsRepo.Update(ctx, Credit); err != nil {
				return fmt.Errorf("credits.Checkout: %w", err)
			}
		}
		transacrions = u.resolver.ResolveCredits(credits)
		return nil
	})
	return transacrions, err
}

func (u *creditManager) CreditTransaction(c context.Context, channelID valueObjects.ChannelID, buyer valueObjects.UserID, payers []valueObjects.UserID, credit valueObjects.Money) error {

	return u.debitsRepo.Atomic(c, func(ctx context.Context) error {
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
}
