package model

import "time"

type AppVersion struct {
	ID          int64      `gorm:"primaryKey;column:id;comment:主键" json:"id,string"`
	Version     *string    `gorm:"column:version;comment:版本号" json:"version,omitempty"`
	VersionName *string    `gorm:"column:version_name;comment:版本名称" json:"versionName,omitempty"`
	ApkURL      *string    `gorm:"column:apk_url;comment:APK下载地址" json:"apkUrl,omitempty"`
	ApkType     *int32     `gorm:"column:apk_type;comment:APK类型" json:"apkType,omitempty"`
	AuditState  *int32     `gorm:"column:audit_state;comment:审核状态" json:"auditState,omitempty"`
	Content     *string    `gorm:"column:content;comment:更新内容" json:"content,omitempty"`
	Remark      *string    `gorm:"column:remark;comment:备注" json:"remark,omitempty"`
	Status      *int32     `gorm:"column:status;comment:状态" json:"status,omitempty"`
	CreateTime  *time.Time `gorm:"column:create_time;comment:创建时间" json:"createTime,omitempty"`
	UpdateTime  *time.Time `gorm:"column:update_time;comment:更新时间" json:"updateTime,omitempty"`
	CreatorID   *int64     `gorm:"column:creator_id;comment:创建人ID" json:"creatorId,omitempty"`
	ModifierID  *int64     `gorm:"column:modifier_id;comment:修改人ID" json:"modifierId,omitempty"`
	AutiTime    *time.Time `gorm:"column:auti_time;comment:审核时间" json:"autiTime,omitempty"`
	AutiID      *int64     `gorm:"column:auti_id;comment:审核人ID" json:"autiId,omitempty"`
	UseOrgID    *int64     `gorm:"column:use_org_id;comment:使用组织ID" json:"useOrgId,omitempty"`
	CreateOrgID *int64     `gorm:"column:create_org_id;comment:创建组织ID" json:"createOrgId,omitempty"`
	ExtField1   *string    `gorm:"column:ext_field1;comment:扩展字段1" json:"extField1,omitempty"`
	ExtField2   *string    `gorm:"column:ext_field2;comment:扩展字段2" json:"extField2,omitempty"`
	ExtField3   *string    `gorm:"column:ext_field3;comment:扩展字段3" json:"extField3,omitempty"`
	ExtField4   *string    `gorm:"column:ext_field4;comment:扩展字段4" json:"extField4,omitempty"`
	ExtField5   *string    `gorm:"column:ext_field5;comment:扩展字段5" json:"extField5,omitempty"`
}

func (AppVersion) TableName() string {
	return "admin_app_version"
}
