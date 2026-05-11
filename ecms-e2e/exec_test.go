package e2e

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Exec API", func() {
	execURL := func() string {
		return baseURL + "/api/ecms/v1alpha1/exec"
	}

	doExec := func(params url.Values, body io.Reader) (*http.Response, []byte) {
		u, _ := url.Parse(execURL())
		u.RawQuery = params.Encode()
		req, err := http.NewRequest(http.MethodPost, u.String(), body)
		Expect(err).NotTo(HaveOccurred())
		req.Header.Set("Authorization", authHeader)
		resp, err := httpClient.Do(req)
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())
		return resp, data
	}

	exitCode := func(resp *http.Response) int {
		code := resp.Header.Get("x-exit-code")
		if code == "" {
			return -1
		}
		i, err := strconv.Atoi(code)
		Expect(err).NotTo(HaveOccurred())
		return i
	}

	Describe("POST /api/ecms/v1alpha1/exec", func() {

		It("1.1 应成功执行简单命令并返回输出和 exit code 0", func() {
			params := url.Values{"command": {"echo"}, "stdout": {"true"}}
			params.Add("command", "hello")
			resp, data := doExec(params, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(exitCode(resp)).To(Equal(0))
			Expect(string(data)).To(Equal("hello\n"))
		})

		It("1.2 应支持多条命令参数", func() {
			params := url.Values{"command": {"ls"}, "stdout": {"true"}}
			params.Add("command", "-la")
			params.Add("command", "/tmp")
			resp, data := doExec(params, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(exitCode(resp)).To(Equal(0))
			Expect(string(data)).To(ContainSubstring("total"))
		})

		It("1.3 命令不存在应返回 404", func() {
			params := url.Values{"command": {"nonexistent_cmd_xyz"}, "stderr": {"true"}}
			resp, _ := doExec(params, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("1.4 命令非零退出应返回对应 exit code", func() {
			params := url.Values{"command": {"sh"}, "stdout": {"true"}}
			params.Add("command", "-c")
			params.Add("command", "exit 42")
			resp, _ := doExec(params, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(exitCode(resp)).To(Equal(42))
		})

		It("1.5 缺少 command 参数应返回 500", func() {
			params := url.Values{}
			resp, _ := doExec(params, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
		})

		It("1.6 stdin 传递应使 cat 输出输入内容", func() {
			params := url.Values{"command": {"cat"}, "stdin": {"true"}, "stdout": {"true"}}
			input := "hello stdin world"
			resp, data := doExec(params, strReader(input))
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(exitCode(resp)).To(Equal(0))
			Expect(string(data)).To(Equal(input))
		})

		It("1.7 stderr 捕获应返回错误输出", func() {
			params := url.Values{"command": {"ls"}, "stderr": {"true"}, "stdout": {"false"}}
			params.Add("command", "/nonexistent_path_xyz")
			resp, data := doExec(params, nil)
			Expect(exitCode(resp)).NotTo(Equal(0))
			Expect(string(data)).To(ContainSubstring("No such file or directory"))
		})

		It("1.8 stdout=false 不应返回输出", func() {
			params := url.Values{"command": {"echo"}, "stdout": {"false"}}
			params.Add("command", "hello")
			resp, data := doExec(params, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(exitCode(resp)).To(Equal(0))
			Expect(string(data)).To(BeEmpty())
		})

		It("1.9 应正确处理带特殊字符的命令", func() {
			params := url.Values{"command": {"echo"}, "stdout": {"true"}}
			params.Add("command", "hello 'world' \"foo\" &")
			resp, data := doExec(params, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(exitCode(resp)).To(Equal(0))
			Expect(string(data)).To(Equal("hello 'world' \"foo\" &\n"))
		})
	})
})

type strReader string

func (s strReader) Read(p []byte) (int, error) {
	return copy(p, []byte(s)), io.EOF
}

func (s strReader) Close() error {
	return nil
}

var _ = fmt.Sprintf
