package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ConfirmExpensiveOperation asks the user to confirm an expensive operation
func ConfirmExpensiveOperation(message string, cost string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\n⚠️  COST WARNING ⚠️\n")
	fmt.Printf("%s\n", message)
	fmt.Printf("Estimated cost: %s\n", cost)
	fmt.Printf("\nDo you want to proceed? (yes/no): ")

	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y", nil
}

// ShowCostEstimate displays a cost estimate without requiring confirmation
func ShowCostEstimate(message string, cost string) {
	fmt.Printf("\n💰 Cost Estimate: %s\n", cost)
	fmt.Printf("%s\n\n", message)
}