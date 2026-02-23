// Package seeds registers all database seeds via init().
//
// Import this package with a blank identifier in your main.go:
//
//	import _ "myapp/seeds"
package seeds

import "github.com/jianhuahe/geeder"

func init() {
	geeder.Register("001_create_products_table", `
		CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			price REAL NOT NULL
		)
	`)

	geeder.Register("002_seed_products", `
		INSERT INTO products (name, price) VALUES ('Widget', 9.99);
		INSERT INTO products (name, price) VALUES ('Gadget', 24.99);
		INSERT INTO products (name, price) VALUES ('Doohickey', 4.99);
	`)

	geeder.Register("003_create_categories_table", `
		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		)
	`)

	geeder.Register("004_seed_categories", `
		INSERT INTO categories (name) VALUES ('Electronics');
		INSERT INTO categories (name) VALUES ('Tools');
	`)
}
