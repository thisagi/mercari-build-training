package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
)

type Item struct {
	Name      string  `json:"name"`
	Category  string  `json:"category"`
}

type ItemList struct {
	ItemList    []Item   `json:"items"`
} 

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

// POSTされたときに読み込みが行われる
func addItem(c echo.Context) error {
	// itemファイルの指定
	item_file := "../db/items.json"

	// Get form data name and category
	name := c.FormValue("name")
	category := c.FormValue("category")
	fmt.Printf("Receive item: %s category: %s\n", name, category)

	// 受け取った名前とカテゴリーをItem構造体へ変換
	new_item := Item{ name, category }
	
	// ファイルから読み込みを行う
	// fileを開く
	read_file, err := os.Open(item_file)
	if err != nil {
		c.Logger().Fatalf("JSONファイルを開けません %v",err)
	}
	defer read_file.Close()
	// ファイルを読み込む
	inputJsonData, err := os.ReadFile(item_file)
	if err != nil {
		c.Logger().Fatalf("JSONデータを読み込めません %v",err)
	}
	// ファイルの内容を構造体に変換
	var items ItemList
	if err := json.Unmarshal(inputJsonData, &items); err != nil {
		c.Logger().Fatalf("JSONデータを変換できません %v",err)
	}
	fmt.Printf("%+v\n", items.ItemList)

	// 読み込んだ内容に今回の内容を追加する
	items.ItemList = append(items.ItemList, new_item)
	fmt.Printf("%+v\n", items)

	// ファイルに書き込みを行う
	file, err := os.Create(item_file)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(items); err != nil {
		log.Fatal(err)
	}

	// response
	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}
	return c.JSON(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	// getをしたときにrootに飛ぶようになっている
	e.GET("/", root)
	e.POST("/items", addItem)
	// get
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
