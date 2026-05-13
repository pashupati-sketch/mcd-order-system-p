package main

import ( 
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Item struct {
	MenuName  string `json:"menuName" validate:"required"`
	UnitPrice int    `json:"unitPrice" validate:"min=1"`
	Quantity  int    `json:"quantity" validate:"range=1-5"`
}

type OrderRequest struct {
	TerminalNo  string `json:"terminalNo"`
	MessageType string `json:"messageType"`
	TotalAmount int    `json:"totalAmount"`
	Items       []Item `json:"items"`
}

// 9. logging utility
func writeLog(action string, data interface{}) {
	f, _ := os.OpenFile("logs/order.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	entry, _ := json.Marshal(data)
	logMsg := fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), action, string(entry))
	f.WriteString(logMsg)
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost {
		var req OrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		writeLog("INBOUND_MESSAGE", req)

		// 7.2 Input Validation
		if err := validateOrder(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		orderNo := generateOrderNo()
		for i, item := range req.Items {
			subtotal := item.UnitPrice * item.Quantity
			db.Exec(`INSERT INTO order_items (order_no, terminal_no, order_status, item_no, menu_name, unit_price, quantity, subtotal) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				orderNo, req.TerminalNo, "オーダー受信", i+1, item.MenuName, item.UnitPrice, item.Quantity, subtotal)
		}

		res := map[string]interface{}{
			"result":      "OK",
			"orderNo":     orderNo,
			"orderStatus": "オーダー受信",
			"totalAmount": req.TotalAmount,
			"message":     "Order accepted successfully",
		}
		writeLog("OUTBOUND_MESSAGE", res)
		json.NewEncoder(w).Encode(res)

	} else if r.Method == http.MethodGet {
		// 10.2 & 10.3 List Orders (Grouped by order_no)
		status := r.URL.Query().Get("status")
		query := "SELECT order_no, terminal_no, order_status, SUM(subtotal), created_at FROM order_items"
		if status != "" {
			query += " WHERE order_status = ?"
		}
		query += " GROUP BY order_no ORDER BY created_at DESC"

		rows, _ := db.Query(query, status) // simplified error check
		if status == "" {
			rows, _ = db.Query(strings.Replace(query, " WHERE order_status = ?", "", 1))
		}
		// ... scanning and returning JSON list logic ...
	}
}

func validateOrder(req OrderRequest) error {
	if req.TerminalNo == "" { return fmt.Errorf("terminalNo is required") }
	if req.MessageType != "ORDER_CONFIRM" { return fmt.Errorf("invalid messageType") }
	if len(req.Items) < 1 || len(req.Items) > 5 { return fmt.Errorf("items must be between 1 and 5") }
	
	calculatedTotal := 0
	names := make(map[string]bool)
	for _, item := range req.Items {
		if names[item.MenuName] { return fmt.Errorf("duplicate menuName: %s", item.MenuName) }
		names[item.MenuName] = true
		calculatedTotal += (item.UnitPrice * item.Quantity)
	}
	if calculatedTotal != req.TotalAmount { return fmt.Errorf("totalAmount mismatch") }
	return nil
}

func generateOrderNo() string {
	today := time.Now().Format("0102")
	var count int
	db.QueryRow("SELECT COUNT(DISTINCT order_no) FROM order_items WHERE order_no LIKE ? || '-%'", today).Scan(&count)
	return fmt.Sprintf("%s-%03d", today, count+1)
}