package auth

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"testing"

	"github.com/stretchr/testify/assert"
)

type c_type []byte

type User struct {
	ID   string            `json:"id"`
	Name string            `json:"name"`
	Age  int               `json:"age"`
	Xx   c_type            `json:"xx"`
	Yy   []int             `json:"yy"`
	Zz   map[string]string `json:"zz"`
}

type asession struct {
}

func (a asession) CustomerFunc() string {
	return "customer"
}

func (a asession) ValueContextKey() string {
	return "ok"
}

func (a asession) GenerateAccessToken() (string, error) {
	return "ok", nil
}

func (a asession) ParseToken(ctx context.Context, token string) error {
	return nil
}

func (a asession) SetIntoGinCtx(ctx *gin.Context) error {
	return nil
}

func (a asession) GetFromGinCtx(ctx *gin.Context) (interface{}, error) {
	return nil, nil
}

func (a asession) IsParsed() bool {
	return true
}

func (a asession) GetToken(ctx *gin.Context) (string, error) {
	return "ok", nil
}

func TestNewSession(t *testing.T) {
	s := NewSession(asession{})
	assert.NotNil(t, s)
	as, ok := s.(asession)
	assert.True(t, ok)
	assert.Equal(t, "customer", as.CustomerFunc())

	s = NewSession(&asession{})
	assert.NotNil(t, s)
	asPtr, ok := s.(*asession)
	assert.True(t, ok)
	assert.Equal(t, "customer", asPtr.CustomerFunc())
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

	s = NewSession(u2)
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
		Xx:   c_type{'a', 'b'},
		Yy:   []int{1, 2},
		Zz:   map[string]string{"test": "ok"},
	}
	s := NewSession(u)
	token, err := s.GenerateAccessToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	b, _ := json.Marshal(u)
	var um map[string]interface{}
	json.Unmarshal(b, &um)

	tests := []struct {
		input interface{}
		want  interface{}
	}{
		{nil, um},
		{User{}, u},
		{&User{}, u},
		{&u, u},
	}

	for _, test := range tests {
		s = NewSession(test.input)
		err = s.ParseToken(context.Background(), token)
		assert.NoError(t, err)
		ds, ok := s.(*DefaultSession)
		if ok {
			assert.NotNil(t, ds.Ext)
		}
		assert.Equal(t, test.want, ds.Ext)
	}
}
