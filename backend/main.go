package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	ImageURL    string `json:"image_url"`
	Watering    string `json:"watering"`
	Light       string `json:"light"`
	CategoryID  int    `json:"category_id"`
	Category    string `json:"category"`
	Stock       bool   `json:"stock"`
}

type OrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type Order struct {
	ID            int         `json:"id"`
	CustomerName  string      `json:"customer_name"`
	CustomerEmail string      `json:"customer_email"`
	Address       string      `json:"address"`
	Items         []OrderItem `json:"items"`
	Total         int         `json:"total"`
}

var db *sql.DB

func main() {

	initFlag := flag.Bool("init", false, "Initialize database with seed data")
	flag.Parse()

	var err error

	db, err = sql.Open("sqlite", "./floraverde.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTables()

	if *initFlag {
		seedData()
	} else {

		var count int
		row := db.QueryRow("SELECT COUNT(*) FROM products")
		if err := row.Scan(&count); err == nil && count == 0 {
			log.Println("Base de datos vacía, sembrando datos...")
			seedData()
		}
	}

	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	http.HandleFunc("/api/products", enableCORS(productsHandler))
	http.HandleFunc("/api/categories", enableCORS(getCategories))
	http.HandleFunc("/api/orders", enableCORS(createOrder))

	port := "8080"
	fmt.Printf("Servidor corriendo en http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func createTables() {
	query := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		price INTEGER,
		image_url TEXT,
		watering TEXT,
		light TEXT,
		category_id INTEGER,
		stock BOOLEAN,
		FOREIGN KEY(category_id) REFERENCES categories(id)
	);

	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		customer_name TEXT,
		customer_email TEXT,
		address TEXT,
		total INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS order_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id INTEGER,
		product_id INTEGER,
		quantity INTEGER,
		FOREIGN KEY(order_id) REFERENCES orders(id),
		FOREIGN KEY(product_id) REFERENCES products(id)
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Error al crear tablas:", err)
	}
}

func seedData() {
	categories := []string{"Interior", "Exterior", "Suculentas", "Herramientas", "Cactus"}
	categoryIDsByName := make(map[string]int)

	categoryStatement, err := db.Prepare("INSERT INTO categories(name) VALUES(?)")
	if err != nil {
		log.Fatal(err)
	}
	defer categoryStatement.Close()

	for _, categoryName := range categories {
		result, err := categoryStatement.Exec(categoryName)
		if err != nil {
			log.Fatal(err)
		}
		id, _ := result.LastInsertId()
		categoryIDsByName[categoryName] = int(id)
	}

	products := []struct {
		Name         string
		Description  string
		Price        int
		ImageURL     string
		Watering     string
		Light        string
		CategoryName string
		Stock        bool
	}{
		{Name: "Monstera Deliciosa", Description: "Planta tropical de interior con hojas grandes y vistosas. Ideal para espacios amplios con luz indirecta.", Price: 15990, ImageURL: "https://d17jkdlzll9byv.cloudfront.net/wp-content/uploads/2022/07/monstera-deliciosa-003.jpg", Watering: "Riego moderado", Light: "Luz indirecta", CategoryName: "Interior", Stock: true},
		{Name: "Pothos Dorado", Description: "Planta colgante de fácil cuidado. Perfecta para principiantes y espacios con poca luz natural.", Price: 12990, ImageURL: "https://res.cloudinary.com/fronda/image/upload/f_auto,q_auto,c_fill,g_center,w_528,h_704/productos/fol/10012/10012157_1.jpg?02-01-2024", Watering: "Poco riego", Light: "Luz baja-media", CategoryName: "Interior", Stock: true},
		{Name: "Helecho Boston", Description: "Planta purificadora del aire con follaje exuberante. Ideal para baños y cocinas con humedad.", Price: 9990, ImageURL: "https://cdn.be.green/small/63d3e7c8713d7602115927.jpg", Watering: "Riego frecuente", Light: "Luz indirecta", CategoryName: "Interior", Stock: true},
		{Name: "Suculenta Mix", Description: "Set de 3 suculentas variadas en macetas decorativas. Requieren poco riego y mantenimiento mínimo.", Price: 8990, ImageURL: "https://cdnx.jumpseller.com/www-feelflowers-cl/image/33635783/thumb/1079/1079?1680287114", Watering: "Poco riego", Light: "Luz directa", CategoryName: "Suculentas", Stock: true},
		{Name: "Cactus San Pedro", Description: "Cactus columnar de rápido crecimiento. Resistente y de bajo mantenimiento, ideal para exterior.", Price: 6990, ImageURL: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRClTLm2tMnPTTNoi6SjGDUxFXT9-QxUNmjzg&s", Watering: "Muy poco riego", Light: "Sol directo", CategoryName: "Cactus", Stock: true},
		{Name: "Ficus Lyrata", Description: "También conocida como Higuera de hoja de violín. Planta de interior elegante con hojas grandes.", Price: 18990, ImageURL: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTWQpoUzihEF7JX5YBKxr_ERZ7sfDoY7Ymvbg&s", Watering: "Riego moderado", Light: "Luz brillante", CategoryName: "Interior", Stock: true},
		{Name: "Sansevieria", Description: "Una de las plantas más resistentes. Perfecta para oficinas y espacios con poca luz.", Price: 11990, ImageURL: "https://www.jardinerosenlima.com/wp-content/uploads/2023/03/Beneficios-y-cuidados-lengua-de-suegra.png", Watering: "Muy poco riego", Light: "Luz baja-alta", CategoryName: "Interior", Stock: true},
		{Name: "Aloe Vera", Description: "Planta medicinal con múltiples beneficios. Fácil de cuidar y de propiedades curativas.", Price: 7990, ImageURL: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSQrao-I2Go9SnPZupwF0vFa0o1tVbt4GiN6pdMnmqrVPpenCATOJdo-iRY3DhWtJuSbLc&usqp=CAU", Watering: "Poco riego", Light: "Luz directa", CategoryName: "Interior", Stock: true},
		{Name: "Lavanda", Description: "Planta aromática de flores violetas. Ideal para jardines y balcones soleados.", Price: 9990, ImageURL: "https://cdn.shopify.com/s/files/1/0272/1392/2339/files/Lavanda-dentata_22o__cocoantracita_comprar-plantas-online_plantas-de-interior.jpg?v=1689089438", Watering: "Poco riego", Light: "Sol directo", CategoryName: "Exterior", Stock: true},
	}

	productStatement, err := db.Prepare("INSERT INTO products(name, description, price, image_url, watering, light, category_id, stock) VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer productStatement.Close()

	_, _ = db.Exec("DELETE FROM products")

	for _, product := range products {
		categoryID := categoryIDsByName[product.CategoryName]
		_, err := productStatement.Exec(product.Name, product.Description, product.Price, product.ImageURL, product.Watering, product.Light, categoryID, product.Stock)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("¡Base de datos sembrada exitosamente!")
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func getProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	category := r.URL.Query().Get("category")
	var rows *sql.Rows
	var err error

	query := `
		SELECT p.id, p.name, p.description, p.price, p.image_url, p.watering, p.light, p.category_id, c.name as category, p.stock 
		FROM products p 
		JOIN categories c ON p.category_id = c.id
	`

	if category != "" {
		query += " WHERE c.name = ?"
		rows, err = db.Query(query, category)
	} else {
		rows, err = db.Query(query)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.ImageURL, &product.Watering, &product.Light, &product.CategoryID, &product.Category, &product.Stock); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		products = append(products, product)
	}

	if products == nil {
		products = []Product{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	transaction, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := transaction.Exec("INSERT INTO orders(customer_name, customer_email, address, total) VALUES(?, ?, ?, ?)",
		order.CustomerName, order.CustomerEmail, order.Address, order.Total)
	if err != nil {
		transaction.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	orderID, _ := result.LastInsertId()

	orderItemStatement, err := transaction.Prepare("INSERT INTO order_items(order_id, product_id, quantity) VALUES(?, ?, ?)")
	if err != nil {
		transaction.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer orderItemStatement.Close()

	for _, item := range order.Items {
		_, err = orderItemStatement.Exec(orderID, item.ProductID, item.Quantity)
		if err != nil {
			transaction.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := transaction.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": orderID, "status": "created"})
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getProducts(w, r)
	case "POST":
		createProduct(w, r)
	case "PUT":
		updateProduct(w, r)
	case "DELETE":
		deleteProduct(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	statement, err := db.Prepare("INSERT INTO products(name, description, price, image_url, watering, light, category_id, stock) VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer statement.Close()

	result, err := statement.Exec(product.Name, product.Description, product.Price, product.ImageURL, product.Watering, product.Light, product.CategoryID, product.Stock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	product.ID = int(id)

	var categoryName string
	err = db.QueryRow("SELECT name FROM categories WHERE id = ?", product.CategoryID).Scan(&categoryName)
	if err == nil {
		product.Category = categoryName
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func updateProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr != "" {
		id, _ := strconv.Atoi(idStr)
		product.ID = id
	}

	if product.ID == 0 {
		http.Error(w, "ID de producto requerido", http.StatusBadRequest)
		return
	}

	statement, err := db.Prepare("UPDATE products SET name=?, description=?, price=?, image_url=?, watering=?, light=?, category_id=?, stock=? WHERE id=?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer statement.Close()

	_, err = statement.Exec(product.Name, product.Description, product.Price, product.ImageURL, product.Watering, product.Light, product.CategoryID, product.Stock, product.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var categoryName string
	err = db.QueryRow("SELECT name FROM categories WHERE id = ?", product.CategoryID).Scan(&categoryName)
	if err == nil {
		product.Category = categoryName
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID de producto requerido", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID de producto inválido", http.StatusBadRequest)
		return
	}

	statement, err := db.Prepare("DELETE FROM products WHERE id=?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer statement.Close()

	_, err = statement.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func getCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		categories = append(categories, category)
	}

	if categories == nil {
		categories = []Category{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}
