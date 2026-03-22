package tx

import (
	"context"
	"github.com/hashicorp/go-multierror"
)

type Transacter interface {
	Do(ctx context.Context) error
	Undo(ctx context.Context) error
}

func Do(doCtx, undoCtx context.Context, actors ...Transacter) error {
	var errs *multierror.Error

	var undoErrs *multierror.Error
	defer func() {
		if errs.ErrorOrNil() != nil {
			for _, actor := range actors {
				undoErrs = multierror.Append(undoErrs, actor.Undo(undoCtx))
			}
			errs = multierror.Append(errs, undoErrs)
		}
	}()

	for _, actor := range actors {
		err := actor.Do(doCtx)
		if err != nil {
			errs = multierror.Append(errs, err)
			return errs.ErrorOrNil()
		}
	}

	return errs.ErrorOrNil()
}
