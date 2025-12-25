// Package storage 提供文件存储和管理功能，支持过期时间
package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	// ErrFileNotFound 文件不存在错误
	ErrFileNotFound = errors.New("file not found")
	// ErrFileExpired 文件已过期错误
	ErrFileExpired = errors.New("file has expired")
	// ErrInvalidFileID 无效的文件ID错误
	ErrInvalidFileID = errors.New("invalid file ID")
)

// FileMetadata 文件元数据
type FileMetadata struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	UploadTime  time.Time `json:"upload_time"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Storage 文件存储管理器
type Storage struct {
	baseDir      string
	defaultTTL   time.Duration
	metadataDir  string
	filesDir     string
	mu           sync.RWMutex
	cleanupTicker *time.Ticker
	stopCleanup  chan struct{}
}

// NewStorage 创建文件存储管理器
func NewStorage(baseDir string, defaultTTL time.Duration) (*Storage, error) {
	if baseDir == "" {
		baseDir = "./storage"
	}

	metadataDir := filepath.Join(baseDir, "metadata")
	filesDir := filepath.Join(baseDir, "files")

	// 创建目录
	for _, dir := range []string{baseDir, metadataDir, filesDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	s := &Storage{
		baseDir:     baseDir,
		defaultTTL:  defaultTTL,
		metadataDir: metadataDir,
		filesDir:    filesDir,
		stopCleanup: make(chan struct{}),
	}

	// 启动过期文件清理协程
	s.startCleanup()

	return s, nil
}

// Save 保存文件
// file: 上传的文件
// ttl: 过期时间，0 表示使用默认值
func (s *Storage) Save(file multipart.File, header *multipart.FileHeader, ttl time.Duration) (*FileMetadata, error) {
	// 生成唯一文件ID
	id, err := generateID()
	if err != nil {
		return nil, err
	}

	// 设置过期时间
	if ttl <= 0 {
		ttl = s.defaultTTL
	}
	expiresAt := time.Now().Add(ttl)

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 保存文件到磁盘
	filePath := s.getFilePath(id)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// 创建元数据
	metadata := &FileMetadata{
		ID:          id,
		Name:        header.Filename,
		Size:        header.Size,
		ContentType: header.Header.Get("Content-Type"),
		UploadTime:  time.Now(),
		ExpiresAt:   expiresAt,
	}

	// 保存元数据
	if err := s.saveMetadata(metadata); err != nil {
		// 删除已保存的文件
		os.Remove(filePath)
		return nil, err
	}

	return metadata, nil
}

// SaveFromReader 从 Reader 保存文件
func (s *Storage) SaveFromReader(r io.Reader, filename, contentType string, size int64, ttl time.Duration) (*FileMetadata, error) {
	// 生成唯一文件ID
	id, err := generateID()
	if err != nil {
		return nil, err
	}

	// 设置过期时间
	if ttl <= 0 {
		ttl = s.defaultTTL
	}
	expiresAt := time.Now().Add(ttl)

	// 读取文件内容
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 保存文件到磁盘
	filePath := s.getFilePath(id)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// 创建元数据
	metadata := &FileMetadata{
		ID:          id,
		Name:        filename,
		Size:        int64(len(content)),
		ContentType: contentType,
		UploadTime:  time.Now(),
		ExpiresAt:   expiresAt,
	}

	// 保存元数据
	if err := s.saveMetadata(metadata); err != nil {
		// 删除已保存的文件
		os.Remove(filePath)
		return nil, err
	}

	return metadata, nil
}

// Get 获取文件
func (s *Storage) Get(id string) (*FileMetadata, []byte, error) {
	metadata, err := s.getMetadata(id)
	if err != nil {
		return nil, nil, err
	}

	// 检查是否过期
	if time.Now().After(metadata.ExpiresAt) {
		// 异步删除过期文件
		go s.Delete(id)
		return nil, nil, ErrFileExpired
	}

	// 读取文件内容
	filePath := s.getFilePath(id)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	return metadata, content, nil
}

// GetMetadata 获取文件元数据
func (s *Storage) GetMetadata(id string) (*FileMetadata, error) {
	metadata, err := s.getMetadata(id)
	if err != nil {
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(metadata.ExpiresAt) {
		return nil, ErrFileExpired
	}

	return metadata, nil
}

// Delete 删除文件
func (s *Storage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	metadataPath := s.getMetadataPath(id)
	filePath := s.getFilePath(id)

	// 删除元数据
	os.Remove(metadataPath)

	// 删除文件
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists 检查文件是否存在且未过期
func (s *Storage) Exists(id string) bool {
	metadata, err := s.GetMetadata(id)
	if err != nil {
		return false
	}
	return metadata != nil
}

// startCleanup 启动过期文件清理
func (s *Storage) startCleanup() {
	s.cleanupTicker = time.NewTicker(5 * time.Minute)

	go func() {
		for {
			select {
			case <-s.cleanupTicker.C:
				s.cleanupExpired()
			case <-s.stopCleanup:
				s.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanupCleanup 清理过期文件
func (s *Storage) cleanupExpired() {
	entries, err := os.ReadDir(s.metadataDir)
	if err != nil {
		return
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		metadataPath := filepath.Join(s.metadataDir, entry.Name())
		data, err := os.ReadFile(metadataPath)
		if err != nil {
			continue
		}

		// 简单的解析（生产环境应使用 JSON 解析）
		fileID := entry.Name()
		metadata, err := s.parseMetadata(data)
		if err != nil {
			continue
		}

		if now.After(metadata.ExpiresAt) {
			s.Delete(fileID)
		}
	}
}

// Stop 停止存储管理器
func (s *Storage) Stop() {
	close(s.stopCleanup)
}

// saveMetadata 保存元数据
func (s *Storage) saveMetadata(metadata *FileMetadata) error {
	metadataPath := s.getMetadataPath(metadata.ID)

	// 使用 JSON 格式保存
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

// getMetadata 获取文件元数据
func (s *Storage) getMetadata(id string) (*FileMetadata, error) {
	metadataPath := s.getMetadataPath(id)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	return s.parseMetadata(data)
}

// parseMetadata 解析元数据
func (s *Storage) parseMetadata(data []byte) (*FileMetadata, error) {
	var metadata FileMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

// getFilePath 获取文件路径
func (s *Storage) getFilePath(id string) string {
	return filepath.Join(s.filesDir, id)
}

// getMetadataPath 获取元数据路径
func (s *Storage) getMetadataPath(id string) string {
	return filepath.Join(s.metadataDir, id+".json")
}

// generateID 生成唯一ID
func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
