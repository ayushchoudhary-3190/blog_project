package main

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID       uint `gorm:"primaryKey;autoincrement"`
	Name     string
	Email    string `gorm:"unique;not null"`
	Password string
}

type NewBlog struct {
	ID      uint   `gorm:"primaryKey;autoincrement"`
	Title   string `json:"title" form:"title"`
	Genre   string `json:"genre" form:"genre"`
	Content string `json:"content" form:"content"`
	UserID  uint   // just store the user ID, no foreign key constraint
}

func isAuthenticated(c *fiber.Ctx) error {
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve session",
		})
	}

	userID := sess.Get("userID")
	if userID == nil {
		return c.Redirect("/signin")
	}
	return c.Next()
}

// create session variable
var store = session.New()

func main() {

	//databse connection
	dsn := "root:root@tcp(127.0.0.1:3306)/users?charset=utf8mb4&parseTime=True&loc=Local"
	dsn1 := "root:root@tcp(127.0.0.1:3306)/blogs?charset=utf8mb4&parseTime=True&loc=Local"

	//open users database connction
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connet database :", err)
	}

	//automatically create users table
	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal("failed to migrate users database:", err)
	}

	//open blogs database connection
	db1, err := gorm.Open(mysql.Open(dsn1), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to conenect blogs database :", err)
	}

	//automatically create blogs table
	err = db1.AutoMigrate(&NewBlog{})
	if err != nil {
		log.Fatal("failed to migrate blogs database :", err)
	}

	type SigninInput struct {
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
	}

	type SignupInput struct {
		Name     string `json:"name" form:"name"`
		Email    string `json:"email" form:"email"`
		Password string `json:"password"  form:"password"`
	}

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Static("/static", "./static")

	app.Get("/", func(c *fiber.Ctx) error {
		//retrieve the session
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to retrieve session",
			})
		}

		//check if user is logged in
		userID := sess.Get("userID")
		isLoggedin := userID != nil

		//return homepage
		return c.Render("index", fiber.Map{
			"title":      "welcome to my blog",
			"isLoggedin": isLoggedin,
		})
	})

	// render sign in page

	app.Get("/signin", func(c *fiber.Ctx) error {
		return c.Render("sign_in", fiber.Map{
			"Title": "sign in page",
		})
	})

	//SIGNIN SESSION
	app.Post("/signin", func(c *fiber.Ctx) error {
		// Parse form input
		input := new(SigninInput)
		if err := c.BodyParser(input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid input",
			})
		}

		// Query the database for the user
		var user User
		result := db.Where("email = ?", input.Email).First(&user)
		if result.Error != nil {
			// If user does not exist, redirect to the sign-up page
			if result.Error == gorm.ErrRecordNotFound {
				return c.Redirect("/signup")
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to query database",
			})
		}

		// Verify the password (assuming passwords are stored as plain text for simplicity)
		// In production, always hash and compare passwords using a library like bcrypt
		if user.Password != input.Password {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid email or password",
			})
		}

		//Store the session
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to create a session",
			})
		}

		// Store user ID in session
		sess.Set("userID", user.ID)
		if err := sess.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to save session",
			})
		}

		// If successful, return to home page
		return c.Redirect("/")

	})

	//SIGNUP STUFF-------------------------------------
	// render and setup route for sign up page

	app.Get("/signup", func(c *fiber.Ctx) error {
		return c.Render("sign_up", fiber.Map{
			"Title": "sign up page",
		})
	})

	app.Post("/signup", func(c *fiber.Ctx) error {
		input := new(SignupInput)
		if err := c.BodyParser(input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid input",
			})
		}

		//validate if all fields are filled
		if input.Name == "" || input.Email == "" || input.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "all fields are required",
			})
		}

		//check if user exists
		var existingUser User
		result := db.Where("email = ?", input.Email).First(&existingUser)
		if result.Error == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "email already in use",
			})
		}

		//create a new user
		newUser := User{
			Name:     input.Name,
			Email:    input.Email,
			Password: input.Password,
		}

		//save the user
		if err := db.Create(&newUser).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to create a user",
			})
		}

		//return success message
		return c.JSON(fiber.Map{
			"message": "sign-up successful",
			"user":    newUser.Name,
		})
	})

	//logout functionality
	app.Get("/logout", func(c *fiber.Ctx) error {
		//retrieve the session
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "fialed to retirve session",
			})
		}

		//destroy session
		if err := sess.Destroy(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to destroy session",
			})
		}

		return c.Redirect("/")
	})

	//NEW BLOG CREATION PAGE-------------------------------

	//render the page
	app.Get("/create", isAuthenticated, func(c *fiber.Ctx) error {
		return c.Render("new_blog", fiber.Map{
			"Title": "Blog creation page",
		})
	})

	app.Post("/create", isAuthenticated, func(c *fiber.Ctx) error {
		type BlogInput struct {
			Title   string `form:"title"`
			Genre   string `form:"genre"`
			Content string `form:"content"`
		}

		// Retrieve the session
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "failed to retrieve session",
			})
		}

		userID := sess.Get("userID")

		// Convert userID to an integer
		var userIDUint uint

		switch v := userID.(type) {
		case int:
			userIDUint = uint(v)
		case int64:
			userIDUint = uint(v)
		case float64:
			userIDUint = uint(v)
		case uint:
			userIDUint = v
		case string:
			parsed, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to parse user ID from string",
				})
			}
			userIDUint = uint(parsed)
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to retrieve user ID",
			})
		}
		if userIDUint == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid user session",
			})
		}
		input := new(BlogInput)
		if err := c.BodyParser(input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid input",
			})
		}

		// Debug: Print parsed input
		log.Printf("Parsed blog input: Title='%s', Genre='%s', Content='%s'", input.Title, input.Genre, input.Content)

		if input.Title == "" || input.Genre == "" || input.Content == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "All fields are required",
			})
		}

		// Save the blog to the database
		newBlog := NewBlog{
			Title:   input.Title,
			Genre:   input.Genre,
			Content: input.Content,
			UserID:  userIDUint,
		}
		if err := db1.Create(&newBlog).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save blog",
			})
		}

		return c.Redirect("/")
	})

	//render all blogs created by user under myblogs button
	// Render all blogs created by the logged-in user
	app.Get("/blogs", isAuthenticated, func(c *fiber.Ctx) error {
		// Retrieve session
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to retrieve session",
			})
		}

		// ...existing code...
		userID := sess.Get("userID")
		var userIDUint uint

		switch v := userID.(type) {
		case int:
			userIDUint = uint(v)
		case int64:
			userIDUint = uint(v)
		case float64:
			userIDUint = uint(v)
		case uint:
			userIDUint = v
		case string:
			// Try to parse string to uint
			parsed, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to parse user ID from string",
				})
			}
			userIDUint = uint(parsed)
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to retrieve user ID",
			})
		}

		// Fetch blogs for this user from the blogs database
		var blogs []NewBlog
		if err := db1.Where("user_id = ?", userIDUint).Find(&blogs).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to fetch blogs",
			})
		}

		return c.Render("my_blogs", fiber.Map{
			"Blogs": blogs,
		})
	})

	app.Listen(":8080")
}
