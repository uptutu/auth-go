# Auth Middleware for Gin

## What is it?
一款简单易用的可以为任意结构体生成认证 JWT 并解析 Token 为目标结构体的中间件。

## How to use it?
关于认证模块简单的来说分为以下几个步骤：
1. 认证用户信息
2. 为用户生成对应的认证信息并生成 Token
3. 解析 Token 拿到用户信息
4. 使用用户信息

## Usage Cases
### Generate JWT
```go
import "github.com/uptutu/auth"

    user = User{
	    ID:     "id code", 
	    Name:   "user name", 
	    Avatar: "avatar url",
    }
    token, err = auth.NewSession(user).GenerateAccessToken()
	
    // 设置 Cookie
    // c *gin.Context
    auth.SetCookie(c, token)

```

### Parse Token in Gin Middleware
```go
    // 使用认证中间件
    // NewSession 函数中传入生成 Token 的结构体实例或指针皆可
    // 解析到的用户信息会存在 ctx *gin.Context 中
    r.Use(auth.Identify(auth.NewSession(User{})))

    // 使用认证中间件
    r.GET("/", func(c *gin.Context) {
        data, err := auth.GetAuthenticationDataFrom(c, auth.ValueContextKey)
        if err != nil {
            c.AbortWithStatus(401)
            return
        }
    // 使用自己定义的用户信息断言
        user := data.(User)
        c.JSON(200, gin.H{
            "user": user,
        })

    })
```
建议将从上下文和断言的代码放在一个函数中，可以复用

```go
func GetAuthenticationData(ctx *gin.Context) (*User, error) {
    d, err := auth.GetAuthenticationDataFrom(ctx, auth.ValueContextKey)
    if err != nil {
        return nil, errors.Wrap(err, "get authentication data failed")
    }
    a, ok := d.(User)
    if !ok {
        return nil, errors.Wrap(err, "get authentication data failed")
    }
    return &a, nil
}


func main() {
	...
    r.GET("/", func(c *gin.Context) {
        user, err := GetAuthenticationData(c)
        if err != nil {
            c.AbortWithStatus(401)
            return
        }
        c.JSON(200, gin.H{
            "user": user,
        })
    }
        ...
```
# Advanced
实现自己实现 `auth.Session` 的接口，用一个自己的结构体生成和解析 Token。
```go
type Session interface {
	ValueContextKey() string
	GenerateAccessToken() (string, error)
	ParseToken(ctx context.Context, token string) error
	SetIntoGinCtx(ctx *gin.Context)
	GetFromGinCtx(ctx *gin.Context) (interface{}, error)
}

...
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

func (a asession) SetIntoGinCtx(ctx *gin.Context) {
return
}

func (a asession) GetFromGinCtx(ctx *gin.Context) (interface{}, error) {
return nil, nil
}

```
中间件上的使用流程上也没有改变：
```go
    // 使用认证中间件
    // NewSession 函数中传入生成 Token 的结构体实例或指针皆可
    // 解析到的用户信息会存在 ctx *gin.Context 中
    r.Use(auth.Identify(auth.NewSession(asession{})))

    // 使用认证中间件
    r.GET("/", func(c *gin.Context) {
        data, err := auth.GetAuthenticationDataFrom(c, asession{}.ValueContextKey())
        if err != nil {
            c.AbortWithStatus(401)
            return
        }
    // 使用自己定义的用户信息断言
        user := data.(asession)
        c.JSON(200, gin.H{
            "user": user,
        })

    })
```

# License 
MIT