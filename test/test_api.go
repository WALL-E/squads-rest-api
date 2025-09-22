package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type HealthResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
	Version   string `json:"version"`
	BinName   string `json:"bin_name"`
}

type StepResult struct {
	Step     string
	Method   string
	URL      string
	Success  bool
	Response string
}

var (
	baseURL = "http://localhost:8090"
	report  []StepResult
)

func httpRequest(method, urlStr string, body interface{}) Response {
	var reqBody io.Reader
	if body != nil {
		bs, _ := json.Marshal(body)
		reqBody = bytes.NewReader(bs)
	}

	req, _ := http.NewRequest(method, urlStr, reqBody)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Response{Success: false, Message: err.Error()}
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var r Response
	if err := json.Unmarshal(respBytes, &r); err != nil {
		return Response{Success: false, Message: string(respBytes)}
	}
	return r
}

func printStep(step string, resp Response, method, url string) {
	color := "\033[32m" // green
	if !resp.Success {
		color = "\033[31m" // red
	}
	fmt.Printf("%s[%s] %s\nRequest: %s %s\nResponse: %+v\033[0m\n\n",
		color, step, color, method, url, resp)

	respStr, _ := json.MarshalIndent(resp, "", "  ")
	report = append(report, StepResult{
		Step: step, Method: method, URL: url, Success: resp.Success, Response: string(respStr),
	})
}

func healthCheck() bool {
	req, _ := http.NewRequest("GET", baseURL+"/health", nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var healthResp HealthResponse
	if err := json.Unmarshal(respBytes, &healthResp); err != nil {
		fmt.Printf("Failed to parse health response: %v\n", err)
		return false
	}

	fmt.Printf("Health Check Response:\n")
	fmt.Printf("  Success: %v\n", healthResp.Success)
	fmt.Printf("  Message: %s\n", healthResp.Message)
	fmt.Printf("  Build Time: %s\n", healthResp.BuildTime)
	fmt.Printf("  Git Commit: %s\n", healthResp.GitCommit)
	fmt.Printf("  Version: %s\n", healthResp.Version)
	fmt.Printf("  Bin Name: %s\n", healthResp.BinName)

	// Validate bin_name field
	if healthResp.BinName == "" {
		fmt.Printf("❌ FAIL: bin_name field is empty\n")
		return false
	}
	if healthResp.BinName != "squads-rest-api" {
		fmt.Printf("❌ FAIL: bin_name field expected 'squads-rest-api', got '%s'\n", healthResp.BinName)
		return false
	}
	
	fmt.Printf("✅ PASS: bin_name field validation successful\n")
	return healthResp.Success
}

func testAPI() {
	var resp Response
	
	// Step 0: Health Check
	step := "Step 0: Health Check"
	fmt.Printf("\n=== %s ===\n", step)
	if !healthCheck() {
		fmt.Println("❌ Health check failed, but continuing with tests")
	} else {
		fmt.Println("✅ Health check passed")
	}

	// Step 1: Create Multisig
	step = "Step 1: Create Multisig"
	timestamp := time.Now().Unix()
	m1 := map[string]interface{}{
		"multisig_address": fmt.Sprintf("multisig_test_%d", timestamp),
		"name":             "Test Multisig",
		"description":      "testing multisig",
		"logo":             "https://example.com/test-logo.png",
	}
	resp = httpRequest("POST", baseURL+"/multisigs", m1)
	printStep(step, resp, "POST", baseURL+"/multisigs")
	if !resp.Success {
		fmt.Println("Cannot continue, Step 1 failed")
		return
	}
	multisigID := resp.Data.(map[string]interface{})["multisig_address"].(string)

	// Step 2: Create Vault
	step = "Step 2: Create Vault"
	v1 := map[string]interface{}{
		"vault_address": fmt.Sprintf("vault_test_%d", timestamp),
		"name":          "Test Vault 1",
		"description":   "vault testing",
	}
	resp = httpRequest("POST", fmt.Sprintf("%s/multisigs/%s/vaults", baseURL, multisigID), v1)
	printStep(step, resp, "POST", fmt.Sprintf("%s/multisigs/%s/vaults", baseURL, multisigID))
	vaultAddress := ""
	if resp.Success {
		vaultAddress = v1["vault_address"].(string)
	}

	// Step 3: Create Member
	step = "Step 3: Create Member"
	mem1 := map[string]interface{}{
		"member_address": fmt.Sprintf("member_test_%d", timestamp),
		"name":           "Test Member 1",
		"description":    "member testing",
	}
	resp = httpRequest("POST", fmt.Sprintf("%s/multisigs/%s/members", baseURL, multisigID), mem1)
	printStep(step, resp, "POST", fmt.Sprintf("%s/multisigs/%s/members", baseURL, multisigID))
	memberAddress := ""
	if resp.Success {
		memberAddress = mem1["member_address"].(string)
	}

	// Step 4: Create Spend
	step = "Step 4: Create Spend"
	sp1 := map[string]interface{}{
		"spend_address": fmt.Sprintf("spend_test_%d", timestamp),
		"name":          "Test Spend 1",
		"description":   "spend testing",
	}
	resp = httpRequest("POST", fmt.Sprintf("%s/multisigs/%s/spends", baseURL, multisigID), sp1)
	printStep(step, resp, "POST", fmt.Sprintf("%s/multisigs/%s/spends", baseURL, multisigID))
	spendAddress := ""
	if resp.Success {
		spendAddress = sp1["spend_address"].(string)
	}

	// Step 5: Create Transaction
	step = "Step 5: Create Transaction"
	tx1 := map[string]interface{}{
		"indexNumber": 1,
		"signature":   "5VfYXLjUCHxjuUjzRXHXGVa4tkNwjiuFqoebHkcd6xJv9AA1L3cVYQoKDrSgUEuS2ggNDtHRnpBMxs1oPkuBjzTz",
	}
	resp = httpRequest("POST", fmt.Sprintf("%s/multisigs/%s/transactions", baseURL, multisigID), tx1)
	printStep(step, resp, "POST", fmt.Sprintf("%s/multisigs/%s/transactions", baseURL, multisigID))
	transactionIndexNumber := 0
	if resp.Success {
		if data, ok := resp.Data.(map[string]interface{}); ok {
			if indexNum, exists := data["indexNumber"]; exists {
				transactionIndexNumber = int(indexNum.(float64))
			}
		}
	}

	// Step 6: List Multisigs
	step = "Step 6: List Multisigs"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs?sort=%s", baseURL, url.QueryEscape("name")), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs?sort=%s", baseURL, url.QueryEscape("name")))

	// Step 7: Get Multisig
	step = "Step 7: Get Multisig"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs/%s", baseURL, multisigID), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs/%s", baseURL, multisigID))

	// Step 8: List Vaults
	step = "Step 8: List Vaults"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs/%s/vaults?sort=%s", baseURL, multisigID, url.QueryEscape("name asc")), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs/%s/vaults?sort=%s", baseURL, multisigID, url.QueryEscape("name asc")))

	// Step 9: List Members
	step = "Step 9: List Members"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs/%s/members?sort=%s", baseURL, multisigID, url.QueryEscape("name desc")), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs/%s/members?sort=%s", baseURL, multisigID, url.QueryEscape("name desc")))

	// Step 10: List Spends
	step = "Step 10: List Spends"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs/%s/spends?sort=%s", baseURL, multisigID, url.QueryEscape("name asc")), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs/%s/spends?sort=%s", baseURL, multisigID, url.QueryEscape("name asc")))

	// Step 11: List Transactions
	step = "Step 11: List Transactions"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs/%s/transactions?sort=%s", baseURL, multisigID, url.QueryEscape("index_number asc")), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs/%s/transactions?sort=%s", baseURL, multisigID, url.QueryEscape("index_number asc")))

	// Step 12: Delete Transaction
	step = "Step 12: Delete Transaction"
	resp = httpRequest("DELETE", fmt.Sprintf("%s/multisigs/%s/transactions/%d", baseURL, multisigID, transactionIndexNumber), nil)
	printStep(step, resp, "DELETE", fmt.Sprintf("%s/multisigs/%s/transactions/%d", baseURL, multisigID, transactionIndexNumber))

	// Step 13: Delete Spend
	step = "Step 13: Delete Spend"
	resp = httpRequest("DELETE", fmt.Sprintf("%s/multisigs/%s/spends/%s", baseURL, multisigID, spendAddress), nil)
	printStep(step, resp, "DELETE", fmt.Sprintf("%s/multisigs/%s/spends/%s", baseURL, multisigID, spendAddress))

	// Step 14: Delete Member
	step = "Step 14: Delete Member"
	resp = httpRequest("DELETE", fmt.Sprintf("%s/multisigs/%s/members/%s", baseURL, multisigID, memberAddress), nil)
	printStep(step, resp, "DELETE", fmt.Sprintf("%s/multisigs/%s/members/%s", baseURL, multisigID, memberAddress))

	// Step 15: Delete Vault
	step = "Step 15: Delete Vault"
	resp = httpRequest("DELETE", fmt.Sprintf("%s/multisigs/%s/vaults/%s", baseURL, multisigID, vaultAddress), nil)
	printStep(step, resp, "DELETE", fmt.Sprintf("%s/multisigs/%s/vaults/%s", baseURL, multisigID, vaultAddress))

	// Step 16: Delete Multisig
	step = "Step 16: Delete Multisig"
	resp = httpRequest("DELETE", fmt.Sprintf("%s/multisigs/%s", baseURL, multisigID), nil)
	printStep(step, resp, "DELETE", fmt.Sprintf("%s/multisigs/%s", baseURL, multisigID))

	// Generate HTML report
	generateReport()
}

func generateReport() {
	html := "<html><head><title>API Test Report</title></head><body>"
	html += "<h1>API Test Report</h1><table border='1' cellpadding='5'><tr><th>Step</th><th>Method</th><th>URL</th><th>Success</th><th>Response</th></tr>"
	for _, r := range report {
		color := "green"
		if !r.Success {
			color = "red"
		}
		html += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td style='color:%s'>%v</td><td><pre>%s</pre></td></tr>",
			r.Step, r.Method, r.URL, color, r.Success, r.Response)
	}
	html += "</table></body></html>"

	os.WriteFile("report.html", []byte(html), 0644)
	fmt.Println("✅ Report saved to report.html")
}

func main() {
	testAPI()
}
