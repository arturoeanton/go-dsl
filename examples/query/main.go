package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/arturoeanton/go-dsl/pkg/dslbuilder"
)

// Product represents a product in our system
type Product struct {
	ID       int
	Name     string
	Category string
	Price    float64
	Stock    int
}

func main() {
	// Create a query DSL
	query := dslbuilder.New("QueryDSL")

	// Sample data
	products := []Product{
		{1, "Laptop Dell", "Electronics", 1200.00, 5},
		{2, "Mouse Logitech", "Electronics", 25.00, 50},
		{3, "Desk Chair", "Furniture", 350.00, 10},
		{4, "Standing Desk", "Furniture", 600.00, 3},
		{5, "USB Cable", "Electronics", 10.00, 100},
		{6, "Monitor 27\"", "Electronics", 400.00, 8},
		{7, "Office Lamp", "Furniture", 45.00, 20},
	}

	// Define tokens - keywords first with high priority  
	query.KeywordToken("BUSCAR", "buscar")
	query.KeywordToken("LISTAR", "listar") 
	query.KeywordToken("CONTAR", "contar")
	query.KeywordToken("PRODUCTOS", "productos")
	query.KeywordToken("DONDE", "donde")
	query.KeywordToken("ES", "es")
	query.KeywordToken("MAYOR", "mayor")
	query.KeywordToken("MENOR", "menor")
	query.KeywordToken("CONTIENE", "contiene")
	query.KeywordToken("CATEGORIA", "categoria")
	query.KeywordToken("PRECIO", "precio")
	query.KeywordToken("STOCK", "stock")
	query.KeywordToken("NOMBRE", "nombre")
	
	// Non-keyword tokens
	query.Token("STRING", "\"[^\"]*\"")
	query.Token("NUMBER", "[0-9]+\\.?[0-9]*")
	query.Token("VALUE", "[a-zA-Z][a-zA-Z0-9]*")

	// Define grammar rules - MOST specific rules first, then general ones
	// STRING patterns (most specific)
	query.Rule("query", []string{"BUSCAR", "PRODUCTOS", "DONDE", "NOMBRE", "CONTIENE", "STRING"}, "filteredQueryString")
	query.Rule("query", []string{"LISTAR", "PRODUCTOS", "DONDE", "NOMBRE", "CONTIENE", "STRING"}, "filteredQueryString")
	query.Rule("query", []string{"CONTAR", "PRODUCTOS", "DONDE", "NOMBRE", "CONTIENE", "STRING"}, "filteredQueryString")
	query.Rule("query", []string{"BUSCAR", "PRODUCTOS", "DONDE", "CATEGORIA", "ES", "STRING"}, "filteredQueryString")
	query.Rule("query", []string{"LISTAR", "PRODUCTOS", "DONDE", "CATEGORIA", "ES", "STRING"}, "filteredQueryString")
	query.Rule("query", []string{"CONTAR", "PRODUCTOS", "DONDE", "CATEGORIA", "ES", "STRING"}, "filteredQueryString")
	
	// NUMBER patterns (second most specific)
	query.Rule("query", []string{"BUSCAR", "PRODUCTOS", "DONDE", "PRECIO", "MAYOR", "NUMBER"}, "filteredQueryNumber")
	query.Rule("query", []string{"LISTAR", "PRODUCTOS", "DONDE", "PRECIO", "MAYOR", "NUMBER"}, "filteredQueryNumber")
	query.Rule("query", []string{"CONTAR", "PRODUCTOS", "DONDE", "PRECIO", "MAYOR", "NUMBER"}, "filteredQueryNumber")
	query.Rule("query", []string{"BUSCAR", "PRODUCTOS", "DONDE", "PRECIO", "MENOR", "NUMBER"}, "filteredQueryNumber")
	query.Rule("query", []string{"LISTAR", "PRODUCTOS", "DONDE", "PRECIO", "MENOR", "NUMBER"}, "filteredQueryNumber")
	query.Rule("query", []string{"CONTAR", "PRODUCTOS", "DONDE", "PRECIO", "MENOR", "NUMBER"}, "filteredQueryNumber")
	query.Rule("query", []string{"BUSCAR", "PRODUCTOS", "DONDE", "STOCK", "MAYOR", "NUMBER"}, "filteredQueryNumber")
	query.Rule("query", []string{"LISTAR", "PRODUCTOS", "DONDE", "STOCK", "MAYOR", "NUMBER"}, "filteredQueryNumber")
	query.Rule("query", []string{"CONTAR", "PRODUCTOS", "DONDE", "STOCK", "MAYOR", "NUMBER"}, "filteredQueryNumber")
	query.Rule("query", []string{"BUSCAR", "PRODUCTOS", "DONDE", "STOCK", "MENOR", "NUMBER"}, "filteredQueryNumber")
	query.Rule("query", []string{"LISTAR", "PRODUCTOS", "DONDE", "STOCK", "MENOR", "NUMBER"}, "filteredQueryNumber")
	query.Rule("query", []string{"CONTAR", "PRODUCTOS", "DONDE", "STOCK", "MENOR", "NUMBER"}, "filteredQueryNumber")
	
	// VALUE patterns (less specific)
	query.Rule("query", []string{"BUSCAR", "PRODUCTOS", "DONDE", "CATEGORIA", "ES", "VALUE"}, "filteredQuery")
	query.Rule("query", []string{"LISTAR", "PRODUCTOS", "DONDE", "CATEGORIA", "ES", "VALUE"}, "filteredQuery")
	query.Rule("query", []string{"CONTAR", "PRODUCTOS", "DONDE", "CATEGORIA", "ES", "VALUE"}, "filteredQuery")
	
	// Simple queries (most general - should be last)
	query.Rule("query", []string{"BUSCAR", "PRODUCTOS"}, "simpleQuery") 
	query.Rule("query", []string{"LISTAR", "PRODUCTOS"}, "simpleQuery")
	query.Rule("query", []string{"CONTAR", "PRODUCTOS"}, "simpleQuery")

	// Register Go functions
	query.Set("filterProducts", func(field, operator, value string) []Product {
		var filtered []Product

		for _, p := range products {
			match := false

			switch field {
			case "categoria":
				switch operator {
				case "es":
					match = strings.EqualFold(p.Category, value)
				case "contiene":
					match = strings.Contains(strings.ToLower(p.Category), strings.ToLower(value))
				}
			case "nombre":
				switch operator {
				case "contiene":
					match = strings.Contains(strings.ToLower(p.Name), strings.ToLower(value))
				}
			case "precio":
				// Price comparison handled separately
			case "stock":
				// Stock comparison handled separately
			}

			if match {
				filtered = append(filtered, p)
			}
		}

		return filtered
	})

	query.Set("filterByNumber", func(field, operator string, value float64) []Product {
		var filtered []Product

		for _, p := range products {
			match := false

			switch field {
			case "precio":
				switch operator {
				case "mayor":
					match = p.Price > value
				case "menor":
					match = p.Price < value
				case "es":
					match = p.Price == value
				}
			case "stock":
				switch operator {
				case "mayor":
					match = float64(p.Stock) > value
				case "menor":
					match = float64(p.Stock) < value
				case "es":
					match = float64(p.Stock) == value
				}
			}

			if match {
				filtered = append(filtered, p)
			}
		}

		return filtered
	})

	// Define actions
	query.Action("simpleQuery", func(args []interface{}) (interface{}, error) {
		action := args[0].(string)

		switch action {
		case "listar":
			return products, nil
		case "contar":
			return len(products), nil
		case "buscar":
			return products, nil
		default:
			return products, nil
		}
	})

	query.Action("filteredQuery", func(args []interface{}) (interface{}, error) {
		if len(args) >= 6 {
			action := args[0].(string)
			field := args[3].(string)
			operator := args[4].(string)
			value := args[5].(string)

			filterFn, _ := query.Get("filterProducts")
			filter := filterFn.(func(string, string, string) []Product)
			filtered := filter(field, operator, value)

			switch action {
			case "contar":
				return len(filtered), nil
			default:
				return filtered, nil
			}
		}
		return nil, fmt.Errorf("invalid query")
	})

	query.Action("filteredQueryString", func(args []interface{}) (interface{}, error) {
		if len(args) >= 6 {
			action := args[0].(string)
			field := args[3].(string)
			operator := args[4].(string)
			value := strings.Trim(args[5].(string), "\"")

			filterFn, _ := query.Get("filterProducts")
			filter := filterFn.(func(string, string, string) []Product)
			filtered := filter(field, operator, value)

			switch action {
			case "contar":
				return len(filtered), nil
			default:
				return filtered, nil
			}
		}
		return nil, fmt.Errorf("invalid query")
	})

	query.Action("filteredQueryNumber", func(args []interface{}) (interface{}, error) {
		if len(args) >= 6 {
			action := args[0].(string)
			field := args[3].(string)
			operator := args[4].(string)
			valueStr := args[5].(string)

			var value float64
			fmt.Sscanf(valueStr, "%f", &value)

			filterFn, _ := query.Get("filterByNumber")
			filter := filterFn.(func(string, string, float64) []Product)
			filtered := filter(field, operator, value)

			switch action {
			case "contar":
				return len(filtered), nil
			default:
				return filtered, nil
			}
		}
		return nil, fmt.Errorf("invalid query")
	})

	fmt.Println("Query DSL Demo")
	fmt.Println("==============\n")

	// Test queries
	queries := []string{
		`listar productos`,
		`contar productos`,
		`buscar productos donde categoria es Electronics`,
		`listar productos donde precio mayor 100`,
		`contar productos donde stock menor 10`,
		`buscar productos donde nombre contiene "Desk"`,
		`listar productos donde categoria es Furniture`,
	}

	for _, q := range queries {
		fmt.Printf("Query: %s\n", q)

		result, err := query.Parse(q)
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		output := result.GetOutput()

		switch v := output.(type) {
		case int:
			fmt.Printf("Resultado: %d productos\n", v)
		case []Product:
			fmt.Printf("Resultado: %d productos encontrados\n", len(v))
			for _, p := range v {
				fmt.Printf("  - %s (%s) $%.2f [Stock: %d]\n",
					p.Name, p.Category, p.Price, p.Stock)
			}
		default:
			fmt.Printf("Resultado: %v (tipo: %T)\n", v, v)
		}
		fmt.Println()
	}

	// Demo using context
	fmt.Println("\nDemo con Context:")
	fmt.Println("=================\n")

	// Set a minimum price context
	ctx := map[string]interface{}{
		"minPrice": 50.0,
		"maxPrice": 500.0,
	}

	result, err := query.Use(`listar productos donde precio mayor 50`, ctx)
	if err != nil {
		log.Fatal(err)
	}

	if prods, ok := result.GetOutput().([]Product); ok {
		fmt.Printf("Productos con precio > $50:\n")
		for _, p := range prods {
			if p.Price < 500 { // Apply maxPrice from context
				fmt.Printf("  - %s: $%.2f\n", p.Name, p.Price)
			}
		}
	}
}
