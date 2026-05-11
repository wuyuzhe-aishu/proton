package e2e

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var baseURL string
var httpClient *http.Client
var authHeader string

func TestECMSE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ECMS E2E Suite")
}

func generateToken(username string, epoch int64) string {
	seed := fmt.Sprintf("%s:%d", username, epoch)
	hash := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("%x", hash)
}

func makeAuthHeader(username string) string {
	now := time.Now().Unix()
	epochMin := now - now%60
	token := generateToken(username, epochMin)
	credentials := fmt.Sprintf("%s:%s", username, token)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return "Basic " + encoded
}

var _ = BeforeSuite(func() {
	log.SetOutput(GinkgoWriter)

	baseURL = os.Getenv("ECMS_SERVER")
	if baseURL == "" {
		baseURL = "http://fedora.hyper-v.local:9202"
	}

	username := os.Getenv("ECMS_USERNAME")
	if username == "" {
		username = "admin"
	}
	authHeader = makeAuthHeader(username)

	httpClient = &http.Client{}

	req, err := http.NewRequest("POST", baseURL+"/api/ecms/v1alpha1/exec", nil)
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Authorization", authHeader)
	resp, err := httpClient.Do(req)
	Expect(err).NotTo(HaveOccurred(), "ECMS server must be reachable")
	defer resp.Body.Close()
	fmt.Fprintf(GinkgoWriter, "ECMS server responded with status %d\n", resp.StatusCode)

	fmt.Fprintf(GinkgoWriter, "Connected to ECMS server at %s (user: %s)\n", baseURL, username)
})
