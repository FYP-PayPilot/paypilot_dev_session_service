package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_TableName(t *testing.T) {
	u := User{}
	assert.Equal(t, "users", u.TableName())
}

func TestUser_Structure(t *testing.T) {
	u := User{
		Email:     "test@example.com",
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
		IsActive:  true,
	}

	assert.Equal(t, "test@example.com", u.Email)
	assert.Equal(t, "testuser", u.Username)
	assert.Equal(t, "Test", u.FirstName)
	assert.Equal(t, "User", u.LastName)
	assert.True(t, u.IsActive)
}
