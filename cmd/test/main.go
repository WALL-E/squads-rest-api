package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
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

func main() {
	// Step 0: Health Check
	step := "Step 0: Health Check"
	resp := httpRequest("GET", baseURL+"/health", nil)
	printStep(step, resp, "GET", baseURL+"/health")
	if !resp.Success {
		fmt.Println("⚠️ Warning: Health check failed, but continuing with tests...")
	}

	// Step 1: Create Multisig
	step = "Step 1: Create Multisig"
	m1 := map[string]interface{}{
		"multisig_address": "multisig_test_1",
		"name":             "Test Multisig",
		"description":      "testing multisig",
		"logo":             "https://example.com/logo1.png",
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
		"vault_address": "vault_test_1",
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
		"member_address": "member_test_1",
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
		"spend_address": "spend_test_1",
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

	// Step 6: Get Transaction
	step = "Step 6: Get Transaction"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs/%s/transactions/%d", baseURL, multisigID, transactionIndexNumber), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs/%s/transactions/%d", baseURL, multisigID, transactionIndexNumber))

	// Step 7: List Multisigs
	step = "Step 7: List Multisigs"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs?sort=%s", baseURL, url.QueryEscape("name")), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs?sort=%s", baseURL, url.QueryEscape("name")))

	// Step 8: Get Multisig
	step = "Step 8: Get Multisig"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs/%s", baseURL, multisigID), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs/%s", baseURL, multisigID))

	// Step 9: Update Multisig
	step = "Step 9: Update Multisig"
	m1Update := map[string]interface{}{
		"name":        "Updated Test Multisig",
		"description": "updated testing multisig",
		"logo":        "https://example.com/updated-logo.png",
	}
	resp = httpRequest("PUT", fmt.Sprintf("%s/multisigs/%s", baseURL, multisigID), m1Update)
	printStep(step, resp, "PUT", fmt.Sprintf("%s/multisigs/%s", baseURL, multisigID))

	// Step 10: Update Transaction
	step = "Step 10: Update Transaction"
	txUpdate := map[string]interface{}{
		"signature": "6WgZXMkVDHyjvVkARYIYHWb5ulOxkjvGqpfcIlde7yKw0BB2M4dWZRpLEsSgVFvT3hhOEuISopCNyt2pQlvCk0U0",
	}
	resp = httpRequest("PUT", fmt.Sprintf("%s/multisigs/%s/transactions/%d", baseURL, multisigID, transactionIndexNumber), txUpdate)
	printStep(step, resp, "PUT", fmt.Sprintf("%s/multisigs/%s/transactions/%d", baseURL, multisigID, transactionIndexNumber))

	// Step 11: List Vaults
	step = "Step 11: List Vaults"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs/%s/vaults?sort=%s", baseURL, multisigID, url.QueryEscape("name asc")), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs/%s/vaults?sort=%s", baseURL, multisigID, url.QueryEscape("name asc")))

	// Step 12: List Members
	step = "Step 12: List Members"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs/%s/members?sort=%s", baseURL, multisigID, url.QueryEscape("name desc")), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs/%s/members?sort=%s", baseURL, multisigID, url.QueryEscape("name desc")))

	// Step 13: List Spends
	step = "Step 13: List Spends"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs/%s/spends?sort=%s", baseURL, multisigID, url.QueryEscape("name asc")), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs/%s/spends?sort=%s", baseURL, multisigID, url.QueryEscape("name asc")))

	// Step 14: List Transactions
	step = "Step 14: List Transactions"
	resp = httpRequest("GET", fmt.Sprintf("%s/multisigs/%s/transactions?sort=%s", baseURL, multisigID, url.QueryEscape("index_number asc")), nil)
	printStep(step, resp, "GET", fmt.Sprintf("%s/multisigs/%s/transactions?sort=%s", baseURL, multisigID, url.QueryEscape("index_number asc")))

	// Step 15: Delete Transaction
	step = "Step 15: Delete Transaction"
	resp = httpRequest("DELETE", fmt.Sprintf("%s/multisigs/%s/transactions/%d", baseURL, multisigID, transactionIndexNumber), nil)
	printStep(step, resp, "DELETE", fmt.Sprintf("%s/multisigs/%s/transactions/%d", baseURL, multisigID, transactionIndexNumber))

	// Step 16: Delete Spend
	step = "Step 16: Delete Spend"
	resp = httpRequest("DELETE", fmt.Sprintf("%s/multisigs/%s/spends/%s", baseURL, multisigID, spendAddress), nil)
	printStep(step, resp, "DELETE", fmt.Sprintf("%s/multisigs/%s/spends/%s", baseURL, multisigID, spendAddress))

	// Step 17: Delete Member
	step = "Step 17: Delete Member"
	resp = httpRequest("DELETE", fmt.Sprintf("%s/multisigs/%s/members/%s", baseURL, multisigID, memberAddress), nil)
	printStep(step, resp, "DELETE", fmt.Sprintf("%s/multisigs/%s/members/%s", baseURL, multisigID, memberAddress))

	// Step 18: Delete Vault
	step = "Step 18: Delete Vault"
	resp = httpRequest("DELETE", fmt.Sprintf("%s/multisigs/%s/vaults/%s", baseURL, multisigID, vaultAddress), nil)
	printStep(step, resp, "DELETE", fmt.Sprintf("%s/multisigs/%s/vaults/%s", baseURL, multisigID, vaultAddress))

	// Step 19: Delete Multisig
	step = "Step 19: Delete Multisig"
	resp = httpRequest("DELETE", fmt.Sprintf("%s/multisigs/%s", baseURL, multisigID), nil)
	printStep(step, resp, "DELETE", fmt.Sprintf("%s/multisigs/%s", baseURL, multisigID))

	// Generate HTML report
	generateReport()
}

func generateReport() {
	html := "<html><head><title>API Test Report</title></head><body>"
	html += "<h1>API Test Report (19 steps)</h1><table border='1' cellpadding='5'><tr><th>Step</th><th>Method</th><th>URL</th><th>Success</th><th>Response</th></tr>"
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
