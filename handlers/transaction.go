package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"../db"
	"../models"
)

// CreateTransaction handles the creation of a new transaction.
// It calculates the balance based on the previous balance and cash payment.
func CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var transaction models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := db.GetDB()
	if db == nil {
		http.Error(w, "Database connection not available", http.StatusInternalServerError)
		return
	}

	// Generate a new UUID for the transaction
	transaction.ID = uuid.New()
	transaction.CreatedAt = time.Now()
	transaction.UpdatedAt = time.Now()

	// Get the previous balance for the company
	var previousBalance float64
	err := db.QueryRow("SELECT balance FROM transactions WHERE company_id = $1 ORDER BY created_at DESC LIMIT 1", transaction.CompanyID).Scan(&previousBalance)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, fmt.Sprintf("Error retrieving previous balance: %v", err), http.StatusInternalServerError)
		return
	}

	// Calculate the new balance
	transaction.Balance = previousBalance + transaction.Amount - transaction.Cash

	// Insert the new transaction into the database
	_, err = db.Exec(
		"INSERT INTO transactions (id, date, kg, rate, amount, taxi, cash, balance, company_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
		transaction.ID,
		transaction.Date,
		transaction.KG,
		transaction.Rate,
		transaction.Amount,
		transaction.Taxi,
		transaction.Cash,
		transaction.Balance,
		transaction.CompanyID,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error inserting transaction: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

// GetTransactions retrieves all transaction records.
func GetTransactions(w http.ResponseWriter, r *http.Request) {
	db := db.GetDB()
	if db == nil {
		http.Error(w, "Database connection not available", http.StatusInternalServerError)
		return
	}

	rows, err := db.Query("SELECT id, date, kg, rate, amount, taxi, cash, balance, company_id, created_at, updated_at FROM transactions")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.Date,
			&transaction.KG,
			&transaction.Rate,
			&transaction.Amount,
			&transaction.Taxi,
			&transaction.Cash,
			&transaction.Balance,
			&transaction.CompanyID,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transactions)
}

// GetTransaction retrieves a single transaction by ID.
func GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	db := db.GetDB()
	if db == nil {
		http.Error(w, "Database connection not available", http.StatusInternalServerError)
		return
	}

	var transaction models.Transaction
	err = db.QueryRow("SELECT id, date, kg, rate, amount, taxi, cash, balance, company_id, created_at, updated_at FROM transactions WHERE id = $1", id).Scan(
		&transaction.ID,
		&transaction.Date,
		&transaction.KG,
		&transaction.Rate,
		&transaction.Amount,
		&transaction.Taxi,
		&transaction.Cash,
		&transaction.Balance,
		&transaction.CompanyID,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transaction)
}

// UpdateTransaction updates an existing transaction record.
// Note: Updating a transaction might require recalculating balances for subsequent transactions for the same company.
// This implementation only updates the specific transaction and does NOT cascade balance updates.
// A more robust solution would involve recalculating balances for all subsequent transactions.
func UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var updatedTransaction models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&updatedTransaction); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := db.GetDB()
	if db == nil {
		http.Error(w, "Database connection not available", http.StatusInternalServerError)
		return
	}

	// Get the current transaction to get the company_id and previous details
	var currentTransaction models.Transaction
	err = db.QueryRow("SELECT company_id, balance, amount, cash FROM transactions WHERE id = $1", id).Scan(
		&currentTransaction.CompanyID,
		&currentTransaction.Balance,
		&currentTransaction.Amount,
		&currentTransaction.Cash,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedTransaction.UpdatedAt = time.Now()

	// Note: This update does NOT recalculate the balance based on the change.
	// A proper implementation would need to fetch previous transaction balance,
	// calculate the new balance for this transaction, and then update balances
	// for all subsequent transactions for the same company.
	// For simplicity, we are just updating the provided fields.
	_, err = db.Exec(
		"UPDATE transactions SET date = $1, kg = $2, rate = $3, amount = $4, taxi = $5, cash = $6, balance = $7, updated_at = $8 WHERE id = $9",
		updatedTransaction.Date,
		updatedTransaction.KG,
		updatedTransaction.Rate,
		updatedTransaction.Amount,
		updatedTransaction.Taxi,
		updatedTransaction.Cash,
		updatedTransaction.Balance, // Using the balance provided in the update request
		updatedTransaction.UpdatedAt,
		id,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Fetch the updated transaction to return in the response
	var finalTransaction models.Transaction
	err = db.QueryRow("SELECT id, date, kg, rate, amount, taxi, cash, balance, company_id, created_at, updated_at FROM transactions WHERE id = $1", id).Scan(
		&finalTransaction.ID,
		&finalTransaction.Date,
		&finalTransaction.KG,
		&finalTransaction.Rate,
		&finalTransaction.Amount,
		&finalTransaction.Taxi,
		&finalTransaction.Cash,
		&finalTransaction.Balance,
		&finalTransaction.CompanyID,
		&finalTransaction.CreatedAt,
		&finalTransaction.UpdatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(finalTransaction)
}

// DeleteTransaction deletes a transaction record.
// Note: Deleting a transaction might require recalculating balances for subsequent transactions for the same company.
// This implementation only deletes the specific transaction and does NOT cascade balance updates.
// A more robust solution would involve recalculating balances for all subsequent transactions.
func DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	db := db.GetDB()
	if db == nil {
		http.Error(w, "Database connection not available", http.StatusInternalServerError)
		return
	}

	// Note: This delete does NOT recalculate the balance for subsequent transactions.
	// A proper implementation would need to identify subsequent transactions for the
	// same company and recalculate their balances.
	result, err := db.Exec("DELETE FROM transactions WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Transaction deleted successfully")
}