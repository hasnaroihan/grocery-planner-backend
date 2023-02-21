package db

import (
	"context"

	"github.com/google/uuid"
)

type ScheduleRecipePortion struct {
	RecipeID int64 `json:"recipe_id"`
	Portion  int32 `json:"portion"`
}

type GenerateGroceriesParam struct {
	Author  uuid.NullUUID           `json:"author"`
	Recipes []ScheduleRecipePortion `json:"recipes"`
}

type GenerateGroceriesResult struct {
	Schedule  Schedule               `json:"schedule"`
	Recipes   []GetScheduleRecipeRow `json:"recipes"`
	Groceries []ListGroceriesRow     `json:"groceries"`
}

// Create schedule, create schedule recipes, return list of groceries
func (s *SQLStorage) GenerateGroceries(ctx context.Context, arg GenerateGroceriesParam) (GenerateGroceriesResult, error) {
	var result GenerateGroceriesResult

	err := s.execTx(ctx, func(q *Queries) error {
		var err error

		result.Schedule, err = q.CreateSchedule(ctx, arg.Author)
		if err != nil {
			return err
		}

		for _, recipe := range arg.Recipes {
			_, err = q.CreateScheduleRecipe(
				ctx,
				CreateScheduleRecipeParams{
					ScheduleID: result.Schedule.ID,
					RecipeID:   recipe.RecipeID,
					Portion:    recipe.Portion,
				},
			)
			if err != nil {
				return err
			}
		}
		result.Recipes, err = q.GetScheduleRecipe(ctx, result.Schedule.ID)
		if err != nil {
			return err
		}

		result.Groceries, err = q.ListGroceries(ctx, result.Schedule.ID)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}