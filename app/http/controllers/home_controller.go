package controllers

import (
	"github.com/goravel/framework/contracts/http"

	"bytes"
	"fmt"
	"io/ioutil"
	https "net/http"

	excelize "github.com/xuri/excelize/v2"
)

type HomeController struct {
	//Dependent services
}

func NewHomeController() *HomeController {
	return &HomeController{
		//Inject services
	}
}

func (r *HomeController) Update(ctx http.Context) http.Response {
	// Read the request body\
	input := ctx.Request().Input("input")

	// Do something with the text (e.g., print it)
	return ctx.Response().Json(http.StatusOK, http.Json{
		"message": input,
	})
}

func (r *HomeController) Index(ctx http.Context) http.Response {
	data, err := openURLCust("https://www.dropbox.com/scl/fi/ga9aesugfhxrt2dmuknre/Data-base-aplikasi-bayer-joglopwk-160224.xlsx?rlkey=4x85x8rdq9r3x7wyzgxjnofki&dl=1")
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"message": err,
		})
	}
	dataProduct, err := openURLItem("https://www.dropbox.com/scl/fi/ga9aesugfhxrt2dmuknre/Data-base-aplikasi-bayer-joglopwk-160224.xlsx?rlkey=4x85x8rdq9r3x7wyzgxjnofki&dl=1")
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"message": err,
		})
	}
	return ctx.Response().View().Make("home.html", map[string]any{
		"customer": data,
		"produck": dataProduct,
	})
}

type ExlProduct struct {
	Branch    string
	CustId    string
	CustName  string
	Alamat    string
	Kota      string
	SalesName string
	Channel   string
	Avg2023   string
	Q4Avg2023 string
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
	Q4Avg2023 string
}

func openURLCust(urlLink string) ([]ExlData, error) {
	var exlData []ExlData
	data, err := getData(urlLink)
	if err != nil {
		panic(err)
	}

	// Open the ZIP file with Excelize
	exlz, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		fmt.Println("Reader", err)
		return nil, err
	}

	lst := exlz.GetSheetList()
	if len(lst) == 0 {
		fmt.Println("Empty document")
		return nil, err
	}

	fmt.Println("Sheet list:")
	for _, s := range lst {
		fmt.Println(s)
	}

	defer func() {
		if err = exlz.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Println("Done")
	rows, err := exlz.GetRows("Sheet1")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Iterate over rows and populate the model
	isFirstRow := true
	for _, row := range rows {
		if isFirstRow {
			isFirstRow = false
			continue
		}
		rowData := ExlData{
			Branch:    handleNullValue(row[0]),
			CustId:    handleNullValue(row[1]),
			CustName:  handleNullValue(row[2]),
			Alamat:    handleNullValue(row[3]),
			Kota:      handleNullValue(row[4]),
			SalesName: handleNullValue(row[5]),
			Channel:   handleNullValue(row[6]),
			Avg2023:   handleNullValue(row[7]),
			Q4Avg2023: handleNullValue(row[8]),
		}
		exlData = append(exlData, rowData)
	}
	return exlData, nil
}

func getData(url string) ([]byte, error) {

	r, err := https.Get(url)
	if err != nil {
		panic(err)
	}

	defer r.Body.Close()

	return ioutil.ReadAll(r.Body)
}

func handleNullValue(value string) string {
	if value == "" {
		return " "
	}
	return value
}
