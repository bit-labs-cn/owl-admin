# 实战示例：新增 `notice` 模块

本篇给出一套**可直接照着改的完整示例**：在 `owl-admin` 里新增一个 `notice` 模块。它的复杂度刻意控制在单表 CRUD，适合作为 AI 和开发者真正开工时的模板。

目标：

- 后端新增公告管理模块
- 接入 `model / repository / service / handle / route / binds / migrate`
- 在左侧菜单中展示
- 使用现有权限、操作日志与统一响应

## 设计目标

`notice` 模块作为 `System` 下的一个普通管理模块，约定：

- 表名：`admin_notice`
- 路由前缀：`/api/v1/notices`
- 模块英文名：`notice`
- 模块中文名：`公告管理`
- 菜单路径：`/system/notice/index`
- 访问级别：全部使用 `router.AccessAuthorized`

## 需要修改的文件

```text
app/
├── model/
│   └── notice.go                 # 新增
├── repository/
│   └── notice.go                 # 新增
├── service/
│   └── notice_service.go         # 新增
├── handle/v1/
│   └── notice_handle.go          # 新增
├── app.go                        # 修改 Binds()
├── route/api.go                  # 修改 InitApi() / InitMenu()
└── database/auto_migrate_gen.go  # 修改 Migrate()
```

## 1. `app/model/notice.go`

```go
package model

import "bit-labs.cn/owl/provider/db"

const (
	NoticeStatusEnabled  = 1
	NoticeStatusDisabled = 2
)

type Notice struct {
	db.BaseModel
	Title   string `gorm:"comment:公告标题" json:"title"`
	Content string `gorm:"comment:公告内容" json:"content"`
	Status  int    `gorm:"comment:状态(1启用,2禁用)" json:"status"`
}

func (Notice) TableName() string {
	return "admin_notice"
}
```

说明：

- 沿用项目现有风格，嵌入 `db.BaseModel`
- 状态值使用常量，不直接写魔法数字

## 2. `app/repository/notice.go`

```go
package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

type NoticeRepositoryInterface interface {
	Save(notice *model.Notice) error
	Detail(id any) (*model.Notice, error)
	Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.Notice, err error)
	contract.WithContext[NoticeRepositoryInterface]
}

type NoticeRepository struct {
	db  *gorm.DB
	ctx context.Context
	db.BaseRepository[model.Notice]
}

func NewNoticeRepository(d *gorm.DB) NoticeRepositoryInterface {
	return &NoticeRepository{
		db:             d,
		BaseRepository: db.NewBaseRepository[model.Notice](d),
	}
}

func (r *NoticeRepository) WithContext(ctx context.Context) NoticeRepositoryInterface {
	r.db = r.db.WithContext(ctx)
	r.ctx = ctx
	return r
}

func (r *NoticeRepository) Save(notice *model.Notice) error {
	return r.BaseRepository.Save(notice)
}

func (r *NoticeRepository) Detail(id any) (*model.Notice, error) {
	return r.BaseRepository.Detail(id)
}

func (r *NoticeRepository) Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.Notice, err error) {
	return r.BaseRepository.Retrieve(page, pageSize, fn)
}
```

## 3. `app/service/notice_service.go`

```go
package service

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type CreateNoticeReq struct {
	Title   string `json:"title" validate:"required,min=2,max=100" label:"公告标题"`
	Content string `json:"content" validate:"required,min=2,max=5000" label:"公告内容"`
	Status  int    `json:"status" validate:"required,oneof=1 2" label:"状态"`
}

type UpdateNoticeReq struct {
	ID uint `json:"id,string" validate:"required,gt=0" label:"公告ID"`
	CreateNoticeReq
}

type RetrieveNoticeReq struct {
	router.PageReq
	TitleLike string `json:"title" form:"title" validate:"omitempty,max=100" label:"公告标题"`
	Status    int    `json:"status" form:"status" validate:"omitempty,oneof=1 2" label:"状态"`
}

type NoticeService struct {
	db.BaseRepository[model.Notice]
	repo     repository.NoticeRepositoryInterface
	locker   redis.LockerFactory
	validate *validatorv10.Validate
}

func NewNoticeService(
	repo repository.NoticeRepositoryInterface,
	tx *gorm.DB,
	locker redis.LockerFactory,
	validate *validatorv10.Validate,
) *NoticeService {
	return &NoticeService{
		BaseRepository: db.NewBaseRepository[model.Notice](tx),
		repo:           repo,
		locker:         locker,
		validate:       validate,
	}
}

func (s *NoticeService) CreateNotice(ctx context.Context, req *CreateNoticeReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}

	l := s.locker.New()
	if err := l.Lock("notice:create"); err != nil {
		return err
	}
	defer l.Unlock()

	var notice model.Notice
	if err := copier.Copy(&notice, req); err != nil {
		return err
	}
	return s.repo.WithContext(ctx).Save(&notice)
}

func (s *NoticeService) UpdateNotice(ctx context.Context, req *UpdateNoticeReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}

	l := s.locker.New()
	if err := l.Lock("notice:update:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	notice, err := s.repo.WithContext(ctx).Detail(req.ID)
	if err != nil {
		return err
	}
	if err = copier.Copy(notice, req); err != nil {
		return err
	}
	return s.repo.WithContext(ctx).Save(notice)
}

func (s *NoticeService) DeleteNotice(ctx context.Context, id uint) error {
	l := s.locker.New()
	if err := l.Lock("notice:delete:" + cast.ToString(id)); err != nil {
		return err
	}
	defer l.Unlock()

	return s.BaseRepository.Delete(id)
}

func (s *NoticeService) ChangeStatus(ctx context.Context, req *db.ChangeStatus) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}

	l := s.locker.New()
	if err := l.Lock("notice:status:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	return s.BaseRepository.ChangeStatus(req)
}

func (s *NoticeService) RetrieveNotices(ctx context.Context, req *RetrieveNoticeReq) (count int64, list []model.Notice, err error) {
	if err := s.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	return s.repo.WithContext(ctx).Retrieve(req.Page, req.PageSize, func(tx *gorm.DB) {
		db.AppendWhereFromStruct(tx, req)
		tx.Order("created_at desc")
	})
}
```

## 4. `app/handle/v1/notice_handle.go`

```go
package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type NoticeHandle struct {
	svc *service.NoticeService
}

func NewNoticeHandle(svc *service.NoticeService) *NoticeHandle {
	return &NoticeHandle{svc: svc}
}

func (h *NoticeHandle) ModuleName() (string, string) { return "notice", "公告管理" }

func (h *NoticeHandle) Create(ctx *gin.Context) {
	var req service.CreateNoticeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := h.svc.CreateNotice(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *NoticeHandle) Update(ctx *gin.Context) {
	var req service.UpdateNoticeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.ID = cast.ToUint(ctx.Param("id"))
	if err := h.svc.UpdateNotice(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *NoticeHandle) Delete(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if err := h.svc.DeleteNotice(ctx.Request.Context(), id); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *NoticeHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveNoticeReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	count, list, err := h.svc.RetrieveNotices(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), req.Page, req.PageSize, list)
}

func (h *NoticeHandle) ChangeStatus(ctx *gin.Context) {
	var req db.ChangeStatus
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.ID = cast.ToUint(ctx.Param("id"))
	if err := h.svc.ChangeStatus(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}
```

## 5. 修改 `app/app.go`

在 import 中补：

```go
v1 "bit-labs.cn/owl-admin/app/handle/v1"
"bit-labs.cn/owl-admin/app/repository"
"bit-labs.cn/owl-admin/app/service"
```

在 `Binds()` 中补 3 行：

```go
v1.NewNoticeHandle,
service.NewNoticeService,
repository.NewNoticeRepository,
```

最安全的插入位置：

- handle 构造函数区：靠近 `v1.NewPositionHandle`
- service 构造函数区：靠近 `service.NewPositionService`
- repository 构造函数区：靠近 `repository.NewPositionRepository`

## 6. 修改 `app/route/api.go`

### 6.1 顶部变量

把：

```go
var userMenu, roleMenu, menuMenu, apiMenu, deptMenu, dictMenu, positionMenu *router.Menu
```

改成：

```go
var userMenu, roleMenu, menuMenu, apiMenu, deptMenu, dictMenu, positionMenu, noticeMenu *router.Menu
```

### 6.2 `InitMenu()` 挂到 `System`

在 `System` 的 `Children` 中追加：

```go
noticeMenu,
```

### 6.3 `InitApi()` 的 `Invoke` 参数

在 `positionHandle *v1.PositionHandle,` 之后追加：

```go
noticeHandle *v1.NoticeHandle,
```

### 6.4 新增路由块

建议放在 `position` 块后面：

```go
// notice
{
	r := router.NewRouteInfoBuilder(appName, noticeHandle, gv1, router.MenuOption{
		ComponentName: "SystemNotice",
		Path:          "/system/notice/index",
		Icon:          "ep:bell",
	})

	r.Post("/notices", router.AccessAuthorized, noticeHandle.Create).Name("创建公告").Build()
	r.Delete("/notices/:id", router.AccessAuthorized, noticeHandle.Delete).Name("删除公告").Build()
	r.Put("/notices/:id", router.AccessAuthorized, noticeHandle.Update).Name("更新公告").Build()
	r.Put("/notices/:id/status", router.AccessAuthorized, noticeHandle.ChangeStatus).Name("修改公告状态").Build()
	r.Get("/notices", router.AccessAuthorized, noticeHandle.Retrieve).Name("公告列表").Build()

	noticeMenu = r.GetMenu()
}
```

## 7. 修改 `app/database/auto_migrate_gen.go`

在 `&Position{},` 后追加：

```go
&Notice{},
```

## 8. 开发顺序建议

按这个顺序改最稳：

1. `model`
2. `repository`
3. `service`
4. `handle`
5. `app.go` 的 `Binds()`
6. `route/api.go`
7. `database/auto_migrate_gen.go`

原因：

- 先把依赖链补齐，再接 DI
- 最后再接路由，能减少 `Invoke` 报错时的定位成本

## 9. 这个示例能直接迁移成什么

你可以把 `notice` 替换成任何普通管理模块：

- `announcement`
- `banner`
- `article_category`
- `tenant_notice`

只要它是**单表 CRUD + 状态切换**，都可以直接按本模板替换字段名、路由名、菜单路径和中文文案。

## 完成定义

- 不看源码，只看本文，就能在 `owl-admin` 中加出一个新的标准 CRUD 模块。
- 若需要更复杂能力，再组合 [06-advanced-patterns-and-pitfalls.md](06-advanced-patterns-and-pitfalls.md) 中的 `dict`、`dept`、`role/user` 模式。
