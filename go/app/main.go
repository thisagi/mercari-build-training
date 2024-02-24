package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"crypto/sha256"
	"io"
	"mime/multipart"
	"encoding/hex"
	"strconv"
	"database/sql"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "images"
	db_file = "../db/mercari.sqlite3"
)

type Item struct {
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	ImageName  string  `json:"image_name"`
}

type ItemList struct {
	Items    []Item   `json:"items"`
} 

type Response struct {
	Message string `json:"message"`
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

// dbの内容を読み込む
func openFileAndChengeStr(c echo.Context) (ItemList, error) {
	read_file, err := sql.Open("sqlite3",db_file)
	// dbを開く
	if err != nil {
		c.Logger().Fatalf("Cannot open the db")
	}
	defer read_file.Close()

	// dbを読み込む
	row_data, err := read_file.Query("SELECT name, category, image_name FROM items")
	if err != nil {
		c.Logger().Fatalf("Cannot read the db data")
	}
	defer row_data.Close()

	// 読み込んだ内容を構造体に変換
	var items ItemList
	for row_data.Next(){
		var it Item
		err := row_data.Scan(&it.Name, &it.Category, &it.ImageName)
		if err != nil {
			c.Logger().Fatalf("Cannot read the db data")
		}
		items.Items = append(items.Items, it)
	}

	return items, nil
}

// imageのハッシュ生成
func imageHash(img_file *multipart.FileHeader) (string, error) {
	// 画像ファイルを開く
	img, err := img_file.Open()
	if err != nil {
		fmt.Printf("Cannot open the image\n")
	}
	defer img.Close()

	// hash値計算
	hash := sha256.New()
	if _, err := io.Copy(hash, img); err != nil {
		fmt.Printf("Hash error\n")
	}
	img_name := hex.EncodeToString(hash.Sum(nil)) + ".jpg"
	file_path := ImgDir + "/"+ img_name
	
	// 内容を保存
	file, err := os.Create(file_path)
	if err != nil {
		fmt.Printf("Cannot open the file\n")
	}
	defer file.Close()
	if _, err = io.Copy(file, img); err != nil {
		fmt.Printf("Cannot copy the image\n")
	}

	return img_name, err
}

// アイテムの追加
func addItem(c echo.Context) error {
	// Get form data name and category
	name := c.FormValue("name")
	category := c.FormValue("category")
	
	// 画像をハッシュ化
	img_file, err := c.FormFile("image")
    if err != nil {
        c.Logger().Fatalf("Image retrieval error: %v", err)
    }
	img_name, err := imageHash(img_file)
	if err != nil {
        c.Logger().Fatalf("Hash conversion error: %v", err)
    }

	c.Logger().Infof("Receive item: %s category: %s img_name: %s", name, category, img_name)

	// dbに書き込む
	// dbを開く
	read_file, err := sql.Open("sqlite3",db_file)
	if err != nil {
		c.Logger().Fatalf("Cannot open the db")
	}
	defer read_file.Close()

	// dbに追加
	_, err = read_file.Exec("INSERT INTO items (name, category, image_name) VALUES (?, ?, ?)", 
													name, category, img_name)
	if err != nil {
		c.Logger().Fatalf("Cannot Insert the db")
	}

	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}
	return c.JSON(http.StatusOK, res)
}

// 保存されているアイテムの表示
func getItem(c echo.Context) error {
	// ファイルから読み込み
	items, err := openFileAndChengeStr(c)
	if err != nil {
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, items)
}

// idによるアイテムの表示
func getIdItem(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	
	// ファイル読み込み
	items, err := openFileAndChengeStr(c)
	if err != nil { c.Logger().Fatalf("Cannot read the file %v",err) }

	// 存在するidならitem情報を返す
	if id < 1 || len(items.Items) < id {
		res := "ID does not exist"
		return c.JSON(http.StatusBadRequest, res)
	} else {
		return c.JSON(http.StatusOK, items.Items[id-1])
	}
}

func searchItem(c echo.Context) error {
	// 
	keyword := c.QueryParam("keyword")

	read_file, err := sql.Open("sqlite3",db_file)
	// dbを開く
	if err != nil {
		c.Logger().Fatalf("Cannot open the db")
	}
	defer read_file.Close()

	// dbを読み込む
	row_data, err := read_file.Query("SELECT name, category, image_name FROM items WHERE name LIKE ?",
																					"%"+keyword+"%")
	if err != nil {
		c.Logger().Fatalf("Cannot read the search data")
	}
	defer row_data.Close()

	// 読み込んだ内容を構造体に変換
	var items ItemList
	for row_data.Next(){
		var it Item
		err := row_data.Scan(&it.Name, &it.Category, &it.ImageName)
		if err != nil {
			c.Logger().Fatalf("Cannot read the db data")
		}
		items.Items = append(items.Items, it)
	}

	return c.JSON(http.StatusOK, items)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		//c.Logger().Infof("Image not found: %s", imgPath)
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
	//e.Logger.SetLevel(log.INFO)
	e.Logger.SetLevel(log.DEBUG)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	e.GET("/", root)
	e.GET("/items", getItem)
	e.GET("/items/:id", getIdItem)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)
	e.GET("/search", searchItem)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
