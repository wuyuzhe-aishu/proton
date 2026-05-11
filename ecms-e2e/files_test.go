package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Files API", func() {
	var testRoot string

	filesURL := func(p string) string {
		return baseURL + "/api/ecms/v1alpha1/files" + p
	}

	doRequest := func(method, p string, contentType string, body io.Reader) *http.Response {
		req, err := http.NewRequest(method, filesURL(p), body)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Set("Authorization", authHeader)
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		resp, err := httpClient.Do(req)
		Expect(err).NotTo(HaveOccurred())
		return resp
	}

	readBody := func(resp *http.Response) []byte {
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		return data
	}

	createFile := func(p, content string) {
		resp := doRequest("POST", p, "application/octet-stream", strings.NewReader(content))
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(200), "create file: %s", p)
	}

	createDirViaAPI := func(p string) {
		body, _ := json.Marshal(map[string]string{"type": "directory"})
		resp := doRequest("POST", p, "application/json", bytes.NewReader(body))
		defer resp.Body.Close()
		Expect(resp.StatusCode).To(Equal(200), "create dir: %s", p)
	}

	deletePath := func(p string) int {
		resp := doRequest("DELETE", p, "", nil)
		defer resp.Body.Close()
		return resp.StatusCode
	}

	statPath := func(p string) *http.Response {
		return doRequest("HEAD", p, "", nil)
	}

	readFile := func(p string) (int, []byte) {
		resp := doRequest("GET", p, "", nil)
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, data
	}

	BeforeEach(func() {
		testRoot = fmt.Sprintf("/tmp/ecms-e2e/%d", GinkgoRandomSeed())
		body, _ := json.Marshal(map[string]string{"type": "directory"})
		resp := doRequest("POST", testRoot, "application/json", bytes.NewReader(body))
		defer resp.Body.Close()
	})

	AfterEach(func() {
		deletePath(testRoot)
	})

	// =============================================
	// POST - 创建
	// =============================================
	Describe("POST - create", func() {

		It("2.1 应创建普通文件并返回 200", func() {
			p := testRoot + "/test_create_file.txt"
			content := "hello world"
			resp := doRequest("POST", p, "application/octet-stream", strings.NewReader(content))
			Expect(resp.StatusCode).To(Equal(200))
			resp.Body.Close()

			code, data := readFile(p)
			Expect(code).To(Equal(200))
			Expect(string(data)).To(Equal(content))

			deletePath(p)
		})

		It("2.2 应创建目录并返回 200", func() {
			p := testRoot + "/test_create_dir"
			body, _ := json.Marshal(map[string]string{"type": "directory"})
			resp := doRequest("POST", p, "application/json", bytes.NewReader(body))
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(200))

			code, _ := readFile(p)
			Expect(code).To(Equal(200))

			deletePath(p)
		})

		It("2.3 目录已存在应不报错", func() {
			p := testRoot + "/test_dup_dir"
			createDirViaAPI(p)

			body, _ := json.Marshal(map[string]string{"type": "directory"})
			resp := doRequest("POST", p, "application/json", bytes.NewReader(body))
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(200))

			deletePath(p)
		})

		It("2.4 父目录不存在应返回 400", func() {
			p := testRoot + "/nonexistent_parent/file.txt"
			resp := doRequest("POST", p, "application/octet-stream", strings.NewReader("data"))
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(400))
			data := readBody(resp)
			Expect(string(data)).To(ContainSubstring("parent is not found"))
		})

		It("2.6 type 字段不是 directory 应返回 400", func() {
			p := testRoot + "/test_bad_type"
			body, _ := json.Marshal(map[string]string{"type": "file"})
			resp := doRequest("POST", p, "application/json", bytes.NewReader(body))
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(400))
		})

		It("2.7 缺少 type 字段应返回 400", func() {
			p := testRoot + "/test_no_type"
			body, _ := json.Marshal(map[string]string{})
			resp := doRequest("POST", p, "application/json", bytes.NewReader(body))
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(400))
		})
	})

	// =============================================
	// GET - 读取
	// =============================================
	Describe("GET - read", func() {
		var testFile string

		BeforeEach(func() {
			testFile = testRoot + "/test_read_file.txt"
			createFile(testFile, "read content")
		})

		AfterEach(func() {
			deletePath(testFile)
		})

		It("3.1 应读取文件内容", func() {
			code, data := readFile(testFile)
			Expect(code).To(Equal(200))
			Expect(string(data)).To(Equal("read content"))
		})

		It("3.2 应列出目录", func() {
			code, data := readFile(testRoot)
			Expect(code).To(Equal(200))
			var entries []map[string]interface{}
			err := json.Unmarshal(data, &entries)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).NotTo(BeEmpty())
			Expect(entries[0]).To(HaveKey("name"))
			Expect(entries[0]).To(HaveKey("mode"))
			Expect(entries[0]).To(HaveKey("size"))
		})

		It("3.3 文件不存在应返回 404", func() {
			code, data := readFile(testRoot + "/nonexistent_file_xyz")
			Expect(code).To(Equal(404))
			Expect(string(data)).To(ContainSubstring("not found"))
		})

		It("3.4 空目录应返回空数组", func() {
			emptyDir := testRoot + "/empty_dir"
			createDirViaAPI(emptyDir)
			defer deletePath(emptyDir)

			code, data := readFile(emptyDir)
			Expect(code).To(Equal(200))
			Expect(string(data)).To(MatchJSON("[]"))
		})
	})

	// =============================================
	// PUT - 更新
	// =============================================
	Describe("PUT - update", func() {
		var testFile string

		BeforeEach(func() {
			testFile = testRoot + "/test_update_file.txt"
			createFile(testFile, "original content")
		})

		AfterEach(func() {
			deletePath(testFile)
		})

		It("4.1 应覆盖文件内容", func() {
			resp := doRequest("PUT", testFile, "application/octet-stream", strings.NewReader("updated content"))
			Expect(resp.StatusCode).To(Equal(200))
			resp.Body.Close()

			code, data := readFile(testFile)
			Expect(code).To(Equal(200))
			Expect(string(data)).To(Equal("updated content"))
		})

		It("4.2 文件不存在应返回 404", func() {
			p := testRoot + "/nonexistent_update.txt"
			resp := doRequest("PUT", p, "application/octet-stream", strings.NewReader("data"))
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(404))
		})
	})

	// =============================================
	// DELETE - 删除
	// =============================================
	Describe("DELETE - delete", func() {

		It("5.1 应删除文件", func() {
			p := testRoot + "/test_delete_file.txt"
			createFile(p, "to be deleted")

			code := deletePath(p)
			Expect(code).To(Equal(200))

			readCode, _ := readFile(p)
			Expect(readCode).To(Equal(404))
		})

		It("5.2 应删除目录", func() {
			p := testRoot + "/test_delete_dir"
			createDirViaAPI(p)

			code := deletePath(p)
			Expect(code).To(Equal(200))
		})

		It("5.3 删除不存在的路径应返回 404", func() {
			code := deletePath(testRoot + "/nonexistent_delete_xyz")
			Expect(code).To(Equal(404))
		})
	})

	// =============================================
	// HEAD - 元数据
	// =============================================
	Describe("HEAD - stat", func() {
		var testFile string

		BeforeEach(func() {
			testFile = testRoot + "/test_stat_file.txt"
			createFile(testFile, "stat content")
		})

		AfterEach(func() {
			deletePath(testFile)
		})

		It("6.1 应获取文件元数据 headers", func() {
			resp := statPath(testFile)
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(200))
			Expect(resp.Header.Get("x-st-mode")).NotTo(BeEmpty())
			Expect(resp.Header.Get("x-st-size")).NotTo(BeEmpty())
			Expect(resp.Header.Get("x-st-uid")).NotTo(BeEmpty())
			Expect(resp.Header.Get("x-st-gid")).NotTo(BeEmpty())
		})

		It("6.2 应获取目录元数据", func() {
			resp := statPath(testRoot)
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(200))
			Expect(resp.Header.Get("x-st-mode")).NotTo(BeEmpty())
			Expect(resp.Header.Get("x-st-size")).NotTo(BeEmpty())
		})

		It("6.3 路径不存在应返回 404", func() {
			resp := statPath(testRoot + "/nonexistent_stat_xyz")
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(404))
		})

		It("6.4 follow=false 应返回 symlink 自身元数据", func() {
			linkPath := testRoot + "/test_symlink"
			targetPath := testRoot + "/test_stat_file.txt"

			cmdURL := baseURL + "/api/ecms/v1alpha1/exec"
			params := url.Values{"command": {"ln"}, "stdout": {"false"}, "stderr": {"true"}}
			params.Add("command", "-sf")
			params.Add("command", targetPath)
			params.Add("command", linkPath)
			u, _ := url.Parse(cmdURL)
			u.RawQuery = params.Encode()
			req, _ := http.NewRequest("POST", u.String(), nil)
			req.Header.Set("Authorization", authHeader)
			resp, err := httpClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			resp.Body.Close()
			Expect(resp.Header.Get("x-exit-code")).To(Or(Equal("0"), Equal("")))

			q := url.Values{"follow": {"false"}}
			u2, _ := url.Parse(filesURL(linkPath))
			u2.RawQuery = q.Encode()
			req2, _ := http.NewRequest("HEAD", u2.String(), nil)
			req2.Header.Set("Authorization", authHeader)
			resp2, err := httpClient.Do(req2)
			Expect(err).NotTo(HaveOccurred())
			defer resp2.Body.Close()
			Expect(resp2.StatusCode).To(Equal(200))
			mode := resp2.Header.Get("x-st-mode")
			Expect(mode).To(MatchRegexp("(^0o)?012"))

			deletePath(linkPath)
		})
	})

	// =============================================
	// Movement - 重命名
	// =============================================
	Describe("Movement - rename", func() {
		var src, dst string

		BeforeEach(func() {
			src = testRoot + "/test_move_src.txt"
			dst = testRoot + "/test_move_dst.txt"
			createFile(src, "move content")
		})

		AfterEach(func() {
			deletePath(src)
			deletePath(dst)
		})

		It("7.1 应重命名文件", func() {
			body, _ := json.Marshal(map[string]string{"destination": dst})
			resp := doRequest("POST", src+"/movement", "application/json", bytes.NewReader(body))
			Expect(resp.StatusCode).To(Equal(200))
			resp.Body.Close()

			code, _ := readFile(src)
			Expect(code).To(Equal(404))

			code, data := readFile(dst)
			Expect(code).To(Equal(200))
			Expect(string(data)).To(Equal("move content"))
		})

		It("7.2 源路径不存在应返回 404", func() {
		body, _ := json.Marshal(map[string]string{"destination": dst})
			resp := doRequest("POST", testRoot+"/nonexistent_src/movement", "application/json", bytes.NewReader(body))
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(404))
		})

		It("7.3 缺少 destination 字段应返回 404", func() {
			body, _ := json.Marshal(map[string]string{})
			resp := doRequest("POST", src+"/movement", "application/json", bytes.NewReader(body))
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(404))
		})
	})
})


