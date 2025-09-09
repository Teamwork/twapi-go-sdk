package projects_test

import (
	"context"
	"fmt"

	"github.com/teamwork/twapi-go-sdk"
	"github.com/teamwork/twapi-go-sdk/projects"
	"github.com/teamwork/twapi-go-sdk/session"
)

func ExampleRateUserGet() {
	engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

	req := projects.NewRateUserGetRequest(12345) // User ID
	// Configure optional filters
	req.Filters.IncludeUserCost = true
	// Include supported related resources via enum
	req.Filters.Include = []projects.RateUserGetRequestSideload{
		projects.RateSideloadProjects,
	}

	resp, err := projects.RateUserGet(context.Background(), engine, req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	if resp.InstallationRate != nil {
		fmt.Printf("User installation rate: %d\n", *resp.InstallationRate)
	}
	fmt.Printf("Number of project rates: %d\n", len(resp.ProjectRates))
	fmt.Printf("Number of multi-currency rates: %d\n", len(resp.InstallationRates))

	if resp.UserCost != nil {
		fmt.Printf("User cost: %d\n", *resp.UserCost)
	}
}

func ExampleRateInstallationUserList() {
	engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

	req := projects.NewRateInstallationUserListRequest()
	req.Filters.PageSize = 10 // Get first 10 users

	next, err := twapi.Iterate[projects.RateInstallationUserListRequest, *projects.RateInstallationUserListResponse](
		context.Background(),
		engine,
		req,
	)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	var iteration int
	for {
		iteration++
		fmt.Printf("Iteration %d\n", iteration)

		resp, hasNext, err := next()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		if resp == nil {
			break
		}
		for _, userRate := range resp.UserRates {
			fmt.Printf("User %d has rate %d\n", userRate.User.ID, userRate.Rate)
		}
		if !hasNext {
			break
		}
	}
}

func ExampleRateInstallationUserGet() {
	engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

	req := projects.NewRateInstallationUserGetRequest(12345) // User ID
	req.Filters.Include = []projects.RateInstallationUserGetRequestSideload{
		projects.RateInstallationUserGetRequestSideloadCurrencies,
	}

	resp, err := projects.RateInstallationUserGet(context.Background(), engine, req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Printf("User rate: %d\n", resp.UserRate)
	fmt.Printf("Number of currency rates: %d\n", len(resp.UserRates))
}

func ExampleRateInstallationUserUpdate() {
    engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

    var rate int64 = 5000
    req := projects.NewRateInstallationUserUpdateRequest(12345, &rate) // User ID, Rate (cents)
    _, err := projects.RateInstallationUserUpdate(context.Background(), engine, req)
    if err != nil {
        fmt.Printf("Error: %s\n", err)
        return
    }

	fmt.Println("User rate updated successfully")
}

func ExampleRateInstallationUserBulkUpdate() {
    engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

    var rate int64 = 5000
    req := projects.NewRateInstallationUserBulkUpdateRequest(&rate) // Rate (cents)
    req.IDs = []int64{12345, 12346, 12347}                          // Specific user IDs to update

	resp, err := projects.RateInstallationUserBulkUpdate(context.Background(), engine, req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Printf("Updated %d users with rate %d\n", len(resp.IDs), resp.Rate)
}

func ExampleRateProjectGet() {
	engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

	req := projects.NewRateProjectGetRequest(67890) // Project ID
	req.Filters.Include = []projects.RateProjectGetRequestSideload{
		projects.RateProjectGetRequestSideloadCurrencies,
	}

	resp, err := projects.RateProjectGet(context.Background(), engine, req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

    fmt.Printf("Project rate: %d\n", resp.ProjectRate)
    fmt.Printf("Rate value: %.2f\n", resp.Rate.Amount)
}

func ExampleRateProjectUpdate() {
    engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

    var rate int64 = 7500
    req := projects.NewRateProjectUpdateRequest(67890, &rate) // Project ID, Rate (cents)
    _, err := projects.RateProjectUpdate(context.Background(), engine, req)
    if err != nil {
        fmt.Printf("Error: %s\n", err)
        return
    }

	fmt.Println("Project rate updated successfully")
}

func ExampleRateProjectAndUsersUpdate() {
    engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

    req := projects.NewRateProjectAndUsersUpdateRequest(67890, int64(7500)) // Project ID, Rate (cents)

    // Add user-specific rate exceptions
    req.UserRates = []projects.ProjectUserRateRequest{
        {
            User: twapi.Relationship{
                ID:   12345,
                Type: "user",
            },
            UserRate: int64(8000), // Higher rate for this specific user (cents)
        },
        {
            User: twapi.Relationship{
                ID:   12346,
                Type: "user",
            },
            UserRate: int64(6000), // Lower rate for this user (cents)
        },
    }

	_, err := projects.RateProjectAndUsersUpdate(context.Background(), engine, req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Println("Project and user rates updated successfully")
}

func ExampleRateProjectUserList() {
	engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

	req := projects.NewRateProjectUserListRequest(67890) // Project ID
	req.Filters.SearchTerm = "john"
	req.Filters.OrderBy = "name"
	req.Filters.OrderMode = "asc"
	req.Filters.PageSize = 20

	next, err := twapi.Iterate[projects.RateProjectUserListRequest, *projects.RateProjectUserListResponse](
		context.Background(),
		engine,
		req,
	)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	// Pull the first page and print results
	resp, _, err := next()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	if resp != nil {
		fmt.Printf("Found %d user rates for project\n", len(resp.UserRates))
		for _, userRate := range resp.UserRates {
			fmt.Printf("User %d effective rate: %d\n", userRate.User.ID, userRate.EffectiveRate)
		}
	}
}

func ExampleRateProjectUserGet() {
	engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

	req := projects.NewRateProjectUserGetRequest(67890, 12345) // Project ID, User ID
	req.Filters.Include = []projects.RateProjectUserGetRequestSideload{
		projects.RateProjectUserGetRequestSideloadCurrencies,
	}

	resp, err := projects.RateProjectUserGet(context.Background(), engine, req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

    fmt.Printf("User rate for project: %.2f\n", resp.UserRate.Amount)
    fmt.Printf("Rate value: %d\n", resp.Rate)
}

func ExampleRateProjectUserUpdate() {
    engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

    var rate int64 = 8500
    req := projects.NewRateProjectUserUpdateRequest(67890, 12345, &rate) // Project ID, User ID, Rate (cents)
    resp, err := projects.RateProjectUserUpdate(context.Background(), engine, req)
    if err != nil {
        fmt.Printf("Error: %s\n", err)
        return
    }

	fmt.Printf("Updated user rate to: %d\n", resp.UserRate)
}

func ExampleRateProjectUserHistoryGet() {
	engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

	req := projects.NewRateProjectUserHistoryGetRequest(67890, 12345) // Project ID, User ID
	req.Filters.OrderMode = "desc"                                    // Most recent first
	req.Filters.PageSize = 10

	next, err := twapi.Iterate[projects.RateProjectUserHistoryGetRequest, *projects.RateProjectUserHistoryGetResponse](
		context.Background(),
		engine,
		req,
	)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	for {
		resp, hasNext, err := next()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		if resp == nil {
			break
		}

		for _, history := range resp.UserRateHistory {
			fmt.Printf("Rate: %d", history.Rate)
			if history.FromDate != nil {
				fmt.Printf(" (effective from %s)", history.FromDate.Format("2006-01-02"))
			}
			if history.ToDate != nil {
				fmt.Printf(" (until %s)", history.ToDate.Format("2006-01-02"))
			}
			fmt.Println()
		}

		if !hasNext {
			break
		}
	}
}

// Example of working with pagination across all pages
func ExampleRateProjectUserList_pagination() {
	engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

	req := projects.NewRateProjectUserListRequest(67890)
	req.Filters.PageSize = 50

	allUserRates := []projects.EffectiveUserProjectRate{}

	next, err := twapi.Iterate[projects.RateProjectUserListRequest, *projects.RateProjectUserListResponse](
		context.Background(),
		engine,
		req,
	)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	for {
		resp, hasNext, err := next()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		if resp == nil {
			break
		}
		allUserRates = append(allUserRates, resp.UserRates...)
		if !hasNext {
			break
		}
	}

	fmt.Printf("Total user rates collected: %d\n", len(allUserRates))
}

// Example showing enhanced metadata and multi-currency features
func ExampleRateProjectUserList_metadata() {
	engine := twapi.NewEngine(session.NewBearerToken("your_token", "https://your-domain.teamwork.com"))

	req := projects.NewRateProjectUserListRequest(67890) // Project ID
	next, err := twapi.Iterate[projects.RateProjectUserListRequest, *projects.RateProjectUserListResponse](
		context.Background(),
		engine,
		req,
	)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	resp, _, err := next()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Printf("Found %d user rates for project\n", len(resp.UserRates))

	for _, userRate := range resp.UserRates {
		fmt.Printf("User %d effective rate: %d\n", userRate.User.ID, userRate.EffectiveRate)

		// Show rate source information
		if userRate.Source != nil {
			fmt.Printf("  Rate source: %s\n", *userRate.Source)
		}

		// Show temporal information
		if userRate.FromDate != nil {
			fmt.Printf("  Effective from: %s\n", userRate.FromDate.Format("2006-01-02"))
		}

		// Show update metadata
		if userRate.UpdatedAt != nil {
			fmt.Printf("  Last updated: %s\n", userRate.UpdatedAt.Format("2006-01-02"))
		}

		// Show billable rate with currency
		if userRate.BillableRate != nil {
			fmt.Printf("  Billable rate: %.2f (Currency ID: %d)\n",
				userRate.BillableRate.Rate, userRate.BillableRate.Currency.ID)
		}
	}
}
