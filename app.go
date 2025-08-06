package main

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// App struct
type App struct {
	ctx context.Context
	db  *sql.DB
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化数据库
	err := a.initDatabase()
	if err != nil {
		fmt.Printf("数据库初始化失败: %v\n", err)
	}
}

// shutdown is called when the app shuts down
func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		a.db.Close()
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// HandleFileDrop handles dropped files
func (a *App) HandleFileDrop(files []string) {
	fmt.Printf("Dropped files: %v\n", files)

	// 处理每个拖放的文件
	for _, file := range files {
		// 检查文件信息
		fileInfo, err := os.Stat(file)
		if err != nil {
			fmt.Printf("获取文件信息失败 %s: %v\n", file, err)
			continue
		}

		var firstImage string
		var fileType string

		if fileInfo.IsDir() {
			// 处理普通文件夹
			firstImage, err = a.getFirstImageFromFolder(file)
			if err != nil {
				fmt.Printf("读取文件夹 %s 失败: %v\n", file, err)
				continue
			}
			fileType = "folder"
			fmt.Printf("从文件夹 %s 中读取到第一张图片: %s\n", file, firstImage)
		} else if strings.HasSuffix(strings.ToLower(file), ".zip") {
			// 处理zip文件
			firstImage, err = a.getFirstImageFromZip(file)
			if err != nil {
				fmt.Printf("读取zip文件 %s 失败: %v\n", file, err)
				continue
			}
			fileType = "zip"
			fmt.Printf("从 %s 中读取到第一张图片: %s\n", file, firstImage)
		} else {
			continue // 跳过不支持的文件类型
		}

		// 保存到数据库
		err = a.saveComicToDatabase(file, fileType, firstImage, fileInfo.Size())
		if err != nil {
			fmt.Printf("保存到数据库失败 %s: %v\n", file, err)
		} else {
			fmt.Printf("成功保存到数据库: %s\n", file)
		}
	}
}

// getFirstImageFromZip 以流的方式读取zip中的第一个图片（广度优先搜索子目录）
func (a *App) getFirstImageFromZip(zipPath string) (string, error) {
	// 打开zip文件
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("打开zip文件失败: %v", err)
	}
	defer reader.Close()

	// 使用广度优先搜索获取所有图片文件
	imageFiles := a.bfsSearchImages(reader.File)

	fmt.Printf("找到 %d 个图片文件\n", len(imageFiles))
	if len(imageFiles) > 0 {
		fmt.Printf("图片文件列表: %v\n", imageFiles)
	}

	if len(imageFiles) == 0 {
		return "", fmt.Errorf("zip文件中没有找到图片文件")
	}

	// 按照自然顺序排序文件名
	fmt.Printf("排序前的文件列表: %v\n", imageFiles)
	sort.Slice(imageFiles, func(i, j int) bool {
		return a.naturalSort(imageFiles[i], imageFiles[j])
	})
	fmt.Printf("排序后的文件列表: %v\n", imageFiles)

	// 获取第一个图片文件
	firstImageName := imageFiles[0]

	// 尝试URL解码处理中文路径
	decodedImageName, err := url.QueryUnescape(firstImageName)
	if err != nil {
		fmt.Printf("Warning: Failed to decode image name, using original: %v\n", err)
		decodedImageName = firstImageName
	}

	fmt.Printf("Original image name: %s\n", firstImageName)
	fmt.Printf("Decoded image name: %s\n", decodedImageName)

	// 找到对应的zip文件条目
	var firstImageFile *zip.File
	for _, file := range reader.File {
		// 尝试多种编码方式匹配文件名
		if file.Name == firstImageName || file.Name == decodedImageName {
			firstImageFile = file
			fmt.Printf("Found exact match: %s\n", file.Name)
			break
		}
	}

	if firstImageFile == nil {
		// 如果没找到，尝试模糊匹配
		fmt.Printf("Exact match not found, trying fuzzy match...\n")
		for _, file := range reader.File {
			// 检查文件名是否包含目标图片的基本信息
			if strings.Contains(file.Name, filepath.Base(decodedImageName)) {
				firstImageFile = file
				fmt.Printf("Found file by fuzzy match: %s\n", file.Name)
				break
			}
		}
	}

	if firstImageFile == nil {
		return "", fmt.Errorf("无法找到排序后的第一个图片文件: %s (decoded: %s)", firstImageName, decodedImageName)
	}

	// 以流的方式读取图片数据
	rc, err := firstImageFile.Open()
	if err != nil {
		return "", fmt.Errorf("打开图片文件失败: %v", err)
	}
	defer rc.Close()

	// 读取图片数据到内存（这里可以根据需要修改为直接处理或保存到文件）
	imageData, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("读取图片数据失败: %v", err)
	}

	fmt.Printf("成功读取图片 %s，大小: %d 字节\n", firstImageName, len(imageData))
	fmt.Printf("zip文件路径 %s\n", zipPath)
	return zipPath + "!" + firstImageName, nil
}

// getFirstImageFromFolder 从普通文件夹中读取第一个图片（广度优先搜索子目录）
func (a *App) getFirstImageFromFolder(folderPath string) (string, error) {
	// 使用广度优先搜索获取所有图片文件
	imageFiles, err := a.bfsSearchImagesFromFolder(folderPath)
	if err != nil {
		return "", fmt.Errorf("搜索文件夹失败: %v", err)
	}

	fmt.Printf("找到 %d 个图片文件\n", len(imageFiles))
	if len(imageFiles) > 0 {
		fmt.Printf("图片文件列表: %v\n", imageFiles)
	}

	if len(imageFiles) == 0 {
		return "", fmt.Errorf("文件夹中没有找到图片文件")
	}

	// 按照自然顺序排序文件名
	fmt.Printf("排序前的文件列表: %v\n", imageFiles)
	sort.Slice(imageFiles, func(i, j int) bool {
		return a.naturalSort(imageFiles[i], imageFiles[j])
	})
	fmt.Printf("排序后的文件列表: %v\n", imageFiles)

	// 获取第一个图片文件
	firstImagePath := imageFiles[0]

	// 读取图片文件信息
	fileInfo, err := os.Stat(firstImagePath)
	if err != nil {
		return "", fmt.Errorf("获取图片文件信息失败: %v", err)
	}

	fmt.Printf("成功读取图片 %s，大小: %d 字节\n", firstImagePath, fileInfo.Size())

	return firstImagePath, nil
}

// initDatabase 初始化SQLite数据库
func (a *App) initDatabase() error {
	// 数据库文件路径
	dbPath := "comic.db"

	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	a.db = db

	// 创建表
	err = a.createTables()
	if err != nil {
		return fmt.Errorf("创建表失败: %v", err)
	}

	fmt.Printf("数据库初始化成功: %s\n", dbPath)
	return nil
}

// createTables 创建数据库表
func (a *App) createTables() error {
	// 创建漫画信息表
	createComicTable := `
	CREATE TABLE IF NOT EXISTS comics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		file_path TEXT UNIQUE NOT NULL,
		file_type TEXT NOT NULL,
		first_image TEXT,
		file_size INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// 创建图片信息表
	createImageTable := `
	CREATE TABLE IF NOT EXISTS images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		comic_id INTEGER NOT NULL,
		file_name TEXT NOT NULL,
		file_path TEXT NOT NULL,
		file_size INTEGER,
		width INTEGER,
		height INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (comic_id) REFERENCES comics (id)
	);`

	// 执行创建表语句
	_, err := a.db.Exec(createComicTable)
	if err != nil {
		return fmt.Errorf("创建comics表失败: %v", err)
	}

	_, err = a.db.Exec(createImageTable)
	if err != nil {
		return fmt.Errorf("创建images表失败: %v", err)
	}

	return nil
}

// bfsSearchImages 使用广度优先搜索算法搜索图片文件
func (a *App) bfsSearchImages(files []*zip.File) []string {
	var imageFiles []string
	imageExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".webp": true, ".tiff": true, ".tif": true,
	}

	// 构建目录结构
	dirs := make(map[string][]string)
	var rootFiles []string

	for _, file := range files {
		// 跳过目录本身
		if strings.HasSuffix(file.Name, "/") {
			continue
		}

		// 获取文件所在目录
		dir := filepath.Dir(file.Name)
		if dir == "." {
			// 根目录文件
			rootFiles = append(rootFiles, file.Name)
		} else {
			// 子目录文件
			dirs[dir] = append(dirs[dir], file.Name)
		}
	}

	// 广度优先搜索：按层级搜索
	// 1. 先检查根目录的图片文件
	for _, fileName := range rootFiles {
		ext := strings.ToLower(filepath.Ext(fileName))
		if imageExtensions[ext] {
			imageFiles = append(imageFiles, fileName)
		}
	}

	// 2. 如果根目录没有图片，按目录层级广度优先搜索子目录
	if len(imageFiles) == 0 {
		// 按目录层级排序（广度优先）
		sortedDirs := a.sortDirectoriesByLevel(dirs)

		// 遍历每个目录层级
		for _, dir := range sortedDirs {
			for _, fileName := range dirs[dir] {
				ext := strings.ToLower(filepath.Ext(fileName))
				if imageExtensions[ext] {
					imageFiles = append(imageFiles, fileName)
				}
			}
		}
	}

	return imageFiles
}

// sortDirectoriesByLevel 按目录层级排序，实现真正的广度优先
func (a *App) sortDirectoriesByLevel(dirs map[string][]string) []string {
	// 计算每个目录的层级深度
	dirLevels := make(map[string]int)
	for dir := range dirs {
		dirLevels[dir] = a.getDirectoryLevel(dir)
	}

	// 按层级分组
	levelGroups := make(map[int][]string)
	for dir, level := range dirLevels {
		levelGroups[level] = append(levelGroups[level], dir)
	}

	// 按层级顺序构建结果
	var result []string
	for level := 1; level <= len(levelGroups); level++ {
		if dirs, exists := levelGroups[level]; exists {
			// 同一层级内按目录名排序
			sort.Strings(dirs)
			result = append(result, dirs...)
		}
	}

	return result
}

// bfsSearchImagesFromFolder 使用广度优先搜索算法从文件夹中搜索图片文件
func (a *App) bfsSearchImagesFromFolder(folderPath string) ([]string, error) {
	var imageFiles []string
	imageExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".webp": true, ".tiff": true, ".tif": true,
	}

	fmt.Printf("开始搜索文件夹: %s\n", folderPath)

	// 使用广度优先搜索遍历文件夹
	queue := []string{folderPath}
	visited := make(map[string]bool)

	for len(queue) > 0 {
		currentPath := queue[0]
		queue = queue[1:]

		if visited[currentPath] {
			continue
		}
		visited[currentPath] = true

		// 读取当前目录
		entries, err := os.ReadDir(currentPath)
		if err != nil {
			continue // 跳过无法读取的目录
		}

		// 先处理文件（广度优先：同层级先处理文件）
		for _, entry := range entries {
			if !entry.IsDir() {
				filePath := filepath.Join(currentPath, entry.Name())
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if imageExtensions[ext] {
					imageFiles = append(imageFiles, filePath)
				}
			}
		}

		// 再添加子目录到队列（广度优先：同层级处理完文件后再处理子目录）
		for _, entry := range entries {
			if entry.IsDir() {
				subDirPath := filepath.Join(currentPath, entry.Name())
				queue = append(queue, subDirPath)
			}
		}
	}

	return imageFiles, nil
}

// getDirectoryLevel 获取目录的层级深度
func (a *App) getDirectoryLevel(dir string) int {
	if dir == "." {
		return 0
	}
	// 计算路径分隔符的数量来确定层级
	return strings.Count(dir, "/") + 1
}

// naturalSort 实现自然排序，支持混合字符串的数字排序
func (a *App) naturalSort(aStr, bStr string) bool {
	// 提取文件名（不含扩展名）
	aName := strings.TrimSuffix(filepath.Base(aStr), filepath.Ext(aStr))
	bName := strings.TrimSuffix(filepath.Base(bStr), filepath.Ext(bStr))

	// 使用自然排序比较
	return a.compareNatural(aName, bName)
}

// compareNatural 实现自然排序比较，支持字母数字混合
func (a *App) compareNatural(strA, strB string) bool {
	// 将字符串分割为数字和非数字部分
	aParts := a.splitStringAndNumber(strA)
	bParts := a.splitStringAndNumber(strB)

	// 比较每个部分
	minLen := len(aParts)
	if len(bParts) < minLen {
		minLen = len(bParts)
	}

	for i := 0; i < minLen; i++ {
		aPart := aParts[i]
		bPart := bParts[i]

		// 如果两个部分都是数字，按数字大小比较
		if a.isNumeric(aPart) && a.isNumeric(bPart) {
			aNum := a.parseNumber(aPart)
			bNum := a.parseNumber(bPart)
			if aNum != bNum {
				return aNum < bNum
			}
		} else {
			// 否则按字符串比较
			if aPart != bPart {
				return aPart < bPart
			}
		}
	}

	// 如果前面部分都相同，长度短的排在前面
	return len(aParts) < len(bParts)
}

// splitStringAndNumber 将字符串分割为数字和非数字部分
func (a *App) splitStringAndNumber(s string) []string {
	var parts []string
	var current string
	var isDigit bool

	for _, char := range s {
		charIsDigit := char >= '0' && char <= '9'

		if len(current) == 0 {
			// 第一个字符
			current = string(char)
			isDigit = charIsDigit
		} else if charIsDigit == isDigit {
			// 相同类型，继续累积
			current += string(char)
		} else {
			// 类型改变，保存当前部分并开始新部分
			parts = append(parts, current)
			current = string(char)
			isDigit = charIsDigit
		}
	}

	// 添加最后一部分
	if len(current) > 0 {
		parts = append(parts, current)
	}

	return parts
}

// isNumeric 检查字符串是否为纯数字
func (a *App) isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return len(s) > 0
}

// parseNumber 解析数字字符串
func (a *App) parseNumber(s string) int {
	var num int
	fmt.Sscanf(s, "%d", &num)
	return num
}

// saveComicToDatabase 保存漫画信息到数据库
func (a *App) saveComicToDatabase(filePath, fileType, firstImage string, fileSize int64) error {
	if a.db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	// 获取文件名作为标题
	title := filepath.Base(filePath)

	// 插入或更新漫画信息
	query := `
	INSERT OR REPLACE INTO comics (title, file_path, file_type, first_image, file_size, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	_, err := a.db.Exec(query, title, filePath, fileType, firstImage, fileSize, time.Now())
	if err != nil {
		return fmt.Errorf("插入漫画信息失败: %v", err)
	}

	return nil
}

// GetComicsFromDatabase 从数据库获取所有漫画信息
func (a *App) GetComicsFromDatabase() ([]map[string]interface{}, error) {
	if a.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	query := `SELECT id, title, file_path, file_type, first_image, file_size, created_at, updated_at FROM comics ORDER BY updated_at DESC`

	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询漫画信息失败: %v", err)
	}
	defer rows.Close()

	var comics []map[string]interface{}
	for rows.Next() {
		var id int64
		var title, filePath, fileType, firstImage, createdAt, updatedAt string
		var fileSize int64

		err := rows.Scan(&id, &title, &filePath, &fileType, &firstImage, &fileSize, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		comic := map[string]interface{}{
			"id":         id,
			"title":      title,
			"filePath":   filePath,
			"fileType":   fileType,
			"firstImage": firstImage,
			"fileSize":   fileSize,
			"createdAt":  createdAt,
			"updatedAt":  updatedAt,
		}
		comics = append(comics, comic)
	}

	return comics, nil
}

// DeleteComicFromDatabase 从数据库删除漫画信息
func (a *App) DeleteComicFromDatabase(comicID int64) error {
	if a.db == nil {
		return fmt.Errorf("数据库未初始化")
	}

	query := `DELETE FROM comics WHERE id = ?`
	_, err := a.db.Exec(query, comicID)
	if err != nil {
		return fmt.Errorf("删除漫画信息失败: %v", err)
	}

	return nil
}

// SearchComicsInDatabase 在数据库中搜索漫画
func (a *App) SearchComicsInDatabase(keyword string) ([]map[string]interface{}, error) {
	if a.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	query := `SELECT id, title, file_path, file_type, first_image, file_size, created_at, updated_at 
			  FROM comics 
			  WHERE title LIKE ? OR file_path LIKE ? 
			  ORDER BY updated_at DESC`

	searchPattern := "%" + keyword + "%"
	rows, err := a.db.Query(query, searchPattern, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("搜索漫画信息失败: %v", err)
	}
	defer rows.Close()

	var comics []map[string]interface{}
	for rows.Next() {
		var id int64
		var title, filePath, fileType, firstImage, createdAt, updatedAt string
		var fileSize int64

		err := rows.Scan(&id, &title, &filePath, &fileType, &firstImage, &fileSize, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		comic := map[string]interface{}{
			"id":         id,
			"title":      title,
			"filePath":   filePath,
			"fileType":   fileType,
			"firstImage": firstImage,
			"fileSize":   fileSize,
			"createdAt":  createdAt,
			"updatedAt":  updatedAt,
		}
		comics = append(comics, comic)
	}

	return comics, nil
}

// GetImageData 获取图片数据，支持comic://协议
func (a *App) GetImageData(imagePath string) ([]byte, error) {
	// 移除comic://协议前缀
	if strings.HasPrefix(imagePath, "comic://") {
		imagePath = strings.TrimPrefix(imagePath, "comic://")
	}

	// 检查文件是否存在
	if _, err := os.Stat(imagePath); err != nil {
		return nil, fmt.Errorf("图片文件不存在: %v", err)
	}

	// 读取图片文件
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("读取图片文件失败: %v", err)
	}

	return imageData, nil
}

// GetImageBase64 获取图片的Base64编码，用于前端显示
func (a *App) GetImageBase64(imagePath string) (string, error) {
	imageData, err := a.GetImageData(imagePath)
	if err != nil {
		return "", err
	}

	// 获取文件扩展名来确定MIME类型
	ext := strings.ToLower(filepath.Ext(imagePath))
	var mimeType string
	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".bmp":
		mimeType = "image/bmp"
	case ".webp":
		mimeType = "image/webp"
	case ".tiff", ".tif":
		mimeType = "image/tiff"
	default:
		mimeType = "image/jpeg" // 默认
	}

	// 转换为Base64
	base64Data := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(imageData))
	return base64Data, nil
}
