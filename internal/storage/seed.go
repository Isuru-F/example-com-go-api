package storage

import "ecom-book-store-sample-api/internal/models"

// Seed seeds users and products for demo
func Seed(store *MemoryStore) {
	// Users
	store.CreateUser(&models.User{Email: "john@email.com", Name: "John Doe"})
	store.CreateUser(&models.User{Email: "jane@email.com", Name: "Jane Smith"})
	store.CreateUser(&models.User{Email: "admin@email.com", Name: "Admin User"})

	// Products (books)
	products := []models.Product{
		{Title: "The Pragmatic Programmer", Author: "Andrew Hunt", Description: "Journey to Mastery", Price: 45.00, Stock: 50},
		{Title: "Clean Code", Author: "Robert C. Martin", Description: "A Handbook of Agile Software Craftsmanship", Price: 39.99, Stock: 60},
		{Title: "Design Patterns", Author: "Erich Gamma", Description: "Elements of Reusable OO Software", Price: 59.99, Stock: 40},
		{Title: "Introduction to Algorithms", Author: "CLRS", Description: "Comprehensive algorithms text", Price: 89.50, Stock: 35},
		{Title: "Refactoring", Author: "Martin Fowler", Description: "Improving the Design of Existing Code", Price: 49.99, Stock: 45},
		{Title: "You Don't Know JS Yet", Author: "Kyle Simpson", Description: "Deep dive into JavaScript", Price: 29.99, Stock: 70},
		{Title: "Operating Systems: Three Easy Pieces", Author: "Remzi Arpaci-Dusseau", Description: "OS concepts", Price: 25.00, Stock: 80},
		{Title: "Deep Learning", Author: "Goodfellow, Bengio, Courville", Description: "Foundational deep learning book", Price: 120.00, Stock: 20},
		{Title: "Domain-Driven Design", Author: "Eric Evans", Description: "Tackling Complexity in the Heart of Software", Price: 74.99, Stock: 30},
		{Title: "Computer Networks", Author: "Andrew S. Tanenbaum", Description: "Networking principles", Price: 65.00, Stock: 55},
	}
	for i := range products { store.CreateProduct(&products[i]) }
}
