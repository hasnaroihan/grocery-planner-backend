package db

import (
	"context"
	"database/sql"
)

type TxUpdateRecipeParams struct {
	Recipe UpdateRecipeParams	`json:"recipe"`
	ListIngredients []ListIngredientParam	`json:"ingredients"`
}

func (s *SQLStorage) UpdateRecipeTx(ctx context.Context, arg TxUpdateRecipeParams) (RecipeResult, error) {
	var result RecipeResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.Recipe, err = q.UpdateRecipe(ctx, arg.Recipe)
		if err != nil {
			return err
		}

		for _, item := range arg.ListIngredients {
			var update UpdateRecipeIngredientParams

			if item.ID.Valid {
				update.RecipeID = result.Recipe.ID
				update.IngredientID = item.ID.Int32
				update.Amount = item.Amount
				update.UnitID = item.UnitID
			} else {
				ingredient, err := q.CreateIngredient(
					ctx,
					CreateIngredientParams{
						Name: item.Name,
						DefaultUnit: sql.NullInt32{
							Int32: item.UnitID,
							Valid: true,
						},
					},
				)

				if err != nil && err != sql.ErrNoRows {
					return err
				}
				if err == sql.ErrNoRows {
					ingredient, err = q.SearchIngredientName(ctx, item.Name)
					if err != nil {
						return err
					}
				}

				update.RecipeID = result.Recipe.ID
				update.IngredientID = ingredient.ID
				update.Amount = item.Amount
				update.UnitID = item.UnitID
			}

			_, err = q.UpdateRecipeIngredient(ctx, update)
			if err != nil {
				return err
			}
		}

		result.Ingredients, err = q.GetRecipeIngredients(ctx, result.Recipe.ID)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}