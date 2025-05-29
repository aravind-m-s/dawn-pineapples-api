package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aravind-m-s/dawn-pineapples-api/db"
	"github.com/aravind-m-s/dawn-pineapples-api/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func CreateCompany(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var company models.Company
	_ = json.NewDecoder(r.Body).Decode(&company)

	// Generate a new UUID for the company
	company.ID = uuid.New()

	sqlStatement := `INSERT INTO companies (id, name, image_url) VALUES ($1, $2, $3) RETURNING id`
	err := db.GetDB().QueryRow(sqlStatement, company.ID, company.Name, company.Image).Scan(&company.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(company)
}

func GetCompanies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var companies []models.Company

	sqlStatement := `SELECT id, name, image_url FROM companies`
	rows, err := db.GetDB().Query(sqlStatement)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var company models.Company
		err = rows.Scan(&company.ID, &company.Name, &company.Image)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		companies = append(companies, company)
	}

	json.NewEncoder(w).Encode(companies)
}

func GetCompany(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]

	companyID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	var company models.Company
	sqlStatement := `SELECT id, name, image_url FROM companies WHERE id = $1`
	row := db.GetDB().QueryRow(sqlStatement, companyID)

	err = row.Scan(&company.ID, &company.Name, &company.Image)
	switch {
	case err == sql.ErrNoRows:
		http.Error(w, "Company not found", http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(company)
}

func UpdateCompany(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]

	companyID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	var company models.Company
	_ = json.NewDecoder(r.Body).Decode(&company)
	company.ID = companyID // Ensure the ID from the URL is used

	sqlStatement := `UPDATE companies SET name = $2, image_url = $3 WHERE id = $1`
	res, err := db.GetDB().Exec(sqlStatement, company.ID, company.Name, company.Image)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if count == 0 {
		http.Error(w, "Company not found", http.StatusNotFound)
		return
	}

	// Fetch the updated company to return it in the response
	updatedCompany := models.Company{}
	sqlStatement = `SELECT id, name, image_url FROM companies WHERE id = $1`
	row := db.GetDB().QueryRow(sqlStatement, companyID)
	err = row.Scan(&updatedCompany.ID, &updatedCompany.Name, &updatedCompany.Image)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedCompany)
}

func DeleteCompany(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]

	companyID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid company ID", http.StatusBadRequest)
		return
	}

	sqlStatement := `DELETE FROM companies WHERE id = $1`
	res, err := db.GetDB().Exec(sqlStatement, companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if count == 0 {
		http.Error(w, "Company not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": fmt.Sprintf("Company with ID %s deleted successfully", id)})
}
