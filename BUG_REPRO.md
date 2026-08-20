# BUG_REPRO：查询不存在用户/无效 token 触发 nil 解引用 panic

## Bug 是什么

- `internal/service/user_service.go`：`Me` 调用 `FindByID` 后用 `_` 丢掉 error，用户不存在时返回 `(nil, nil)`；`UpdateProfile` 继续拿 nil 解引用字段，直接 panic。
- `internal/middleware/auth.go`：`RequireAuth` 调用 `ParseToken` 与 `FindByID` 后同样丢掉 error，token 无效或用户被删除时 `claims`/`user` 为 nil，随后 `claims.UserID` / `user.ID` 解引用 panic。
- 同一反模式：忽略错误后解引用必然为 nil 的指针。

## 如何触发

1. 查询不存在的用户资料（`GET /users/me` 对已删除用户）；
2. 更新不存在的用户资料（`PUT /users/me`）；
3. 携带无效 token 访问受保护接口；
4. 用户被删除后携带旧 token 访问。

## 真实错误信息

`go test` 实测输出（关键段）：

```
panic: runtime error: invalid memory address or nil pointer dereference [recovered, repanicked]

goroutine 34 [running]:
	internal/service/user_service.go:36 +0x44
	internal/service/user_nil_p301_test.go:28 +0x358
```
