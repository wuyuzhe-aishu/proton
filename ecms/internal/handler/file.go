package handler

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/proton/ecms/internal/response"
)

type FileHandler struct{}

func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

func (h *FileHandler) Register(r *gin.RouterGroup) {
	g := r.Group("/files/*path", h.setFilePath)
	g.POST("", h.post)
	g.GET("", h.get)
	g.PUT("", h.put)
	g.DELETE("", h.delete)
	g.HEAD("", h.head)
}

func (h *FileHandler) post(c *gin.Context) {
	filePath := h.getFilePath(c)

	if before, ok := strings.CutSuffix(filePath, "/movement"); ok {
		h.handleMovement(c, before)
		return
	}

	contentType := c.GetHeader("Content-Type")
	if contentType == "application/json" {
		var body map[string]any
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, 400, 400000000, "invalid request", nil)
			return
		}

		fileType, ok := body["type"].(string)
		if !ok || fileType == "" {
			response.Error(c, 400, 400000000, "required field type is missing", nil)
			return
		}

		if fileType != "directory" {
			response.Error(c, 400, 400000000, "field type should be directory", nil)
			return
		}

		if _, err := os.Stat(filePath); err == nil {
			return
		}

		if err := os.MkdirAll(filePath, 0755); err != nil {
			response.Error(c, 500, 500000000, "failed to create directory", err.Error())
			return
		}
		return
	}

	parent := filepath.Dir(filePath)
	if parent != "." {
		if _, err := os.Stat(parent); os.IsNotExist(err) {
			response.Error(c, 400, 400000000, "parent is not found", gin.H{"path": parent})
			return
		}
		if info, err := os.Stat(parent); err == nil && !info.IsDir() {
			response.Error(c, 400, 400000000, "parent is not a directory", gin.H{"path": parent})
			return
		}
		if !isWritable(parent) {
			response.Error(c, 400, 400000000, "parent is not writable", gin.H{"path": parent})
			return
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		response.Error(c, 500, 500000000, "failed to create file", err.Error())
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, c.Request.Body); err != nil {
		response.Error(c, 500, 500000000, "failed to write file", err.Error())
		return
	}

	c.Status(200)
}

func (h *FileHandler) get(c *gin.Context) {
	filePath := h.getFilePath(c)

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		response.Error(c, 404, 404000000, "not found", gin.H{"path": filePath})
		return
	}
	if err != nil {
		response.Error(c, 500, 500000000, "server error", err.Error())
		return
	}

	h.setStatHeaders(c, info)

	if info.IsDir() {
		entries, err := os.ReadDir(filePath)
		if err != nil {
			response.Error(c, 500, 500000000, "failed to read directory", err.Error())
			return
		}

		results := make([]gin.H, 0)
		for _, entry := range entries {
			stat, _ := entry.Info()
			results = append(results, gin.H{
				"name": entry.Name(),
				"mode": stat.Mode().String(),
				"size": stat.Size(),
			})
		}
		response.Success(c, results)
		return
	}

	mimeType := mime.TypeByExtension(filepath.Ext(filePath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	c.Header("Content-Type", mimeType)
	c.File(filePath)
}

func (h *FileHandler) put(c *gin.Context) {
	filePath := h.getFilePath(c)

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		response.Error(c, 404, 404000000, "not found", gin.H{"path": filePath})
		return
	}
	if err != nil {
		response.Error(c, 500, 500000000, "server error", err.Error())
		return
	}

	_ = info

	if !isWritable(filePath) {
		response.Error(c, 400, 400000000, "target is not writable", gin.H{"path": filePath})
		return
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		response.Error(c, 500, 500000000, "failed to open file", err.Error())
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, c.Request.Body); err != nil {
		response.Error(c, 500, 500000000, "failed to write file", err.Error())
		return
	}

	c.Status(200)
}

func (h *FileHandler) delete(c *gin.Context) {
	filePath := h.getFilePath(c)

	info, err := os.Lstat(filePath)
	if os.IsNotExist(err) {
		response.Error(c, 404, 404000000, "not found", gin.H{"path": filePath})
		return
	}
	if err != nil {
		response.Error(c, 500, 500000000, "server error", err.Error())
		return
	}

	parent := filepath.Dir(filePath)
	if parent == "." {
		parent = "."
	}
	if !isWritable(parent) {
		response.Error(c, 400, 400000000, "parent is not writable", gin.H{"parent": parent})
		return
	}

	if info.IsDir() {
		if err := os.RemoveAll(filePath); err != nil {
			response.Error(c, 500, 500000000, "failed to remove directory", err.Error())
			return
		}
	} else if info.Mode()&os.ModeSymlink != 0 {
		if err := os.Remove(filePath); err != nil {
			response.Error(c, 500, 500000000, "failed to remove file", err.Error())
			return
		}
	} else {
		if err := os.Remove(filePath); err != nil {
			response.Error(c, 500, 500000000, "failed to remove file", err.Error())
			return
		}
	}

	c.Status(200)
}

func (h *FileHandler) head(c *gin.Context) {
	filePath := h.getFilePath(c)

	if !exists(filePath) {
		response.Error(c, 404, 404000000, "not found", gin.H{"path": filePath})
		return
	}

	follow := c.DefaultQuery("follow", "true")
	var info os.FileInfo
	var err error

	if follow == "true" {
		info, err = os.Stat(filePath)
	} else {
		info, err = os.Lstat(filePath)
	}
	if err != nil {
		response.Error(c, 500, 500000000, "server error", err.Error())
		return
	}

	h.setStatHeaders(c, info)

	if !info.IsDir() {
		mimeType := mime.TypeByExtension(filepath.Ext(filePath))
		if mimeType != "" {
			c.Header("Content-Type", mimeType)
		}
	}

	c.Status(200)
}

func (h *FileHandler) handleMovement(c *gin.Context, filePath string) {
	if _, err := os.Lstat(filePath); os.IsNotExist(err) {
		response.Error(c, 404, 404000000, "not found", gin.H{"path": filePath})
		return
	}

	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, 400, 400000000, "invalid request", nil)
		return
	}

	dest, ok := body["destination"]
	if !ok || dest == "" {
		response.Error(c, 404, 404000000, "invalid request", gin.H{"detail": body})
		return
	}

	if err := os.Rename(filePath, dest); err != nil {
		response.Error(c, 500, 500000000, "failed to move file", err.Error())
		return
	}

	c.Status(200)
}

func (h *FileHandler) setStatHeaders(c *gin.Context, info os.FileInfo) {
	var mode uint32 = uint32(info.Mode())
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		mode = stat.Mode
		c.Header("x-st-uid", strconv.FormatUint(uint64(stat.Uid), 10))
		c.Header("x-st-gid", strconv.FormatUint(uint64(stat.Gid), 10))
	}
	c.Header("x-st-mode", fmt.Sprintf("%#o", mode))
	c.Header("x-st-size", strconv.FormatInt(info.Size(), 10))
}

const contextKeyFilePath = "filePath"

func (h *FileHandler) setFilePath(c *gin.Context) {
	p, err := url.PathUnescape(strings.TrimPrefix(c.Param("path"), "/"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, 400000000, fmt.Sprintf("invalid file path: %v", err), nil)
		return
	}
	p = "/" + p
	log.Println("file path:", p)
	c.Set(contextKeyFilePath, p)
}
func (h *FileHandler) getFilePath(c *gin.Context) string {
	p, ok := c.Get(contextKeyFilePath)
	if ok && p != nil {
		return p.(string)
	}
	return ""
}

func isWritable(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return info.Mode().Perm()&0200 != 0
	}
	return info.Mode().Perm()&0200 != 0
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
