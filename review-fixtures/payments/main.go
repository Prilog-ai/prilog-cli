// A small, deliberately faulty review fixture. Copy to a dedicated test repo.
package main

import "fmt"

func costPerItem(total, count int) int { return total / count }

func main() {
	defer func() {
		if failure := recover(); failure != nil {
			fmt.Printf("ERROR payment quote failed: %v\n", failure)
		}
	}()
	fmt.Println(costPerItem(100, 0))
}
