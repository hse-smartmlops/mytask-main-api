package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	models "emplacc-api/internal/domain"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	return db
}

/*func TestCreateUserWithId_50SequentialUsers(t *testing.T) {
	db := setupTestDB(t)

	const total = 50

	start := time.Now()
	defer func() {
		t.Logf("✅ Created %d users in %v", total, time.Since(start))
	}()

	for i := 1; i <= total; i++ {
		userID := uuid.New()
		email := fmt.Sprintf("user%d@example.com", i)
		firstName := fmt.Sprintf("FirstName%d", i)
		lastName := fmt.Sprintf("LastName%d", i)
		isActive := true

		req := request.UserCreateRequest{
			Email:         &email,
			FirstName:     &firstName,
			LastName:      &lastName,
			IsActive:      &isActive,
			EmailVerified: &isActive, // чередуем verified/unverified
		}

		t.Run(fmt.Sprintf("create user #%d", i), func(t *testing.T) {
			// 👇 Вызываем функцию с передачей db
			err := controller.CreateUserWithId(db, req, userID)
			require.NoError(t, err, "Failed to create user #%d with ID %s", i, userID)

			// Проверяем, что пользователь действительно создан
			var user models.User
			result := db.First(&user, "id = ?", userID)
			require.NoError(t, result.Error, "User #%d not found in DB after creation", i)

			assert.Equal(t, email, user.Email)
			assert.Equal(t, firstName, user.FirstName)
			assert.Equal(t, lastName, user.LastName)
			assert.True(t, user.IsActive)
			assert.Equal(t, i%2 == 0, user.EmailVerified)
		})
	}

	// Финальная проверка: убедимся, что в базе ровно 50 записей
	var count int64
	db.Model(&models.User{}).Count(&count)
	assert.Equal(t, int64(total), count, "Expected %d users in DB, got %d", total, count)
}*/