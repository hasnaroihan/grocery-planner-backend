package db

import "context"

func (s *SQLStorage) GetRecipeTx(ctx context.Context, id int64) (RecipeResult, error) {
	var result RecipeResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.Recipe, err = q.GetRecipe(ctx, id)
		if err != nil {
			return err
		}

		result.Ingredients, err = q.GetRecipeIngredients(ctx, id)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}