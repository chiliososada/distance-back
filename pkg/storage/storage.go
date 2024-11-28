package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"DistanceBack_v1/config"
	"DistanceBack_v1/pkg/auth" // 导入 auth 包
	"DistanceBack_v1/pkg/logger"

	"cloud.google.com/go/storage"
)

// Storage 定义存储接口
type Storage interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, directory string) (string, error)
	DeleteFile(ctx context.Context, fileURL string) error
}

// FirebaseStorage Firebase存储实现
type FirebaseStorage struct {
	bucket     *storage.BucketHandle
	bucketName string
	baseURL    string
}

var defaultStorage Storage

// InitStorage 初始化存储服务
func InitStorage(cfg *config.FirebaseConfig) error {
	// 获取已初始化的 Storage 客户端
	storageClient := auth.GetStorageClient()
	if storageClient == nil {
		return fmt.Errorf("firebase storage client not initialized")
	}

	// 获取默认 bucket
	bucket, err := storageClient.Bucket(cfg.StorageBucket)
	if err != nil {
		return fmt.Errorf("failed to get bucket: %v", err)
	}

	// 创建 Firebase Storage 实例
	defaultStorage = &FirebaseStorage{
		bucket:     bucket,
		bucketName: cfg.StorageBucket,
		baseURL:    fmt.Sprintf("https://storage.googleapis.com/%s", cfg.StorageBucket),
	}

	logger.Info("Firebase Storage initialized successfully")
	return nil
}

// GetStorage 获取存储实例
func GetStorage() Storage {
	return defaultStorage
}

// UploadFile 上传文件
func (s *FirebaseStorage) UploadFile(ctx context.Context, file *multipart.FileHeader, directory string) (string, error) {
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	// 读取文件内容
	buffer := make([]byte, file.Size)
	if _, err = src.Read(buffer); err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// 生成文件路径
	filename := generateFileName(file.Filename)
	objectPath := path.Join(directory, filename)

	// 创建对象句柄
	obj := s.bucket.Object(objectPath)

	// 设置文件访问权限为公开
	objectACL := obj.ACL()
	if err := objectACL.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("failed to set file ACL: %v", err)
	}

	// 创建写入器
	writer := obj.NewWriter(ctx)

	// 设置Content-Type
	contentType := getContentType(file.Filename)
	writer.ContentType = contentType

	// 设置缓存控制
	writer.CacheControl = "public, max-age=86400" // 24小时缓存

	// 写入文件内容
	if _, err := io.Copy(writer, bytes.NewReader(buffer)); err != nil {
		return "", fmt.Errorf("failed to copy file to storage: %v", err)
	}

	// 关闭写入器
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	// 返回文件访问URL
	return fmt.Sprintf("%s/%s", s.baseURL, objectPath), nil
}

// DeleteFile 删除文件
func (s *FirebaseStorage) DeleteFile(ctx context.Context, fileURL string) error {
	// 从URL中提取对象路径
	objectPath := strings.TrimPrefix(fileURL, fmt.Sprintf("%s/", s.baseURL))
	if objectPath == fileURL {
		return fmt.Errorf("invalid file URL")
	}

	// 删除对象
	obj := s.bucket.Object(objectPath)
	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return nil // 如果文件不存在，视为删除成功
		}
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// 生成唯一文件名
func generateFileName(originalName string) string {
	ext := path.Ext(originalName)
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d%s", timestamp, ext)
}

// 获取文件Content-Type
func getContentType(filename string) string {
	ext := strings.ToLower(path.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
