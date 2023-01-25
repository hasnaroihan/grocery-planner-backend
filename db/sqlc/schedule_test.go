package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createRandomSchedule(t *testing.T) Schedule {
	schedule, err := testQueries.CreateSchedule(
		context.Background(),
		uuid.NullUUID{},
	)
	require.NoError(t, err)
	require.NotEmpty(t, schedule)

	require.NotZero(t, schedule.ID)
	require.NotZero(t, schedule.CreatedAt)
	require.Zero(t, schedule.Author.UUID)
	require.False(t, schedule.Author.Valid)

	return schedule
}

func TestCreateSchedule(t *testing.T) {
	schedule, err := testQueries.CreateSchedule(
		context.Background(),
		uuid.NullUUID{},
	)
	require.NoError(t, err)
	require.NotEmpty(t, schedule)

	require.NotZero(t, schedule.ID)
	require.NotZero(t, schedule.CreatedAt)
	require.Zero(t, schedule.Author.UUID)
	require.False(t, schedule.Author.Valid)
}

func TestCreateScheduleRecipe(t *testing.T) {
	recipeNew := CreateRandomRecipe(t)
	scheduleNew := createRandomSchedule(t)

	arg := CreateScheduleRecipeParams {
		ScheduleID: scheduleNew.ID,
		RecipeID: recipeNew.ID,
		Portion: 4,
	}

	scheduleRecipe, err := testQueries.CreateScheduleRecipe(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, scheduleRecipe)
	require.Equal(t, arg.ScheduleID, scheduleRecipe.ScheduleID)
	require.Equal(t, arg.RecipeID, scheduleRecipe.RecipeID)
	require.Equal(t, arg.Portion, scheduleRecipe.Portion)
}

func TestDeleteSchedule(t *testing.T) {
	scheduleNew := createRandomSchedule(t)

	err := testQueries.DeleteSchedule(
		context.Background(),
		scheduleNew.ID,
	)
	require.NoError(t, err)

	// Test read deleted schedule
	schedule, err := testQueries.GetSchedule(
		context.Background(),
		scheduleNew.ID,
	)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, schedule)
}

func TestDeleteScheduleRecipe(t *testing.T) {
	recipe := CreateRandomRecipe(t)
	schedule := createRandomSchedule(t)

	arg := CreateScheduleRecipeParams {
		ScheduleID: schedule.ID,
		RecipeID: recipe.ID,
		Portion: 5,
	}
	scheduleRecipeNew, _ := testQueries.CreateScheduleRecipe(
		context.Background(),
		arg,
	)

	arg2 := DeleteScheduleRecipeParams{
		ScheduleID: scheduleRecipeNew.ScheduleID,
		RecipeID: scheduleRecipeNew.RecipeID,
	}
	err := testQueries.DeleteScheduleRecipe(
		context.Background(),
		arg2,
	)
	require.NoError(t, err)
}

func TestGetSchedule(t *testing.T) {
	scheduleNew := createRandomSchedule(t)

	schedule, err := testQueries.GetSchedule(
		context.Background(),
		scheduleNew.ID,
	)
	require.NoError(t, err)
	require.NotEmpty(t, schedule)

	require.Equal(t, scheduleNew.ID, schedule.ID)
	require.Equal(t, scheduleNew.Author, schedule.Author)
	require.WithinDuration(t, scheduleNew.CreatedAt, schedule.CreatedAt, time.Second)
}

func TestGetScheduleRecipe(t *testing.T) {
	recipe := CreateRandomRecipe(t)
	schedule := createRandomSchedule(t)
	arg := CreateScheduleRecipeParams {
		ScheduleID: schedule.ID,
		RecipeID: recipe.ID,
		Portion: 5,
	}
	scheduleRecipeNew, _ := testQueries.CreateScheduleRecipe(
		context.Background(),
		arg,
	)

	scheduleRecipe, err := testQueries.GetScheduleRecipe(
		context.Background(),
		scheduleRecipeNew.ScheduleID,
	)
	require.NoError(t, err)
	require.NotEmpty(t, scheduleRecipe)
	require.Len(t, scheduleRecipe, 1)
	
	for _,row := range scheduleRecipe {
		require.Equal(t, scheduleRecipeNew.ScheduleID, row.ScheduleID)
		require.Equal(t, scheduleRecipeNew.RecipeID, row.RecipeID)
		require.Equal(t, scheduleRecipeNew.Portion, row.Portion)
		require.Equal(t, recipe.Name, row.Name)
	}
}

func TestListSchedules(t *testing.T) {
	for i := 0; i < 4; i++ {
		createRandomSchedule(t)
	}

	arg := ListSchedulesParams {
		Limit: 2,
		Offset: 2,
	}
	schedule, err := testQueries.ListSchedules(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, schedule)
	require.Len(t, schedule, int(arg.Limit))

	for _,row := range schedule {
		require.NotEmpty(t, row)
	}
}

func TestListGroceries(t *testing.T) {
	recipe, recipeIngredients := CreateRandomRecipeIngredient(t)
	schedule := createRandomSchedule(t)

	arg := CreateScheduleRecipeParams {
		ScheduleID: schedule.ID,
		RecipeID: recipe.ID,
		Portion: 4,
	}
	testQueries.CreateScheduleRecipe(
		context.Background(),
		arg,
	)

	groceries, err := testQueries.ListGroceries(
		context.Background(),
		schedule.ID,
	)
	require.NoError(t, err)
	require.NotEmpty(t, groceries)
	require.Len(t, groceries, 1)

	require.Equal(t, recipeIngredients.IngredientID, groceries[0].ID)
}