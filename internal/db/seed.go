package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ad9311/ninete/internal/logic"
	"github.com/ad9311/ninete/internal/prog"
	"github.com/ad9311/ninete/internal/repo"
)

func RunSeeds() error {
	app, err := prog.Load()
	if err != nil {
		return err
	}

	sqlDB, err := Open()
	if err != nil {
		return err
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			app.Logger.Errorf("failed to close database: %v", err)
		}
	}()

	queries := repo.New(app, sqlDB)

	store := logic.New(app, queries)

	seeds := []struct {
		name       string
		f          func(*logic.Store) error
		skipOnProd bool
	}{
		{
			"user",
			seedUsers,
			true,
		},
		{
			"category",
			seedCategories,
			false,
		},
		{
			"admin",
			func(s *logic.Store) error { return seedAdmin(s, queries) },
			true,
		},
	}

	for _, s := range seeds {
		if s.skipOnProd && app.IsProduction() {
			continue
		}

		if err := s.f(store); err != nil {
			return fmt.Errorf("failed to run %s seeds, %w", s.name, err)
		}
	}

	return nil
}

func newContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()

	return context.WithTimeout(ctx, 30*time.Second)
}

func seedUsers(s *logic.Store) error {
	ctx, cancel := newContext()
	defer cancel()

	commonPassword := "123456789"

	type devUser struct {
		username string
		email    string
	}

	devUsers := []devUser{
		{"admin", "admin@ninete.com"},
		{"john", "john@email.com"},
		{"mario", "mario@email.com"},
		{"claudia", "claudia@email.com"},
	}

	for _, u := range devUsers {
		hashedPassword, err := logic.HashPassword(commonPassword)
		if err != nil {
			return err
		}

		if _, err := s.CreateUser(ctx, repo.InsertUserParams{
			Username:     u.username,
			Email:        u.email,
			PasswordHash: hashedPassword,
		}); err != nil && err.Error() != "UNIQUE constraint failed: users.email" {
			return err
		}
	}

	return nil
}

func CategoryNames() []string {
	return []string{
		"Housing",
		"Transportation",
		"Groceries",
		"Food Delivery",
		"Healthcare",
		"Personal Care",
		"Entertainment",
		"Shopping",
		"Online Shopping",
		"Travel",
		"Financial",
		"Pets",
		"Taxes",
		"Subscriptions",
		"Other",
		"Utilities",
		"Restaurants",
	}
}

func seedCategories(s *logic.Store) error {
	ctx, cancel := newContext()
	defer cancel()

	for _, c := range CategoryNames() {
		_, err := s.CreateCategory(ctx, c)
		if err != nil && !strings.Contains(err.Error(), "UNIQUE constraint failed: categories.") {
			return err
		}
	}

	return nil
}

func seedAdmin(s *logic.Store, q repo.Queries) error {
	ctx, cancel := newContext()
	defer cancel()

	adminUser, err := q.SelectUserByEmail(ctx, "admin@ninete.com")
	if err != nil {
		return fmt.Errorf("failed to get admin user for seeding: %w", err)
	}

	if err := seedExpenses(s, adminUser.ID); err != nil {
		return err
	}

	return seedListWithTasks(s, adminUser.ID)
}

func seedExpenses(s *logic.Store, userID int) error {
	ctx, cancel := newContext()
	defer cancel()

	filters := repo.Filters{
		FilterFields: []repo.FilterField{
			{Name: "user_id", Value: userID, Operator: "="},
		},
	}

	count, err := s.CountExpenses(ctx, filters)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	descriptions := []string{
		"Rent", "Electricity bill", "Grocery run", "Netflix", "Spotify",
		"Gym membership", "Restaurant dinner", "Taxi ride", "Coffee shop", "Flight ticket",
		"Hotel stay", "Doctor visit", "Pharmacy", "Haircut", "New shoes", "Books",
		"Online course", "Cloud storage", "Mobile plan", "Internet bill",
		"Home insurance", "Car insurance", "Fuel", "Parking", "Car wash",
		"Concert ticket", "Movie night", "Streaming service", "Magazine subscription",
		"Pet food", "Vet visit", "Clothing", "Accessories", "Kitchen supplies",
		"Cleaning products", "Laundry", "Home repair", "Gardening", "Yoga class",
		"Lunch delivery", "Dinner delivery", "Breakfast out", "Snacks", "Water delivery",
		"Office supplies", "Printer ink", "Software license", "Domain renewal",
		"Birthday gift", "Charity donation",
	}

	amounts := []uint64{
		120000, 8500, 6300, 1299, 999, 4500, 7800, 2300, 450, 35000,
		18000, 5500, 2100, 2800, 9900, 1500, 4999, 299, 3500, 7200,
	}

	tagPool := []string{"essential", "discretionary", "monthly", "one-time", "recurring"}

	for i := range 50 {
		var tags []string
		if i%3 == 0 {
			tags = []string{tagPool[i%5], tagPool[(i+1)%5]}
		}

		date := time.Now().AddDate(0, -(i / 4), -(i*7)%28).Unix()

		if _, err := s.CreateExpense(ctx, userID, logic.ExpenseParams{
			ExpenseBaseParams: logic.ExpenseBaseParams{
				CategoryID:  (i % 17) + 1,
				Description: descriptions[i%len(descriptions)],
				Amount:      amounts[i%len(amounts)],
			},
			Date: date,
			Tags: tags,
		}); err != nil {
			return err
		}
	}

	return nil
}

func seedListWithTasks(s *logic.Store, userID int) error {
	ctx, cancel := newContext()
	defer cancel()

	filters := repo.Filters{
		FilterFields: []repo.FilterField{
			{Name: "user_id", Value: userID, Operator: "="},
		},
	}

	count, err := s.CountLists(ctx, filters)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	list, err := s.CreateList(ctx, userID, logic.ListParams{Name: "Testing List"})
	if err != nil {
		return err
	}

	taskDescriptions := []string{
		"Write unit tests", "Fix login bug", "Update README", "Review pull request",
		"Deploy to staging", "Set up CI pipeline", "Refactor auth module",
		"Design database schema", "Add pagination", "Implement search",
		"Write API docs", "Optimize queries", "Add error handling", "Create seed data",
		"Set up monitoring", "Configure alerts", "Write integration tests",
		"Update dependencies", "Add rate limiting", "Implement caching",
		"Buy groceries", "Call the dentist", "Pay electricity bill", "Schedule car service",
		"Renew passport", "Book flight", "Plan birthday party", "Organize closet",
		"Back up laptop", "Cancel unused subscriptions",
		"Read Clean Code", "Watch Go tutorial", "Practice Spanish", "Meditate daily",
		"Go for a run", "Cook new recipe", "Journal entry", "Water the plants",
		"Clean the kitchen", "Call mom",
		"Submit expense report", "Prepare presentation", "Send weekly update",
		"Schedule team meeting", "Review sprint backlog", "Update project timeline",
		"Write post-mortem", "Document new feature", "Onboard new teammate",
		"Archive old projects",
	}

	tagPool := []string{"urgent", "low-priority", "personal", "work", "blocked"}

	for i := range 50 {
		priority := (i % 3) + 1

		var tags []string
		if i%3 == 0 {
			tags = []string{tagPool[i%5], tagPool[(i+1)%5]}
		}

		task, err := s.CreateTask(ctx, list.ID, userID, logic.TaskParams{
			Description: taskDescriptions[i%len(taskDescriptions)],
			Priority:    priority,
			Tags:        tags,
		})
		if err != nil {
			return err
		}

		if i%4 == 0 {
			if _, err := s.ToggleTaskDone(ctx, task.ID, userID); err != nil {
				return err
			}
		}
	}

	return nil
}
