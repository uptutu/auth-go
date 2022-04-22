package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestGenerate(t *testing.T) {
	u := User{
		ID:   "id",
		Name: "name",
		Age:  11,
	}
	s := NewSession(u)
	token, err := s.GenerateAccessToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	u2 := User{
		ID:   "id2",
		Name: "name2",
		Age:  22,
	}
	s.Ext = u2
	token2, err := s.GenerateAccessToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token2)

	assert.NotEqual(t, token, token2)
}

func TestParse(t *testing.T) {
	u := User{
		ID:   "new_id",
		Name: "new_name",
		Age:  33,
	}
	s := NewSession(u)
	token, err := s.GenerateAccessToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	s = NewSession(nil)
	err = s.ParseToken(context.Background(), token)
	assert.NoError(t, err)
	assert.NotNil(t, s.Ext)

	s = NewSession(User{})
	err = s.ParseToken(context.Background(), token)
	assert.NoError(t, err)
	if myUser, ok := s.Ext.(User); ok {
		assert.Equal(t, u.ID, myUser.ID)
	}

	s = NewSession(&User{})
	err = s.ParseToken(context.Background(), token)
	assert.NoError(t, err)
	if us, ok := s.Ext.(User); ok {
		assert.Equal(t, u.ID, us.ID)
	}

	s = NewSession(&u)
	token, err = s.GenerateAccessToken()
	assert.NoError(t, err)
	s = NewSession(&User{})
	err = s.ParseToken(context.Background(), token)
	assert.NoError(t, err)
	if us, ok := s.Ext.(User); ok {
		assert.Equal(t, u.ID, us.ID)
	}
}
