package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestModel używany do testów operacji zbiorczych
type TestBulkModel struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func TestBulkCreate(t *testing.T) {
	// Tworzenie bazy danych w pamięci dla testów
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:bulk_create_test_%d?mode=memory&cache=shared", time.Now().UnixNano())), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&TestBulkModel{})
	require.NoError(t, err)

	// Tworzenie repozytorium
	repo := NewGenericRepository(db, &TestBulkModel{})

	// Test case: Poprawne tworzenie wielu rekordów
	t.Run("Successful bulk create", func(t *testing.T) {
		// Dane do utworzenia
		items := []TestBulkModel{
			{Name: "User 1", Email: "user1@example.com", Age: 25},
			{Name: "User 2", Email: "user2@example.com", Age: 30},
			{Name: "User 3", Email: "user3@example.com", Age: 35},
		}

		// Wywołanie metody BulkCreate
		err := repo.BulkCreate(context.Background(), items)

		// Sprawdzenie rezultatu
		require.NoError(t, err)

		// Sprawdzenie czy rekordy istnieją w bazie danych
		var count int64
		err = db.Model(&TestBulkModel{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)

		// Sprawdzenie czy ID zostały przypisane
		var users []TestBulkModel
		err = db.Find(&users).Error
		require.NoError(t, err)

		for i, user := range users {
			assert.NotZero(t, user.ID)
			assert.Equal(t, items[i].Name, user.Name)
			assert.Equal(t, items[i].Email, user.Email)
			assert.Equal(t, items[i].Age, user.Age)
		}
	})

	// Test case: Pusty slice
	t.Run("Empty slice", func(t *testing.T) {
		// GORM zwraca błąd "empty slice found" przy próbie utworzenia pustej tablicy,
		// więc sprawdzamy, czy funkcja zwraca ten błąd, ale nie traktujemy tego jako błąd testu
		err := repo.BulkCreate(context.Background(), []TestBulkModel{})

		// Błąd jest oczekiwany
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty slice found")
	})

	// Test case: Błąd walidacji danych
	t.Run("Validation error", func(t *testing.T) {
		// Test z niepoprawnym typem danych
		assert.NotPanics(t, func() {
			repo.BulkCreate(context.Background(), map[string]interface{}{"invalid": "data"})
		})
	})
}

func TestBulkUpdate(t *testing.T) {
	// Tworzenie bazy danych w pamięci dla testów
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:bulk_update_test_%d?mode=memory&cache=shared", time.Now().UnixNano())), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&TestBulkModel{})
	require.NoError(t, err)

	// Tworzenie repozytorium
	repo := NewGenericRepository(db, &TestBulkModel{})

	// Przygotowanie danych testowych
	testUsers := []TestBulkModel{
		{Name: "User 1", Email: "user1@example.com", Age: 25},
		{Name: "User 2", Email: "user2@example.com", Age: 30},
		{Name: "User 3", Email: "user3@example.com", Age: 35},
	}

	result := db.Create(&testUsers)
	require.NoError(t, result.Error)
	require.Equal(t, int64(3), result.RowsAffected)

	// Test case: Poprawna aktualizacja wielu rekordów
	t.Run("Successful bulk update", func(t *testing.T) {
		// Przygotowanie warunków i aktualizacji
		condition := map[string]interface{}{"name": "User 1"}
		updates := map[string]interface{}{"age": 26, "email": "user1.updated@example.com"}

		// Wywołanie metody BulkUpdate
		err := repo.BulkUpdate(context.Background(), condition, updates)

		// Sprawdzenie rezultatu
		require.NoError(t, err)

		// Sprawdzenie czy dane w bazie są zaktualizowane
		var updatedUser TestBulkModel
		err = db.Where("name = ?", "User 1").First(&updatedUser).Error
		require.NoError(t, err)
		assert.Equal(t, 26, updatedUser.Age)
		assert.Equal(t, "user1.updated@example.com", updatedUser.Email)
	})

	// Test case: Aktualizacja wielu rekordów
	t.Run("Update multiple records", func(t *testing.T) {
		// Przygotowanie warunków i aktualizacji dla wielu rekordów
		condition := map[string]interface{}{"age": 30}
		updates := map[string]interface{}{"age": 31, "email": "updated@example.com"}

		// Wywołanie metody BulkUpdate
		err := repo.BulkUpdate(context.Background(), condition, updates)

		// Sprawdzenie rezultatu
		require.NoError(t, err)

		// Sprawdzenie czy dane w bazie są zaktualizowane
		var updatedUsers []TestBulkModel
		err = db.Where("age = ?", 31).Find(&updatedUsers).Error
		require.NoError(t, err)
		assert.Equal(t, 1, len(updatedUsers))
		assert.Equal(t, "updated@example.com", updatedUsers[0].Email)
	})

	// Test case: Brak pasujących rekordów
	t.Run("No matching records", func(t *testing.T) {
		condition := map[string]interface{}{"name": "Non-existent user"}
		updates := map[string]interface{}{"age": 99}

		err := repo.BulkUpdate(context.Background(), condition, updates)

		// Powinien przejść, ale nie powinien zaktualizować żadnego rekordu
		require.NoError(t, err)

		// Sprawdzenie czy żaden rekord nie został zaktualizowany z wiekiem 99
		var count int64
		err = db.Model(&TestBulkModel{}).Where("age = ?", 99).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
