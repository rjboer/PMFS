package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	PMFS "github.com/rjboer/PMFS"
)

// listProducts prints all products in the database with their IDs.
func listProducts() {
	if len(PMFS.DB.Products) == 0 {
		fmt.Println("No products available.")
		return
	}
	fmt.Println("Products:")
	for _, p := range PMFS.DB.Products {
		fmt.Printf("%d: %s\n", p.ID, p.Name)
	}
}

// createProduct asks the user for a name, creates the product and saves the DB,
// returning the newly created product.
func createProduct(scanner *bufio.Scanner) *PMFS.ProductType {
	fmt.Print("Product name: ")
	if !scanner.Scan() {
		return nil
	}
	name := scanner.Text()
	id, err := PMFS.DB.NewProduct(PMFS.ProductData{Name: name})
	if err != nil {
		log.Printf("NewProduct: %v", err)
		return nil
	}
	if err := PMFS.DB.Save(); err != nil {
		log.Printf("Save DB: %v", err)
	}
	p := &PMFS.DB.Products[id-1]
	fmt.Printf("Created product %s (ID: %d)\n", p.Name, p.ID)
	return p
}

// selectProduct displays a menu of product names and returns the chosen product.
func selectProduct(scanner *bufio.Scanner) *PMFS.ProductType {
	if len(PMFS.DB.Products) == 0 {
		fmt.Println("No products available.")
		return nil
	}

	options := make([]string, len(PMFS.DB.Products))
	for i, p := range PMFS.DB.Products {
		options[i] = p.Name
	}

	idx, err := menuSelect("Select product", options)
	if err != nil {
		return nil
	}
	return &PMFS.DB.Products[idx]
}

// productMenu provides options for a selected product.
func productMenu(scanner *bufio.Scanner, p *PMFS.ProductType) {
	for {
		clearScreen()
		fmt.Println(heading("Product: %s", p.Name))
		options := []string{
			"Project operations",
			"Edit product",
			"Back to product list",
			"Exit",
		}
		idx, err := menuSelect("Select option", options)
		if err != nil {
			return
		}
		switch idx {
		case 0:
			projectOpsMenu(scanner, p)
		case 1:
			fmt.Print("New product name: ")
			if !scanner.Scan() {
				return
			}
			newName := scanner.Text()
			if newName != "" {
				p.Name = newName
				if _, err := PMFS.DB.ModifyProduct(PMFS.ProductData{ID: p.ID, Name: newName}); err != nil {
					log.Printf("ModifyProduct: %v", err)
				} else if err := PMFS.DB.Save(); err != nil {
					log.Printf("Save DB: %v", err)
				}
			}
		case 2:
			return
		case 3:
			fmt.Println("Goodbye!")
			os.Exit(0)
		}
	}
}
