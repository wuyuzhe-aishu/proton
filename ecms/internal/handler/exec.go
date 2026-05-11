package handler

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ExecHandler struct{}

func NewExecHandler() *ExecHandler {
	return &ExecHandler{}
}

func (h *ExecHandler) Register(r *gin.RouterGroup) {
	r.POST("/exec", h.post)
}

func (h *ExecHandler) post(c *gin.Context) {
	commands := c.QueryArray("command")
	if len(commands) == 0 {
		c.JSON(500, gin.H{
			"status":  500000000,
			"message": "command is required",
		})
		return
	}

	stdin := c.DefaultQuery("stdin", "false") == "true"
	stdout := c.DefaultQuery("stdout", "true") == "true"
	stderr := c.DefaultQuery("stderr", "false") == "true"

	cmd := exec.Command(commands[0], commands[1:]...)

	if stdin {
		cmd.Stdin = c.Request.Body
	}

	var buf *bytes.Buffer
	if stdout || stderr {
		buf = new(bytes.Buffer)
	}

	if stdout {
		cmd.Stdout = buf
	}

	if stderr {
		cmd.Stderr = buf
	}

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			if errors.Is(err, exec.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"status":  404000000,
					"message": "command not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  500000000,
				"message": err.Error(),
			})
			return
		}
	}

	if cmd.ProcessState != nil {
		c.Header("x-exit-code", strconv.Itoa(cmd.ProcessState.ExitCode()))
	}
	if buf != nil {
		if _, err := io.Copy(c.Writer, buf); err != nil {
			log.Printf("WARNING: write http response body fail: %s", err)
		}
	}
}
