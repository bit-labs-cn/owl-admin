package service

import (
	"context"
	"fmt"
	htmltemplate "html/template"
	"regexp"
	"strings"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
)

var typeCodePattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,63}$`)

// TypeInput 是消息类型的保存参数。
type TypeInput struct {
	ID              uint
	Code            string
	Name            string
	Description     string
	Kind            int
	DefaultChannels []string
	Status          int
}

// TemplateInput 是消息模板的保存参数。模板不带渠道，渠道由发送时决定。
type TemplateInput struct {
	ID       uint
	TypeCode string
	Title    string
	Body     string
	Link     string
}

// DeliveryView 是带接收人名称的发送记录。
type DeliveryView struct {
	repository.DeliveryDTO
	UserName string `json:"userName"`
}

// TypeOption 是收件箱筛选和发送页用的类型名称。
// Kind 是类型自带的种类，发送页按它决定这条消息是通知还是待办，不再另选。
// DefaultChannels 是该类型的默认渠道，发送页把它当作不可更改的渠道直接带上。
type TypeOption struct {
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Kind            int      `json:"kind"`
	DefaultChannels []string `json:"defaultChannels"`
}

func (s *NotificationService) AdminTypes(ctx context.Context, q repository.TypeQuery) (int64, []model.NotifyType, error) {
	return s.catalog.ListTypes(ctx, q)
}

func (s *NotificationService) TypeOptions(ctx context.Context) ([]TypeOption, error) {
	list, err := s.catalog.ListEnabledTypes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TypeOption, 0, len(list))
	for _, row := range list {
		channels, err := decodeChannels(row.DefaultChannels)
		if err != nil {
			return nil, Invalid("消息类型的默认渠道无法读取")
		}
		out = append(out, TypeOption{
			Code:            row.Code,
			Name:            row.Name,
			Kind:            row.Kind,
			DefaultChannels: normalizeChannels(channels),
		})
	}
	return out, nil
}

func (s *NotificationService) SaveType(ctx context.Context, in TypeInput) error {
	code := strings.TrimSpace(in.Code)
	name := strings.TrimSpace(in.Name)
	if !typeCodePattern.MatchString(code) {
		return Invalid("类型编码需以小写字母开头，只能包含小写字母、数字、点、下划线和短横线")
	}
	if name == "" {
		return Invalid("类型名称不能为空")
	}
	if in.Kind != model.KindNotice && in.Kind != model.KindTodo {
		return Invalid("类型种类无效")
	}
	if in.Status != model.TypeStatusEnabled && in.Status != model.TypeStatusDisabled {
		return Invalid("类型状态无效")
	}
	channels := normalizeChannels(in.DefaultChannels)
	for _, channel := range channels {
		if !knownChannel(channel) {
			return Invalid("默认渠道无效")
		}
	}
	if in.ID == 0 {
		exist, err := s.catalog.FindTypeByCode(ctx, code)
		if err != nil {
			return err
		}
		if exist != nil {
			return Conflict("类型编码已存在")
		}
		return s.catalog.SaveType(ctx, &model.NotifyType{
			Code:            code,
			Name:            name,
			Description:     strings.TrimSpace(in.Description),
			Kind:            in.Kind,
			DefaultChannels: encodeChannels(channels),
			Status:          in.Status,
		})
	}
	current, err := s.findTypeByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if current.Code != code {
		return Invalid("类型编码不能修改")
	}
	current.Name = name
	current.Description = strings.TrimSpace(in.Description)
	current.Kind = in.Kind
	current.DefaultChannels = encodeChannels(channels)
	current.Status = in.Status
	return s.catalog.SaveType(ctx, current)
}

func (s *NotificationService) DeleteType(ctx context.Context, id uint) error {
	current, err := s.findTypeByID(ctx, id)
	if err != nil {
		return err
	}
	n, err := s.catalog.CountTemplates(ctx, current.Code)
	if err != nil {
		return err
	}
	if n > 0 {
		return Conflict("请先删除该类型下的模板")
	}
	return s.catalog.DeleteType(ctx, id)
}

func (s *NotificationService) AdminTemplates(ctx context.Context, q repository.TemplateQuery) (int64, []model.NotifyTemplate, error) {
	listTotal, list, err := s.catalog.ListTemplates(ctx, q)
	return listTotal, list, err
}

func (s *NotificationService) SaveTemplate(ctx context.Context, in TemplateInput) error {
	typeCode := strings.TrimSpace(in.TypeCode)
	if typeCode == "" {
		return Invalid("请选择消息类型")
	}
	typ, err := s.catalog.FindTypeByCode(ctx, typeCode)
	if err != nil {
		return err
	}
	if typ == nil {
		return NotFound("消息类型不存在")
	}
	if err := templateSyntax(in.Title); err != nil {
		return err
	}
	// 正文是富文本模板，按 HTML 模板校验语法
	if err := htmlTemplateSyntax(in.Body); err != nil {
		return err
	}
	if err := templateSyntax(in.Link); err != nil {
		return err
	}
	row := &model.NotifyTemplate{
		TypeCode: typeCode,
		Title:    in.Title,
		Body:     in.Body,
		Link:     in.Link,
	}
	if in.ID == 0 {
		exist, err := s.catalog.FindTemplate(ctx, typeCode)
		if err != nil {
			return err
		}
		if exist != nil {
			return Conflict("该类型已有模板，请直接编辑")
		}
		return s.catalog.SaveTemplate(ctx, row)
	}
	current, err := s.findTemplateByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if current.TypeCode != typeCode {
		return Invalid("模板的消息类型不能修改")
	}
	current.Title = in.Title
	current.Body = in.Body
	current.Link = in.Link
	return s.catalog.SaveTemplate(ctx, current)
}

func (s *NotificationService) DeleteTemplate(ctx context.Context, id uint) error {
	if _, err := s.findTemplateByID(ctx, id); err != nil {
		return err
	}
	return s.catalog.DeleteTemplate(ctx, id)
}

// PreviewTemplate 用示例变量渲染模板，返回渲染后的标题、正文和链接。
// 正文按 HTML 返回：站内信和邮件直接展示，纯文本渠道再按渠道转换。
func (s *NotificationService) PreviewTemplate(title, body, link string, vars map[string]any) (channelCopy, error) {
	tpl := model.NotifyTemplate{Title: title, Body: body, Link: link}
	rendered, err := renderTemplate(&tpl, vars, "", "", "")
	if err != nil {
		return channelCopy{}, Invalid(err.Error())
	}
	return rendered.forChannel(model.ChannelInApp), nil
}

func (s *NotificationService) AdminDeliveries(ctx context.Context, q repository.DeliveryQuery) (int64, []DeliveryView, error) {
	rows, err := s.catalog.ListDeliveriesByMessage(ctx, q.MessageID)
	if err != nil {
		return 0, nil, err
	}
	views, err := s.decorateDeliveries(ctx, rows)
	if err != nil {
		return 0, nil, err
	}
	return int64(len(views)), views, nil
}

// AdminDeliveryBatches 按消息聚合一页发送批次，筛选条件作用于明细。
func (s *NotificationService) AdminDeliveryBatches(ctx context.Context, q repository.DeliveryQuery) (int64, []repository.DeliveryBatchDTO, error) {
	total, list, err := s.catalog.ListDeliveryBatches(ctx, q)
	if err != nil {
		return 0, nil, err
	}
	for i := range list {
		if list[i].TypeName == "" && list[i].Title != "" {
			list[i].TypeName = "未分类"
		}
		if list[i].SenderName == "" {
			list[i].SenderName = "系统"
		}
	}
	return total, list, nil
}

func (s *NotificationService) decorateDeliveries(ctx context.Context, rows []repository.DeliveryDTO) ([]DeliveryView, error) {
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		if row.UserID > 0 {
			ids = append(ids, row.UserID)
		}
	}
	contacts, err := s.catalog.ListContacts(ctx, dedupeUsers(ids))
	if err != nil {
		return nil, err
	}
	names := map[uint]string{}
	for _, contact := range contacts {
		names[contact.ID] = contactName(contact)
	}
	out := make([]DeliveryView, 0, len(rows))
	for _, row := range rows {
		view := DeliveryView{DeliveryDTO: row}
		switch {
		case row.RecipientType == model.RecipientTypeExternal:
			view.UserName = row.RecipientName
			if view.UserName == "" {
				view.UserName = row.RecipientEmail
			}
			if view.UserName == "" {
				view.UserName = "外部人员"
			}
		case row.RecipientType == model.RecipientTypeBroadcast:
			view.UserName = "群机器人"
		case row.UserID > 0:
			view.UserName = names[row.UserID]
			if view.UserName == "" {
				view.UserName = fmt.Sprintf("未知用户(%d)", row.UserID)
			}
		default:
			view.UserName = "未知接收人"
		}
		out = append(out, view)
	}
	return out, nil
}

func (s *NotificationService) findTypeByID(ctx context.Context, id uint) (*model.NotifyType, error) {
	var row model.NotifyType
	err := s.db.WithContext(ctx).First(&row, id).Error
	if err != nil {
		return nil, NotFound("消息类型不存在")
	}
	return &row, nil
}

func (s *NotificationService) findTemplateByID(ctx context.Context, id uint) (*model.NotifyTemplate, error) {
	var row model.NotifyTemplate
	err := s.db.WithContext(ctx).First(&row, id).Error
	if err != nil {
		return nil, NotFound("消息模板不存在")
	}
	return &row, nil
}

func knownChannel(code string) bool {
	switch code {
	case model.ChannelInApp, model.ChannelEmail, model.ChannelSMS, model.ChannelWeCom, model.ChannelDingTalk:
		return true
	default:
		return false
	}
}

func contactName(contact repository.UserContact) string {
	name := strings.TrimSpace(contact.Nickname)
	if name == "" {
		name = strings.TrimSpace(contact.Username)
	}
	return name
}

func templateSyntax(src string) error {
	if strings.TrimSpace(src) == "" {
		return nil
	}
	_, err := renderText(src, map[string]any{})
	if err == nil || strings.Contains(err.Error(), "模板变量缺失") {
		return nil
	}
	return Invalid(err.Error())
}

// htmlTemplateSyntax 只校验富文本模板的语法。
// 这里只 Parse 不执行，避免 html/template 的上下文推断把缺变量也当成语法错误。
func htmlTemplateSyntax(src string) error {
	if strings.TrimSpace(src) == "" {
		return nil
	}
	if _, err := htmltemplate.New("notify").Parse(src); err != nil {
		return Invalid(fmt.Sprintf("正文模板语法错误: %v", err))
	}
	return nil
}
