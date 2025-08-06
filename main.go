package main

import (
	"archive/zip"
	"embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type FileLoader struct {
	http.Handler
}

func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

func (h *FileLoader) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var err error

	println("=== HTTP Request Received ===")
	println("Time:", fmt.Sprintf("%v", req.Header.Get("Date")))

	// 输出完整的URL信息
	println("=== HTTP Request Details ===")
	println("Full URL:", req.URL.String())
	println("Scheme:", req.URL.Scheme)
	println("Host:", req.URL.Host)
	println("Path:", req.URL.Path)
	println("Raw Query:", req.URL.RawQuery)
	println("Fragment:", req.URL.Fragment)
	println("Method:", req.Method)
	println("Remote Address:", req.RemoteAddr)
	println("User Agent:", req.UserAgent())
	println("==========================")

	requestedFilename := req.URL.Path
	println("Requesting file:", requestedFilename)
	println("Contains '!':", strings.Contains(requestedFilename, "!"))

	// 检查是否是ZIP文件中的图片请求
	if strings.Contains(requestedFilename, "!") {
		println("=== Processing ZIP Image Request ===")
		h.handleZipImage(res, requestedFilename)
		println("=== ZIP Image Request Processing Complete ===")
		return
	}

	println("=== Processing Regular File Request ===")
	// 处理普通文件请求
	fileData, err := os.ReadFile(requestedFilename)
	if err != nil {
		println("Error reading file:", err.Error())
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(fmt.Sprintf("Could not load file %s", requestedFilename)))
		return
	}

	println("Successfully read file, size:", len(fileData), "bytes")
	res.Write(fileData)
	println("=== Regular File Request Complete ===")
}

// handleZipImage 处理ZIP文件中的图片请求
func (h *FileLoader) handleZipImage(res http.ResponseWriter, requestPath string) {
	println("=== ZIP Image Request ===")
	println("Request path:", requestPath)

	// 分割路径：zip文件路径!图片路径
	parts := strings.SplitN(requestPath, "!", 2)
	if len(parts) != 2 {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Invalid format. Expected: zipfile!imagepath"))
		return
	}

	zipFilePath := parts[0]
	imagePath := parts[1]

	// URL解码处理中文路径
	decodedImagePath, err := url.QueryUnescape(imagePath)
	if err != nil {
		println("Warning: Failed to decode image path, using original:", err.Error())
		decodedImagePath = imagePath
	}

	println("ZIP file path:", zipFilePath)
	println("Original image path in ZIP:", imagePath)
	println("Decoded image path in ZIP:", decodedImagePath)

	// 检查ZIP文件是否存在
	if _, err := os.Stat(zipFilePath); err != nil {
		res.WriteHeader(http.StatusNotFound)
		res.Write([]byte(fmt.Sprintf("ZIP file not found: %s", zipFilePath)))
		return
	}

	// 打开ZIP文件
	zipReader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(fmt.Sprintf("Could not open ZIP file: %s", err.Error())))
		return
	}
	defer zipReader.Close()

	// 查找指定的图片文件
	var targetFile *zip.File
	for _, file := range zipReader.File {
		// 尝试多种编码方式匹配文件名
		if file.Name == decodedImagePath || file.Name == imagePath {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		// 如果没找到，尝试模糊匹配（处理编码问题）
		println("Exact match not found, trying fuzzy match...")
		for _, file := range zipReader.File {
			println("Available file in ZIP:", file.Name)
			// 检查文件名是否包含目标图片的基本信息
			if strings.Contains(file.Name, filepath.Base(decodedImagePath)) {
				targetFile = file
				println("Found file by fuzzy match:", file.Name)
				break
			}
		}
	}

	if targetFile == nil {
		res.WriteHeader(http.StatusNotFound)
		res.Write([]byte(fmt.Sprintf("Image not found in ZIP: %s (decoded: %s)", imagePath, decodedImagePath)))
		return
	}

	println("Found image in ZIP:", targetFile.Name)
	println("Image size:", targetFile.UncompressedSize64)

	// 打开ZIP中的文件
	fileReader, err := targetFile.Open()
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(fmt.Sprintf("Could not open image in ZIP: %s", err.Error())))
		return
	}
	defer fileReader.Close()

	// 设置正确的Content-Type
	ext := strings.ToLower(filepath.Ext(imagePath))
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".bmp":
		contentType = "image/bmp"
	case ".webp":
		contentType = "image/webp"
	case ".tiff", ".tif":
		contentType = "image/tiff"
	default:
		contentType = "image/jpeg"
	}

	res.Header().Set("Content-Type", contentType)
	res.Header().Set("Cache-Control", "public, max-age=3600") // 缓存1小时
	res.Header().Set("Content-Length", fmt.Sprintf("%d", targetFile.UncompressedSize64))

	// 以流的方式传输图片数据
	written, err := io.Copy(res, fileReader)
	if err != nil {
		println("Error streaming image:", err.Error())
		return
	}

	println("Successfully streamed image:", written, "bytes")
	println("=== ZIP Image Request Complete ===")
}

func main() {
	println("=== Starting R-Comic Application ===")
	println("Initializing HTTP server with custom file loader...")

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "r-comic",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: NewFileLoader(),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
