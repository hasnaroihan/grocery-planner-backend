package db

import (
	"context"
	"database/sql"
	"log"
)

type TxUpdateRecipeParams struct {
	Recipe UpdateRecipeParams	`json:"recipe"`
	ListIngredients []ListIngredientParam	`json:"ingredients"`
}

// Update recipe does not delete recipe_ingredients, it only update recipe details, recipe ingredient details, or create new recipe ingredients
func (s *SQLStorage) UpdateRecipeTx(ctx context.Context, arg TxUpdateRecipeParams) (RecipeResult, error) {
	var result RecipeResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.Recipe, err = q.UpdateRecipe(ctx, arg.Recipe)
		if err != nil {
			log.Print("update recipe")
			return err
		}

		for _, item := range arg.ListIngredients {
			if item.ID.Int32 > 0 {
				_, err = q.UpdateRecipeIngredient(
					ctx,
					UpdateRecipeIngredientParams{
						RecipeID: result.Recipe.ID,
						IngredientID: item.ID.Int32,
						Amount: item.Amount,
						UnitID: item.UnitID,
					},
				)
				if err != nil {
					log.Print("update recipe ingredient")
					return err
				}
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
					log.Print("create ingredient")
					return err
				}
				if err == sql.ErrNoRows {
					ingredient, err = q.SearchIngredientName(ctx, item.Name)
					if err != nil {
						log.Print("search ingredient")
						return err
					}
				}
				_, err = q.CreateRecipeIngredient(
					ctx,
					CreateRecipeIngredientParams{
						IngredientID: ingredient.ID,
						RecipeID: result.Recipe.ID,
						Amount: item.Amount,
						UnitID: item.UnitID,
					})
				if err != nil {
					log.Print("create recipe ingredient")
					return err
				}
			}
		}

		result.Ingredients, err = q.GetRecipeIngredients(ctx, result.Recipe.ID)
		if err != nil {
			log.Print("get recipe ingredients")
			return err
		}

		return nil
	})

	return result, err
}