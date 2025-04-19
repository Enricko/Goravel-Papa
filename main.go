package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	excelize "github.com/xuri/excelize/v2"
)

type Product struct {
	Name  string
	Price float32
}

type ExlProduct struct {
	Code        string
	NameProduct string
	HNA         string
	PPN         string
}

type ExlData struct {
	Branch    string
	CustId    string
	CustName  string
	Alamat    string
	Kota      string
	SalesName string
	Channel   string
	Avg2023   string
	Avg2024   string
	Avg2025   string
	Max       string
}

type PageData struct {
	Customers []ExlData
}

func (data PageData) FindByID(id string) *ExlData {
	for _, customer := range data.Customers {
		if customer.CustId == id {
			return &customer
		}
	}
	return nil
}

func main() {
	// Create a new Gin router
	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Define a route handler
	r.GET("/", indexHandler)

	// Create server configurations
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Try HTTPS first
	log.Printf("Starting server on :8080")
	// r.Run(":8080")
	if err := server.ListenAndServeTLS("cert.pem", "key.pem"); err != nil {
		log.Printf("Failed to start HTTPS server: %v", err)
		log.Printf("Falling back to HTTP")
		// If HTTPS fails, fall back to HTTP
		if err := r.Run(":8080"); err != nil {
			log.Fatal("Error starting HTTP server:", err)
		}
	}
}

func indexHandler(c *gin.Context) {
	// Read customer data from local excel file
	data, err := openLocalCustFile("excel/jogja.xlsx")
	if err != nil {
		log.Printf("Error reading customer data: %v", err)
		c.String(http.StatusInternalServerError, "Error reading customer data")
		return
	}

	// Read product data from local excel file
	dataProduct, err := openLocalProductFile("excel/harga_jogja.xlsx")
	if err != nil {
		log.Printf("Error reading product data: %v", err)
		c.String(http.StatusInternalServerError, "Error reading product data")
		return
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")

	// Execute the template with the data
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		c.String(http.StatusInternalServerError, "Error parsing template")
		return
	}

	// Render the template to the response
	if err := tmpl.Execute(c.Writer, map[string]interface{}{
		"customer": data,
		"product":  dataProduct,
		"time":     formattedTime,
	}); err != nil {
		log.Printf("Error executing template: %v", err)
		c.String(http.StatusInternalServerError, "Error executing template")
		return
	}
}

func openLocalCustFile(filePath string) ([]ExlData, error) {
	var exlData []ExlData

	// Make sure the path is absolute
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute file path: %v", err)
	}

	// Open the Excel file with Excelize
	exlz, err := excelize.OpenFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("error opening excel file: %v", err)
	}
	defer exlz.Close()

	// Try to get rows from "Data Base" sheet
	rows, err := exlz.GetRows("Data Base")
	if err != nil {
		// If "Data Base" sheet doesn't exist, try alternate names
		alternateNames := []string{"Database", "Sheet1", "Data"}
		var rowsFound bool

		for _, sheetName := range alternateNames {
			rows, err = exlz.GetRows(sheetName)
			if err == nil {
				rowsFound = true
				log.Printf("Found customer data in sheet: %s", sheetName)
				break
			}
		}

		if !rowsFound {
			return nil, fmt.Errorf("error getting rows from customer data sheet: %v", err)
		}
	}

	// Iterate over rows and populate the model
	isFirstRow := true
	for _, row := range rows {
		if isFirstRow {
			isFirstRow = false
			continue
		}

		// Make sure we don't go out of bounds with row data
		var rowData ExlData
		rowData.Branch = getRowValue(row, 0)
		rowData.CustId = getRowValue(row, 1)
		rowData.CustName = getRowValue(row, 2)
		rowData.Alamat = getRowValue(row, 3)
		rowData.Kota = getRowValue(row, 4)
		rowData.SalesName = getRowValue(row, 5)
		rowData.Channel = getRowValue(row, 6)
		rowData.Avg2023 = getRowValue(row, 7)
		rowData.Avg2024 = getRowValue(row, 8)
		rowData.Avg2025 = getRowValue(row, 9)
		rowData.Max = getRowValue(row, 10)

		exlData = append(exlData, rowData)
	}

	return exlData, nil
}

func openLocalProductFile(filePath string) ([]ExlProduct, error) {
	var exlData []ExlProduct

	// Make sure the path is absolute
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute file path: %v", err)
	}

	// Open the Excel file with Excelize
	exlz, err := excelize.OpenFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("error opening excel file: %v", err)
	}
	defer exlz.Close()

	// Try to get rows from "APL" sheet first (per user's info)
	rows, err := exlz.GetRows("APL")
	if err != nil {
		// If "APL" sheet doesn't exist, try alternate names
		alternateNames := []string{"DaftarHarga", "Product", "Products", "Sheet2"}
		var rowsFound bool

		for _, sheetName := range alternateNames {
			rows, err = exlz.GetRows(sheetName)
			if err == nil {
				rowsFound = true
				log.Printf("Found product data in sheet: %s", sheetName)
				break
			}
		}

		if !rowsFound {
			return nil, fmt.Errorf("error getting rows from product data sheet: %v", err)
		}
	}

	// Iterate over rows and populate the model
	// Skip header row by starting at index 3 (row 4 in Excel)
	// Based on screenshot, data starts from row 4 (index 3)
	for index, row := range rows {
		// Skip the header rows (0-2) as per the screenshot
		if index < 3 {
			continue
		}

		var rowData ExlProduct
		rowData.Code = getRowValue(row, 0)        // Column A - KODE APL
		rowData.NameProduct = getRowValue(row, 1) // Column B - PRODUK
		rowData.HNA = getRowValue(row, 2)         // Column C - HNA

		// If there's a fourth column for PPN, get it (not shown in screenshot)
		if len(row) > 3 {
			rowData.PPN = getRowValue(row, 3)
		} else {
			rowData.PPN = ""
		}

		exlData = append(exlData, rowData)
	}
	return exlData, nil
}

// Helper function to safely get values from rows
func getRowValue(row []string, index int) string {
	if index < len(row) {
		return handleNullValue(row[index])
	}
	return " "
}

func handleNullValue(value string) string {
	if value == "" {
		return " "
	}
	return value
}
