package storage

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"mime/multipart"
	"path"
	"strings"

	"github.com/nfnt/resize"
)

const (
	// 文件大小限制
	MaxFileSize  = 10 * 1024 * 1024  // 单个文件最大 10MB
	MaxTotalSize = 100 * 1024 * 1024 // 每个用户的总存储限制 100MB

	// 图片处理相关
	MaxImageDimension = 4096     // 最大图片尺寸
	ThumbnailSize     = 300      // 缩略图尺寸
	ThumbnailSuffix   = "_thumb" // 缩略图后缀

	// 允许的文件类型
	AllowedImageTypes = ".jpg,.jpeg,.png,.gif,.webp"

	// 存储目录
	AvatarDirectory = "avatars" // 用户头像目录
	TopicDirectory  = "topics"  // 话题图片目录
	ChatDirectory   = "chats"   // 聊天媒体目录
	TempDirectory   = "temp"    // 临时文件目录
)

// IsImageTypeAllowed 检查图片类型是否允许
func IsImageTypeAllowed(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))
	return strings.Contains(AllowedImageTypes, ext)
}

// GetStorageDirectory 根据文件类型获取存储目录
func GetStorageDirectory(fileType string) string {
	switch strings.ToLower(fileType) {
	case "avatar":
		return AvatarDirectory
	case "topic":
		return TopicDirectory
	case "chat":
		return ChatDirectory
	default:
		return TempDirectory
	}
}

// GenerateThumbPath 生成缩略图路径
func GenerateThumbPath(originalPath string) string {
	ext := path.Ext(originalPath)
	basePath := strings.TrimSuffix(originalPath, ext)
	return fmt.Sprintf("%s%s%s", basePath, ThumbnailSuffix, ext)
}

// ValidateFile 验证文件
func ValidateFile(file *multipart.FileHeader) error {
	// 检查文件大小
	if file.Size > MaxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", MaxFileSize)
	}

	// 检查文件类型
	ext := strings.ToLower(path.Ext(file.Filename))
	if !strings.Contains(AllowedImageTypes, ext) {
		return fmt.Errorf("file type %s is not allowed", ext)
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	// 检查图片尺寸
	img, _, err := image.DecodeConfig(src)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	if img.Width > MaxImageDimension || img.Height > MaxImageDimension {
		return fmt.Errorf("image dimensions exceed maximum allowed size of %dx%d", MaxImageDimension, MaxImageDimension)
	}

	return nil
}

// ResizeImage 调整图片大小
func ResizeImage(img image.Image, maxWidth, maxHeight uint) image.Image {
	return resize.Thumbnail(maxWidth, maxHeight, img, resize.Lanczos3)
}

// GetFileDirectory 根据文件类型获取存储目录
func GetFileDirectory(fileType string) string {
	switch strings.ToLower(fileType) {
	case "avatar":
		return AvatarDirectory
	case "topic":
		return TopicDirectory
	case "chat":
		return ChatDirectory
	default:
		return "misc"
	}
}

// GetImageDimensions 获取图片尺寸
func GetImageDimensions(file *multipart.FileHeader) (width, height int, err error) {
	src, err := file.Open()
	if err != nil {
		return 0, 0, err
	}
	defer src.Close()

	img, _, err := image.DecodeConfig(src)
	if err != nil {
		return 0, 0, err
	}

	return img.Width, img.Height, nil
}
