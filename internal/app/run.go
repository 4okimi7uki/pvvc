package app

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/4okimi7uki/pvvc/internal/ga4"
	"github.com/joho/godotenv"
)

func RunMain() error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()
	propertyID := os.Getenv("PROPERTY_ID")

	client, err := ga4.New(ctx, propertyID, "./service-account.json")

	if err != nil {
		return err
	}

	report, err := client.FetchDailyPageViews(ctx, "2daysAgo", "yesterday")

	for _, r := range report.Rows {
		fmt.Printf("PV: %d, path: %s\n", r.Views, r.PagePath)
	}

	return nil
}
