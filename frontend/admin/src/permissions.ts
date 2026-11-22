/**
 * owl-admin 管理端按钮权限标识。
 * 与后端 `app/route/api.go` 中 RouteInfoBuilder 生成的 permission 严格一致：
 * `admin:<moduleEn>:<处理函数名>`。
 */
export const ADMIN_PERMS = {
  user: {
    create: "admin:user:Create",
    update: "admin:user:Update",
    delete: "admin:user:Delete",
    changeStatus: "admin:user:ChangeStatus",
    batchDelete: "admin:user:BatchDelete",
    batchChangeStatus: "admin:user:BatchChangeStatus",
    batchAssignRoles: "admin:user:BatchAssignRoles",
    batchAssignDept: "admin:user:BatchAssignDept",
    resetPassword: "admin:user:ResetPassword",
    assignRoles: "admin:user:AssignRolesToUser",
    getRoleIds: "admin:user:GetRoleIdsByUserId",
    import: "admin:user:Import",
    importTemplate: "admin:user:ImportTemplate",
    exportImportErrors: "admin:user:ExportImportErrors"
  },
  role: {
    create: "admin:role:Create",
    update: "admin:role:Update",
    delete: "admin:role:Delete",
    changeStatus: "admin:role:ChangeStatus",
    getUsers: "admin:role:GetUsersByRoleID",
    assignMenus: "admin:role:AssignMenusToRole"
  },
  userGroup: {
    create: "admin:userGroup:Create",
    update: "admin:userGroup:Update",
    delete: "admin:userGroup:Delete",
    changeStatus: "admin:userGroup:ChangeStatus",
    getUsers: "admin:userGroup:GetUsersByGroupID",
    getUserIds: "admin:userGroup:GetUserIDsByGroupID"
  },
  dept: {
    create: "admin:dept:Create",
    update: "admin:dept:Update",
    delete: "admin:dept:Delete"
  },
  position: {
    create: "admin:position:Create",
    update: "admin:position:Update",
    delete: "admin:position:Delete",
    changeStatus: "admin:position:ChangeStatus"
  },
  dict: {
    create: "admin:dict:Create",
    update: "admin:dict:Update",
    delete: "admin:dict:Delete",
    createItem: "admin:dict:CreateItem",
    updateItem: "admin:dict:UpdateItem",
    deleteItem: "admin:dict:DeleteItem"
  },
  setting: {
    update: "admin:setting:Update"
  },
  appVersion: {
    create: "admin:app-version:Create",
    update: "admin:app-version:Update",
    delete: "admin:app-version:Delete",
    changeStatus: "admin:app-version:ChangeStatus",
    upload: "admin:app-version:Upload"
  },
  legal: {
    createDocument: "admin:legal:CreateDocument",
    updateDocument: "admin:legal:UpdateDocument",
    deleteDocument: "admin:legal:DeleteDocument",
    createVersion: "admin:legal:CreateVersion",
    updateVersion: "admin:legal:UpdateVersion",
    deleteVersion: "admin:legal:DeleteVersion",
    publishVersion: "admin:legal:PublishVersion",
    createScene: "admin:legal:CreateScene",
    updateScene: "admin:legal:UpdateScene",
    deleteScene: "admin:legal:DeleteScene"
  },
  event: {
    retryOutbox: "admin:event:RetryOutbox",
    outboxDetail: "admin:event:OutboxDetail",
    inboxDetail: "admin:event:InboxDetail",
    cleanup: "admin:event:Cleanup",
    cleanupConsumers: "admin:event:CleanupConsumers"
  },
  monitor: {
    loginLogs: "admin:monitor:LoginLogs",
    operationLogs: "admin:monitor:OperationLogs"
  },
  server: {
    snapshot: "admin:server:Snapshot"
  },
  password: {
    generate: "admin:password:Generate",
    hash: "admin:password:Hash"
  },
  scheduler: {
    list: "admin:scheduler:RetrieveJobs",
    detail: "admin:scheduler:JobDetail",
    create: "admin:scheduler:CreateJob",
    update: "admin:scheduler:UpdateJob",
    delete: "admin:scheduler:DeleteJob",
    changeStatus: "admin:scheduler:ChangeStatus",
    trigger: "admin:scheduler:TriggerJob",
    runs: "admin:scheduler:RetrieveRuns",
    runDetail: "admin:scheduler:RunDetail",
    retryRun: "admin:scheduler:RetryRun",
    cancelRun: "admin:scheduler:CancelRun",
    cleanup: "admin:scheduler:CleanupRuns",
    previewCron: "admin:scheduler:PreviewCron",
    stats: "admin:scheduler:Stats",
    actions: "admin:scheduler:RetrieveActions",
    actionOptions: "admin:scheduler:ActionOptions"
  }
} as const;
