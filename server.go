package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Member struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Record struct {
	ID     string   `json:"id"`
	Amount int      `json:"amount"`
	Payer  Member   `json:"payer"`
	Owes   []Member `json:"owes"`
	Note   string   `json:"note"`
}

// CreateRecord ...
func CreateRecord(c echo.Context) error {
	r := new(Record)
	if err := c.Bind(r); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, r)
}

// GetRecords ...
func GetRecords(c echo.Context) error {
	return c.JSON(http.StatusOK, []Record{})
}

// DeleteRecord ...
func DeleteRecord(c echo.Context) error {
	return c.JSON(http.StatusOK, Record{})
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/records", GetRecords)
	e.POST("/records", CreateRecord)
	e.DELETE("/records/:id", DeleteRecord)

	// Start server
	e.Logger.Fatal(e.Start(":2019"))
}
