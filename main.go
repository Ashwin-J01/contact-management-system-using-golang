package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
)

type Contact struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

type ContactManager struct {
	Contacts []Contact
	FilePath string
}

func (cm *ContactManager) Load() {
	file, err := os.ReadFile(cm.FilePath)
	if err == nil {
		json.Unmarshal(file, &cm.Contacts)
	}
}

func (cm *ContactManager) Save() {
	data, err := json.MarshalIndent(cm.Contacts, "", "  ")
	if err != nil {
		fmt.Println("Failed to save:", err)
		return
	}
	err = os.WriteFile(cm.FilePath, data, 0644)
	if err != nil {
		fmt.Println("Failed to write file:", err)
	}
}

func (cm *ContactManager) AddContact(name, phone, email, address string) {
	cm.Contacts = append(cm.Contacts, Contact{Name: name, Phone: phone, Email: email, Address: address})
	cm.Save()
}

func (cm *ContactManager) SearchContact(name string) *Contact {
	for _, c := range cm.Contacts {
		if strings.EqualFold(c.Name, name) {
			return &c
		}
	}
	return nil
}

func (cm *ContactManager) UpdateContact(oldName, newName, phone, email, address string) bool {
	for i, c := range cm.Contacts {
		if strings.EqualFold(c.Name, oldName) {
			cm.Contacts[i] = Contact{
				Name:    newName,
				Phone:   phone,
				Email:   email,
				Address: address,
			}
			cm.Save()
			return true
		}
	}
	return false
}

func (cm *ContactManager) DeleteContact(name string) bool {
	for i, c := range cm.Contacts {
		if strings.EqualFold(c.Name, name) {
			cm.Contacts = append(cm.Contacts[:i], cm.Contacts[i+1:]...)
			cm.Save()
			return true
		}
	}
	return false
}

var manager = ContactManager{FilePath: "contacts.json"}

func main() {
	manager.Load()

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/view", handleView)
	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/delete", handleDelete)
	http.HandleFunc("/update", handleUpdate)

	fmt.Println("Server starting on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		name := r.FormValue("name")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		address := r.FormValue("address")

		manager.AddContact(name, phone, email, address)
		http.Redirect(w, r, "/view", http.StatusSeeOther)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/add.html"))
	tmpl.Execute(w, nil)
}

func handleView(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/view.html"))
	tmpl.Execute(w, manager.Contacts)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		name := r.FormValue("name")
		contact := manager.SearchContact(name)
		tmpl := template.Must(template.ParseFiles("templates/search.html"))
		tmpl.Execute(w, contact)
		return
	}
	tmpl := template.Must(template.ParseFiles("templates/search.html"))
	tmpl.Execute(w, nil)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		name := r.FormValue("name")
		if manager.DeleteContact(name) {
			http.Redirect(w, r, "/view", http.StatusSeeOther)
			return
		}
	}
	tmpl := template.Must(template.ParseFiles("templates/delete.html"))
	tmpl.Execute(w, nil)
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		oldName := r.FormValue("oldName")
		newName := r.FormValue("name")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		address := r.FormValue("address")

		if manager.UpdateContact(oldName, newName, phone, email, address) {
			http.Redirect(w, r, "/view", http.StatusSeeOther)
			return
		}
		// If update failed, show error
		tmpl := template.Must(template.ParseFiles("templates/update.html"))
		tmpl.Execute(w, map[string]interface{}{
			"Error":   "Contact not found",
			"Name":    newName,
			"Phone":   phone,
			"Email":   email,
			"Address": address,
		})
		return
	}

	// GET request - show update form
	name := r.URL.Query().Get("name")
	contact := manager.SearchContact(name)
	if contact == nil {
		http.Redirect(w, r, "/view", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/update.html"))
	data := map[string]interface{}{
		"Name":    contact.Name,
		"Phone":   contact.Phone,
		"Email":   contact.Email,
		"Address": contact.Address,
	}
	tmpl.Execute(w, data)
}
