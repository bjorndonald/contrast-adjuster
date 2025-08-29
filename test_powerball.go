package main

import (
	"fmt"
)

// testPowerballPrizes demonstrates the Powerball prize calculation system
// This function shows examples of different ticket combinations and their prizes
func testPowerballPrizes() {
	fmt.Println("=== Powerball Prize Calculation Examples ===\n")

	// Example winning numbers (mock data for testing)
	winningNumbers := &WinningNumbers{
		N1:    10,
		N2:    20,
		N3:    30,
		N4:    40,
		N5:    50,
		MBall: 25, // Powerball number
	}

	fmt.Printf("Winning Numbers: %d, %d, %d, %d, %d | Powerball: %d\n\n",
		winningNumbers.N1, winningNumbers.N2, winningNumbers.N3, winningNumbers.N4, winningNumbers.N5, winningNumbers.MBall)

	// Example tickets to test
	exampleTickets := []struct {
		description string
		whiteBalls  []int
		powerball   int
	}{
		{"Jackpot Ticket", []int{10, 20, 30, 40, 50}, 25},
		{"5 White Balls (no Powerball)", []int{10, 20, 30, 40, 50}, 15},
		{"4 White Balls + Powerball", []int{10, 20, 30, 40, 60}, 25},
		{"4 White Balls (no Powerball)", []int{10, 20, 30, 40, 60}, 15},
		{"3 White Balls + Powerball", []int{10, 20, 30, 70, 80}, 25},
		{"3 White Balls (no Powerball)", []int{10, 20, 30, 70, 80}, 15},
		{"2 White Balls + Powerball", []int{10, 20, 90, 100, 110}, 25},
		{"1 White Ball + Powerball", []int{10, 120, 130, 140, 150}, 25},
		{"Powerball Only", []int{160, 170, 180, 190, 200}, 25},
		{"No Matches", []int{160, 170, 180, 190, 200}, 15},
	}

	// Test each ticket
	for _, ticket := range exampleTickets {
		fmt.Printf("Ticket: %s\n", ticket.description)
		fmt.Printf("Numbers: %v | Powerball: %d\n", ticket.whiteBalls, ticket.powerball)

		// Check without Power Play
		result, err := checkPowerballTicket(ticket.whiteBalls, ticket.powerball, winningNumbers, 0)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Result: %s\n", result.PrizeDescription)
			if result.IsWinner {
				fmt.Printf("Prize: $%.2f\n", float64(result.BasePrize)/100)
			} else {
				fmt.Printf("Prize: No Prize\n")
			}
		}

		// Check with Power Play (2x multiplier)
		resultPP, err := checkPowerballTicket(ticket.whiteBalls, ticket.powerball, winningNumbers, 2)
		if err != nil {
			fmt.Printf("Power Play Error: %v\n", err)
		} else if resultPP.PowerPlayMultiplier > 0 {
			fmt.Printf("Power Play (2x): %s\n", resultPP.PowerPlayPrize)
		}

		fmt.Println("---")
	}

	// Demonstrate Power Play multipliers
	fmt.Println("=== Power Play Multiplier Examples ===\n")

	// Test a $100 prize with different multipliers
	basePrize := 10000 // $100 in cents
	multipliers := []int{2, 3, 4, 5, 10}

	fmt.Printf("Base Prize: $%.2f\n\n", float64(basePrize)/100)

	for _, multiplier := range multipliers {
		powerPlayPrize, powerPlayAmount := calculatePowerPlayPrize(basePrize, multiplier)
		fmt.Printf("%dx Multiplier: %s ($%.2f)\n", multiplier, powerPlayPrize, float64(powerPlayAmount)/100)
	}

	// Test $1,000,000 prize with 10x multiplier (should default to 2x)
	fmt.Printf("\n$1,000,000 Prize with 10x Multiplier (should default to 2x):\n")
	jackpotBase := 100000000 // $1,000,000 in cents
	powerPlayPrize, powerPlayAmount := calculatePowerPlayPrize(jackpotBase, 10)
	fmt.Printf("Result: %s ($%.2f)\n", powerPlayPrize, float64(powerPlayAmount)/100)
}

// This file contains test functions for the Powerball prize calculation system
// Run the main program to see the demonstration
