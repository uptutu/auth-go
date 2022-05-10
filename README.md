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
    r.Use(auth.Identify(User{}))

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
## 自定义认证结构体
实现自己实现 `auth.Session` 的接口，用一个自己的结构体生成和解析 Token。
```go
type Session interface {
    GenerateAccessToken() (string, error)
    ParseToken(ctx context.Context, token string) error
    IsParsed() bool
    SetExtIntoGinContext(ctx *gin.Context) error
    GetExtFromGinContext(ctx *gin.Context) (interface{}, error)
    GetToken(ctx *gin.Context) (string, error)
}

...

type customSession struct{}

func ( customSession )GenerateAccessToken() ( string, error) {
//TODO implement me
panic("your own custom token generator")
}

func ( customSession )ParseToken(ctx context.Context, token string) error {
//TODO implement me
panic("your own custom token parser")
}

func ( customSession )IsParsed() bool {
//TODO implement me
panic("implement me")
}

func ( customSession )SetExtIntoGinContext(ctx *gin.Context) error {
//TODO implement me
panic("implement me")
}

func ( customSession )GetExtFromGinContext(ctx *gin.Context) ( interface{}, error) {
//TODO implement me
panic("implement me")
}

func ( customSession )GetToken(ctx *gin.Context) ( string, error) {
//TODO implement me
panic("your own custom token generator")
}



```
中间件上的使用流程上也没有改变：
```go
    // 使用认证中间件
    // NewSession 函数中传入生成 Token 的结构体实例或指针皆可
    // 解析到的用户信息会存在 ctx *gin.Context 中
    r.Use(auth.Identify(customSession{}))

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
## 通过中间件层认证请求
假设现在有一个接口是获取订单信息，检验该请求的订单 ID 是否为该用户的订单
```go
func IsUserOrder(c *gin.Context, sess Session){
	orderID := c.Request.Query("order_id")
    authInfo, ok := sess.(User)
	if !ok {
		c.JSON(http.StatusInternalServerError, nil)
		c.Abrode()
		return
    }
	found := db.Select("1").From("orders").Where("user_id = ?", authInfo.ID).Where("id = ?", orderID).Exists()
    if !found {
        c.JSON(http.StatusBadRequest, )
        c.Abrode()
        return
    }
}

...

r.Use(auth.Identify(auth.NewSession(User{}), IsUserOrder))
r.GET("/order/id", func(c *gin.Context) {
	...
}
```

# License 
MIT