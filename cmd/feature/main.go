package main

import (
	"context"
	"fmt"
	"log"

	// This is required to register the hooks from your schema.
	_ "github.com/jorge-j1m/hackspark_server/ent/runtime"
	_ "github.com/lib/pq"

	"github.com/jorge-j1m/hackspark_server/ent"
	"github.com/jorge-j1m/hackspark_server/ent/usertechnology"
)

func main() {
	// 1. DATABASE CONNECTION & MIGRATION
	// ==========================================
	client, err := ent.Open("postgres", "host=localhost port=5432 user=postgres dbname=hackspark password=postgres sslmode=disable")
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	// Run the auto migration tool.
	// NOTE: In a production environment, it's recommended to use
	// versioned migrations instead of auto-migration.
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	userID := "usr_01k66j586mean8s632an447zms"

	// Get all users technologies:
	userTechs, err := client.UserTechnology.
		Query().
		Where(
			usertechnology.UserID(userID),
		).
		WithTechnology(). // Load related Tag/Technology data
		WithUser().       // Optionally load user data
		All(ctx)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("User Technologies:")
	fmt.Println("----------------------")
	fmt.Println("Total Technologies:", len(userTechs))
	fmt.Println("----------------------")

	// Access all the information
	for _, ut := range userTechs {
		tech := ut.Edges.Technology
		user := ut.Edges.User

		fmt.Printf("User: %s %s\n", user.FirstName, user.LastName)
		fmt.Printf("Technology: %s (%s)\n", tech.Name, tech.Category)
		fmt.Printf("Skill Level: %s\n", ut.SkillLevel)
		fmt.Printf("Years Experience: %f\n", *ut.YearsExperience)
		fmt.Println("---")
	}

}
