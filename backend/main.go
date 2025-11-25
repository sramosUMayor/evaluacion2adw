package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Category struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"unique;not null"`
}

type Product struct {
	ID           uint     `json:"id" gorm:"primaryKey"`
	Name         string   `json:"name" gorm:"not null"`
	Description  string   `json:"description"`
	Price        int      `json:"price"`
	ImageURL     string   `json:"image_url"`
	Watering     string   `json:"watering"`
	Light        string   `json:"light"`
	CategoryID   uint     `json:"category_id"`
	Category     Category `json:"-" gorm:"foreignKey:CategoryID"`
	CategoryName string   `json:"category" gorm:"-"`
	Stock        bool     `json:"stock"`
}

type OrderItem struct {
	ID        uint `json:"-" gorm:"primaryKey"`
	OrderID   uint `json:"-"`
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type Order struct {
	ID            uint        `json:"id" gorm:"primaryKey"`
	CustomerName  string      `json:"customer_name"`
	CustomerEmail string      `json:"customer_email"`
	Address       string      `json:"address"`
	Items         []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
	Total         int         `json:"total"`
	CreatedAt     int64       `json:"created_at" gorm:"autoCreateTime"`
}

var db *gorm.DB

func main() {
	initFlag := flag.Bool("init", false, "Initialize database with seed data")
	flag.Parse()

	var err error
	db, err = gorm.Open(sqlite.Open("floraverde.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	err = db.AutoMigrate(&Category{}, &Product{}, &Order{}, &OrderItem{})
	if err != nil {
		log.Fatal("failed to migrate database:", err)
	}

	if *initFlag {
		seedData()
	} else {
		var count int64
		db.Model(&Product{}).Count(&count)
		if count == 0 {
			log.Println("Base de datos vacía, sembrando datos...")
			seedData()
		}
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	r.Static("/css", "../frontend/css")
	r.Static("/js", "../frontend/js")
	r.Static("/components", "../frontend/components")
	r.StaticFile("/", "../frontend/index.html")
	r.StaticFile("/index.html", "../frontend/index.html")
	r.StaticFile("/catalogo.html", "../frontend/catalogo.html")
	r.StaticFile("/carrito.html", "../frontend/carrito.html")
	r.StaticFile("/checkout.html", "../frontend/checkout.html")
	r.StaticFile("/admin.html", "../frontend/admin.html")
	r.StaticFile("/faq.html", "../frontend/faq.html")
	r.StaticFile("/politicas.html", "../frontend/politicas.html")
	r.StaticFile("/thank-you.html", "../frontend/thank-you.html")

	api := r.Group("/api")
	{
		api.GET("/products", getProducts)
		api.POST("/products", createProduct)
		api.PUT("/products", updateProduct)
		api.DELETE("/products", deleteProduct)
		api.GET("/categories", getCategories)
		api.POST("/orders", createOrder)
	}

	port := "8080"
	fmt.Printf("Servidor corriendo en http://localhost:%s\n", port)
	r.Run(":" + port)
}

func seedData() {
	db.Exec("DELETE FROM order_items")
	db.Exec("DELETE FROM orders")
	db.Exec("DELETE FROM products")
	db.Exec("DELETE FROM categories")

	categories := []Category{
		{Name: "Interior"},
		{Name: "Exterior"},
		{Name: "Suculentas"},
		{Name: "Herramientas"},
		{Name: "Cactus"},
	}

	for i := range categories {
		db.Create(&categories[i])
	}

	getCatID := func(name string) uint {
		for _, c := range categories {
			if c.Name == name {
				return c.ID
			}
		}
		return 0
	}

	products := []Product{
		{Name: "Monstera Deliciosa", Description: "Planta tropical de interior con hojas grandes y vistosas. Ideal para espacios amplios con luz indirecta.", Price: 15990, ImageURL: "https://d17jkdlzll9byv.cloudfront.net/wp-content/uploads/2022/07/monstera-deliciosa-003.jpg", Watering: "Riego moderado", Light: "Luz indirecta", CategoryID: getCatID("Interior"), Stock: true},
		{Name: "Pothos Dorado", Description: "Planta colgante de fácil cuidado. Perfecta para principiantes y espacios con poca luz natural.", Price: 12990, ImageURL: "https://res.cloudinary.com/fronda/image/upload/f_auto,q_auto,c_fill,g_center,w_528,h_704/productos/fol/10012/10012157_1.jpg?02-01-2024", Watering: "Poco riego", Light: "Luz baja-media", CategoryID: getCatID("Interior"), Stock: true},
		{Name: "Helecho Boston", Description: "Planta purificadora del aire con follaje exuberante. Ideal para baños y cocinas con humedad.", Price: 9990, ImageURL: "https://cdn.be.green/small/63d3e7c8713d7602115927.jpg", Watering: "Riego frecuente", Light: "Luz indirecta", CategoryID: getCatID("Interior"), Stock: true},
		{Name: "Suculenta Mix", Description: "Set de 3 suculentas variadas en macetas decorativas. Requieren poco riego y mantenimiento mínimo.", Price: 8990, ImageURL: "https://cdnx.jumpseller.com/www-feelflowers-cl/image/33635783/thumb/1079/1079?1680287114", Watering: "Poco riego", Light: "Luz directa", CategoryID: getCatID("Suculentas"), Stock: true},
		{Name: "Cactus San Pedro", Description: "Cactus columnar de rápido crecimiento. Resistente y de bajo mantenimiento, ideal para exterior.", Price: 6990, ImageURL: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRClTLm2tMnPTTNoi6SjGDUxFXT9-QxUNmjzg&s", Watering: "Muy poco riego", Light: "Sol directo", CategoryID: getCatID("Cactus"), Stock: true},
		{Name: "Ficus Lyrata", Description: "También conocida como Higuera de hoja de violín. Planta de interior elegante con hojas grandes.", Price: 18990, ImageURL: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTWQpoUzihEF7JX5YBKxr_ERZ7sfDoY7Ymvbg&s", Watering: "Riego moderado", Light: "Luz brillante", CategoryID: getCatID("Interior"), Stock: true},
		{Name: "Sansevieria", Description: "Una de las plantas más resistentes. Perfecta para oficinas y espacios con poca luz.", Price: 11990, ImageURL: "https://www.jardinerosenlima.com/wp-content/uploads/2023/03/Beneficios-y-cuidados-lengua-de-suegra.png", Watering: "Muy poco riego", Light: "Luz baja-alta", CategoryID: getCatID("Interior"), Stock: true},
		{Name: "Aloe Vera", Description: "Planta medicinal con múltiples beneficios. Fácil de cuidar y de propiedades curativas.", Price: 7990, ImageURL: "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSQrao-I2Go9SnPZupwF0vFa0o1tVbt4GiN6pdMnmqrVPpenCATOJdo-iRY3DhWtJuSbLc&usqp=CAU", Watering: "Poco riego", Light: "Luz directa", CategoryID: getCatID("Interior"), Stock: true},
		{Name: "Lavanda", Description: "Planta aromática de flores violetas. Ideal para jardines y balcones soleados.", Price: 9990, ImageURL: "https://cdn.shopify.com/s/files/1/0272/1392/2339/files/Lavanda-dentata_22o__cocoantracita_comprar-plantas-online_plantas-de-interior.jpg?v=1689089438", Watering: "Poco riego", Light: "Sol directo", CategoryID: getCatID("Exterior"), Stock: true},
	}

	for _, p := range products {
		db.Create(&p)
	}
	fmt.Println("¡Base de datos sembrada exitosamente!")
}

func getProducts(c *gin.Context) {
	categoryName := c.Query("category")
	var products []Product
	var err error

	tx := db.Preload("Category")

	if categoryName != "" {
		tx = tx.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.name = ?", categoryName)
	}

	if err = tx.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i := range products {
		products[i].CategoryName = products[i].Category.Name
	}

	c.JSON(http.StatusOK, products)
}

func createProduct(c *gin.Context) {
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	db.Preload("Category").First(&product, product.ID)
	product.CategoryName = product.Category.Name

	c.JSON(http.StatusCreated, product)
}

func updateProduct(c *gin.Context) {
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idStr := c.Query("id")
	if idStr != "" {
		id, _ := strconv.Atoi(idStr)
		product.ID = uint(id)
	}

	if product.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de producto requerido"})
		return
	}

	if err := db.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	db.Preload("Category").First(&product, product.ID)
	product.CategoryName = product.Category.Name

	c.JSON(http.StatusOK, product)
}

func deleteProduct(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de producto requerido"})
		return
	}

	if err := db.Delete(&Product{}, idStr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func getCategories(c *gin.Context) {
	var categories []Category
	if err := db.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func createOrder(c *gin.Context) {
	var order Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": order.ID, "status": "created"})
}
