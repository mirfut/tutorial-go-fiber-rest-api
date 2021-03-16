package controllers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/koddr/tutorial-go-fiber-rest-api/app/models"
	"github.com/koddr/tutorial-go-fiber-rest-api/app/validators"
	"github.com/koddr/tutorial-go-fiber-rest-api/pkg/utils"
	"github.com/koddr/tutorial-go-fiber-rest-api/platform/database"
)

// GetBooks func gets all exists books.
// @Description Get all exists books.
// @Summary get all exists books
// @Tags Public
// @Accept json
// @Produce json
// @Success 200 {array} models.Book
// @Router /api/v1/books [get]
func GetBooks(c *fiber.Ctx) error {
	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		// Return status 500 and database connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Get all books.
	books, err := db.GetBooks()
	if err != nil {
		// Return, if books not found.
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": true,
			"msg":   "books were not found",
			"count": 0,
			"books": nil,
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"msg":   nil,
		"count": len(books),
		"books": books,
	})
}

// GetBook func gets book by given ID or 404 error.
// @Description Get book by given ID.
// @Summary get book by given ID
// @Tags Public
// @Accept json
// @Produce json
// @Param id path string true "Book ID"
// @Success 200 {object} models.Book
// @Router /api/v1/book/{id} [get]
func GetBook(c *fiber.Ctx) error {
	// Catch book ID from URL.
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Create database connection.
	db, err := database.OpenDBConnection()
	if err != nil {
		// Return status 500 and database connection error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Get book by ID.
	book, err := db.GetBook(id)
	if err != nil {
		// Return, if book not found.
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": true,
			"msg":   "book with the given ID is not found",
			"book":  nil,
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"msg":   nil,
		"book":  book,
	})
}

// CreateBook func for creates a new book.
// @Description Create a new book.
// @Summary create a new book
// @Tags Private
// @Accept json
// @Produce json
// @Param title body string true "Title"
// @Param author body string true "Author"
// @Success 201 {object} models.Book
// @Router /api/v1/book [post]
func CreateBook(c *fiber.Ctx) error {
	// Get now time.
	now := time.Now().Unix()

	// Get claims from JWT.
	claims, err := utils.ExtractTokenMetadata(c)
	if err != nil {
		// Return status 500 and JWT parse error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Set expiration time from JWT data of current book.
	expires := claims.Expires

	// Set credential `book:create` from JWT data of current book.
	credential := claims.Credentials["book:create"]

	// Create a new book struct.
	book := &models.Book{}

	// Checking received data from JSON body.
	if err := c.BodyParser(book); err != nil {
		// Return, if JSON data is not correct.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Only book with `book:create` credential can create a new book.
	if credential && now < expires {
		// Create a new validator for a book model.
		validate := validators.BookValidator()

		// Validate book fields.
		if err := validate.Struct(book); err != nil {
			// Return, if some fields are not valid.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   utils.ValidatorErrors(err),
			})
		}

		// Create database connection.
		db, err := database.OpenDBConnection()
		if err != nil {
			// Return status 500 and database connection error.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		// Set initialized default data for book:
		book.ID = uuid.New()
		book.CreatedAt = time.Now()
		book.UpdatedAt = time.Time{}
		book.BookStatus = 1 // 0 == draft, 1 == active
		book.BookAttrs = models.BookAttrs{}

		// Create a new book with validated data.
		if err := db.CreateBook(book); err != nil {
			// Return status 500 and create book process error.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}
	} else {
		// Return status 403 and permission denied error.
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": true,
			"msg":   "permission denied, check credentials or expiration time of your token",
			"book":  nil,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error": false,
		"msg":   nil,
		"book":  book,
	})
}

// UpdateBook func for updates book by given ID.
// @Description Update book.
// @Summary update book
// @Tags Private
// @Accept json
// @Produce json
// @Param id body string true "Book ID"
// @Success 202 {object} models.Book
// @Router /api/v1/book [patch]
func UpdateBook(c *fiber.Ctx) error {
	// Get now time.
	now := time.Now().Unix()

	// Get claims from JWT.
	claims, err := utils.ExtractTokenMetadata(c)
	if err != nil {
		// Return status 500 and JWT parse error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Set expiration time from JWT data of current book.
	expires := claims.Expires

	// Set credential `book:update` from JWT data of current book.
	credential := claims.Credentials["book:update"]

	// Create a new book struct.
	book := &models.Book{}

	// Checking received data from JSON body.
	if err := c.BodyParser(book); err != nil {
		// Return, if JSON data is not correct.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Only book with `book:update` credential can update book profile.
	if credential && now < expires {
		// Create a new validator for a book model.
		validate := validators.BookValidator()

		// Validate book fields.
		if err := validate.Struct(book); err != nil {
			// Return, if some fields are not valid.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   utils.ValidatorErrors(err),
			})
		}

		// Create database connection.
		db, err := database.OpenDBConnection()
		if err != nil {
			// Return status 500 and database connection error.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		// Checking, if book with given ID is exists.
		if _, err := db.GetBook(book.ID); err != nil {
			// Return status 404 and book not found error.
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": true,
				"msg":   "book not found",
			})
		}

		// Set book data to update:
		book.UpdatedAt = time.Now()

		// Update book.
		if err := db.UpdateBook(book); err != nil {
			// Return status 500 and book update error.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}
	} else {
		// Return status 403 and permission denied error.
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": true,
			"msg":   "permission denied, check credentials or expiration time of your token",
			"book":  nil,
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"error": false,
		"msg":   nil,
		"book":  book,
	})
}

// DeleteBook func for deletes book by given ID.
// @Description Delete book by given ID.
// @Summary delete book by given ID
// @Tags Private
// @Accept json
// @Produce json
// @Param id body string true "Book ID"
// @Success 200 {string} string "ok"
// @Router /api/v1/book [delete]
func DeleteBook(c *fiber.Ctx) error {
	// Get now time.
	now := time.Now().Unix()

	// Get claims from JWT.
	claims, err := utils.ExtractTokenMetadata(c)
	if err != nil {
		// Return status 500 and JWT parse error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Set expiration time from JWT data of current book.
	expires := claims.Expires

	// Set credential `book:delete` from JWT data of current book.
	credential := claims.Credentials["book:delete"]

	// Create new Book struct
	book := &models.Book{}

	// Check, if received JSON data is valid.
	if err := c.BodyParser(book); err != nil {
		// Return status 500 and JSON parse error.
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Only book with `book:delete` credential can delete book profile.
	if credential && now < expires {
		// Create database connection.
		db, err := database.OpenDBConnection()
		if err != nil {
			// Return status 500 and database connection error.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		// Checking, if book with given ID is exists.
		if _, err := db.GetBook(book.ID); err != nil {
			// Return status 404 and book not found error.
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": true,
				"msg":   "book not found",
			})
		}

		// Delete book by given ID.
		if err := db.DeleteBook(book.ID); err != nil {
			// Return status 500 and delete book process error.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}
	} else {
		// Return status 403 and permission denied error.
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": true,
			"msg":   "permission denied, check credentials or expiration time of your token",
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"msg":   nil,
	})
}
