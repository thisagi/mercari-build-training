package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"encoding/json"
	"crypto/sha256"
	"io"
	"mime/multipart"
	"encoding/hex"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "images"
	item_file = "../db/items.json"
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

// ファイルを開いてbyteデータで内容を読み込む
func openFileAndInput(c echo.Context) ([]byte, error) {
	read_file, err := os.Open(item_file)
	if err != nil {
		fmt.Printf("JSONファイルを開けません")
		return nil, err
	}
	defer read_file.Close()

	// ファイルを読み込む
	inputJsonData, err := os.ReadFile(item_file)
	if err != nil {
		fmt.Printf("JSONデータを読み込めません")
		return nil, err
	}

	return inputJsonData, err
}

func chengeStr(inputJsonData []byte) (ItemList, error){
	// ファイルの内容を構造体に変換
	var items ItemList
	if err := json.Unmarshal(inputJsonData, &items); err != nil {
		fmt.Printf("構造体に変換できません")
		return items, err
	}
	fmt.Printf("読み込んだファイル内容 \n %+v\n", items)
	return items, nil
}

// ファイルに書き込みを行う
func writeFile(items ItemList) error {
	file, err := os.Create(item_file)
	if err != nil {
		fmt.Printf("JSONファイルを開けません")
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(items); err != nil {
		fmt.Printf("書き込みができません")
	}
	return err
}

// imageのハッシュ生成
func imageHash(img_file *multipart.FileHeader) (string, error) {
	// file open
	img, err := img_file.Open()
	if err != nil {
		fmt.Printf("イメージオープンエラー")
	}
	defer img.Close()

	// hash計算
	hash := sha256.New()
	if _, err := io.Copy(hash, img); err != nil {
		fmt.Printf("ハッシュエラー")
	}

	img_name := hex.EncodeToString(hash.Sum(nil)) + ".jpg"

	file_path := ImgDir + "/"+ img_name
	
	file, err := os.Create(file_path)
	if err != nil {
		fmt.Printf("オープンエラー")
	}
	defer file.Close()

	if _, err = io.Copy(file, img); err != nil {
		fmt.Printf("書き込みエラー")
	}

	return img_name, err
}

func addItem(c echo.Context) error {
	// Get form data name and category
	name := c.FormValue("name")
	category := c.FormValue("category")
	
	// 画像に関する取得
	img_file, err := c.FormFile("image")
    if err != nil {
        c.Logger().Fatalf("画像取得エラー: %v", err)
    }	
	
	img_name, err := imageHash(img_file)
	if err != nil {
        c.Logger().Fatalf("ハッシュ変換エラー: %v", err)
    }	

	fmt.Printf("Receive item: %s \ncategory: %s \nimg_name: %s \n", name, category, img_name)
	
	// 受け取った名前とカテゴリーをItem構造体へ変換
	new_item := Item{ name, category, img_name}
	
	// ファイル内容を読み込み
	inputJsonData, err := openFileAndInput(c)
	if err != nil { c.Logger().Fatalf("ファイル読み込みに失敗しました %v",err) }

	// ファイルを構造体に変換
	items, err := chengeStr(inputJsonData)
	if err != nil { c.Logger().Fatalf("JSONデータを変換できません %v",err) }

	// 読み込んだ内容に今回の内容を追加する
	items.Items = append(items.Items, new_item)
	fmt.Printf("%+v\n", items)

	// 書き込み
	err = writeFile(items)
	if err != nil { c.Logger().Fatalf("ファイルへの書き込みに失敗しました %v",err) }

	// response
	message := fmt.Sprintf("item received: %s", name)
	res := Response{Message: message}
	return c.JSON(http.StatusOK, res)
}

func getItem(c echo.Context) error {
	inputJsonData, err := openFileAndInput(c)
	if err != nil { c.Logger().Fatalf("ファイル読み込みに失敗しました %v",err) }

	items, err := chengeStr(inputJsonData)
	if err != nil { c.Logger().Fatalf("ファイルの変換に失敗しました %v",err) }

	bytes, err := json.Marshal(items)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("内容を出力")
	fmt.Println(string(bytes))
	return c.JSONBlob(http.StatusOK, inputJsonData)
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

func getIdItem(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	// ファイル内容を読み込み
	inputJsonData, err := openFileAndInput(c)
	if err != nil { c.Logger().Fatalf("ファイル読み込みに失敗しました %v",err) }

	// ファイルを構造体に変換
	items, err := chengeStr(inputJsonData)
	if err != nil { c.Logger().Fatalf("JSONデータを変換できません %v",err) }

	return c.JSON(http.StatusOK, items.Items[id-1])
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

	e.GET("/", root)
	e.GET("/items", getItem)
	e.GET("/items/:id", getIdItem)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)


	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
