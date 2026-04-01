import "./reset.css";
import { type Ref, h, ref, toRaw, watch, reactive, computed, onMounted } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import type { UserFormData, RoleFormItemProps } from "./types";
import {
  ElForm,
  ElInput,
  ElFormItem,
  ElProgress,
  ElMessageBox
} from "element-plus";
import { handleTree } from "@bit-labs.cn/owl-ui/utils/tree";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import userAvatar from "@bit-labs.cn/owl-ui/assets/user.jpg";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { getKeyList, isAllEmpty, deviceDetection } from "@pureadmin/utils";
import { deptAPI } from "@bit-labs.cn/owl-admin-ui/api/dept";
import { userManageAPI as userAPI } from "@bit-labs.cn/owl-admin-ui/api/user";
import { roleAPI } from "@bit-labs.cn/owl-admin-ui/api/role";
import UserForm from "./UserForm.vue";
import UserRoleForm from "./UserRoleForm.vue";

export function useUserList(tableRef: Ref, treeRef: Ref) {
  const form = reactive({
    deptId: "",
    username: "",
    phone: "",
    status: ""
  });
  const formRef = ref();
  const ruleFormRef = ref();
  const dataList = ref([]);
  const loading = ref(true);
  const avatarInfo = ref();
  const switchLoadMap = ref({});
  const higherDeptOptions = ref();
  const treeData = ref([]);
  const treeLoading = ref(true);
  const selectedNum = ref(0);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  const buttonClass = computed(() => [
    "!h-[20px]",
    "reset-margin",
    "!text-gray-500",
    "dark:!text-white",
    "dark:hover:!text-primary"
  ]);

  const pwdForm = reactive({ newPwd: "" });
  const pwdProgress = [
    { color: "#e74242", text: "非常弱" },
    { color: "#EFBD47", text: "弱" },
    { color: "#ffa500", text: "一般" },
    { color: "#1bbf1b", text: "强" },
    { color: "#008000", text: "非常强" }
  ];
  const curScore = ref();
  const roleOptions = ref([]);

  function onChange({ row, index }) {
    ElMessageBox.confirm(
      `确认要<strong>${row.status === 0 ? "停用" : "启用"}</strong><strong style='color:var(--el-color-primary)'>${row.username}</strong>用户吗?`,
      "系统提示",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
        dangerouslyUseHTMLString: true,
        draggable: true
      }
    )
      .then(() => {
        switchLoadMap.value[index] = Object.assign(
          {},
          switchLoadMap.value[index],
          { loading: true }
        );
        setTimeout(() => {
          switchLoadMap.value[index] = Object.assign(
            {},
            switchLoadMap.value[index],
            { loading: false }
          );
          userAPI.changeStatus(row.id, row.status).then(() => {
            onSearch();
          });
        }, 300);
      })
      .catch(() => {
        row.status === 0 ? (row.status = 1) : (row.status = 0);
      });
  }

  function handleDelete(row) {
    userAPI.deleteUser(row.id).then(() => {
      onSearch();
    });
  }

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    pagination.currentPage = 1;
    onSearch();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val;
    onSearch();
  }

  function handleSelectionChange(val) {
    selectedNum.value = val.length;
    tableRef.value.setAdaptive();
  }

  function onSelectionCancel() {
    selectedNum.value = 0;
    tableRef.value.getTableRef().clearSelection();
  }

  function onbatchDel() {
    const curSelected = tableRef.value.getTableRef().getSelectionRows();
    message(`已删除用户编号为 ${getKeyList(curSelected, "id")} 的数据`, {
      type: "success"
    });
    tableRef.value.getTableRef().clearSelection();
    onSearch();
  }

  async function onSearch() {
    loading.value = true;
    const params = {
      ...toRaw(form),
      page: pagination.currentPage,
      pageSize: pagination.pageSize
    };
    const res = await userAPI.getUserList(params);
    dataList.value = res?.data?.list ?? [];
    pagination.total = res?.total ?? 0;
    pagination.pageSize = res?.pageSize ?? pagination.pageSize;
    pagination.currentPage = res?.currentPage ?? pagination.currentPage;
    setTimeout(() => {
      loading.value = false;
    }, 500);
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    form.deptId = "";
    treeRef.value.onTreeReset();
    onSearch();
  }

  function onTreeSelect({ id, selected }) {
    form.deptId = selected ? id : "";
    onSearch();
  }

  function formatHigherDeptOptions(treeList) {
    if (!treeList || !treeList.length) return;
    const newTreeList = [];
    for (let i = 0; i < treeList.length; i++) {
      treeList[i].disabled = treeList[i].status === 0;
      formatHigherDeptOptions(treeList[i].children);
      newTreeList.push(treeList[i]);
    }
    return newTreeList;
  }

  function openDialog(title = "新增", row?: UserFormData) {
    addDialog({
      title: `${title}用户`,
      props: {
        formInline: {
          title,
          higherDeptOptions: formatHigherDeptOptions(higherDeptOptions.value),
          id: row?.id ?? "",
          nickname: row?.nickname ?? "",
          username: row?.username ?? "",
          password: row?.password ?? "",
          phone: row?.phone ?? "",
          email: row?.email ?? "",
          sex: row?.sex ?? 0,
          status: row?.status ?? 1,
          avatar: row?.avatar ?? "",
          remark: row?.remark ?? "",
          parentId: (row as any)?.depts?.[0]?.id ?? "",
          openCropDialog
        }
      },
      width: "46%",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(UserForm, { ref: formRef, formInline: null }),
      beforeSure: (done, { options }) => {
        const FormRef = formRef.value.getRef();
        const curData = options.props.formInline as UserFormData;
        function chores() {
          done();
          onSearch();
        }
        FormRef.validate(valid => {
          if (valid) {
            if (title === "新增") {
              userAPI.createUser(curData).then(() => chores());
            } else {
              userAPI.updateUser(curData).then(() => {
                if (curData.avatar && curData.id) {
                  return userAPI
                    .changeAvatar(Number(curData.id), curData.avatar)
                    .then(() => chores());
                }
                chores();
              });
            }
          }
        });
      }
    });
  }

  const cropRef = ref();
  /** 打开裁剪头像弹窗，用于修改用户表单内更换头像 */
  function openCropDialog(initialAvatar: string, onDone: (url: string) => void) {
    addDialog({
      title: "裁剪、上传头像",
      width: "40%",
      closeOnClickModal: false,
      fullscreen: deviceDetection(),
      contentRenderer: () =>
        h(ReCropperPreview, {
          ref: cropRef,
          imgSrc: initialAvatar || userAvatar,
          onCropper: (info: string) => (avatarInfo.value = info)
        }),
      beforeSure: done => {
        if (avatarInfo.value) onDone(avatarInfo.value);
        done();
      },
      closeCallBack: () => cropRef.value?.hidePopover?.()
    });
  }

  watch(
    pwdForm,
    ({ newPwd }) =>
      (curScore.value = isAllEmpty(newPwd) ? -1 : zxcvbn(newPwd).score)
  );

  function handleReset(row) {
    addDialog({
      title: `重置 ${row.username} 用户的密码`,
      width: "30%",
      draggable: true,
      closeOnClickModal: false,
      fullscreen: deviceDetection(),
      contentRenderer: () => (
        <>
          <ElForm ref={ruleFormRef} model={pwdForm}>
            <ElFormItem
              prop="newPwd"
              rules={[
                { required: true, message: "请输入新密码", trigger: "blur" }
              ]}
            >
              <ElInput
                clearable
                show-password
                type="password"
                v-model={pwdForm.newPwd}
                placeholder="请输入新密码"
              />
            </ElFormItem>
          </ElForm>
          <div class="mt-4 flex">
            {pwdProgress.map(({ color, text }, idx) => (
              <div
                class="w-[19vw]"
                style={{ marginLeft: idx !== 0 ? "4px" : 0 }}
              >
                <ElProgress
                  striped
                  striped-flow
                  duration={curScore.value === idx ? 6 : 0}
                  percentage={curScore.value >= idx ? 100 : 0}
                  color={color}
                  stroke-width={10}
                  show-text={false}
                />
                <p
                  class="text-center"
                  style={{ color: curScore.value === idx ? color : "" }}
                >
                  {text}
                </p>
              </div>
            ))}
          </div>
        </>
      ),
      closeCallBack: () => (pwdForm.newPwd = ""),
      beforeSure: done => {
        ruleFormRef.value.validate(valid => {
          if (valid) {
            userAPI.resetPassword(row.id, pwdForm.newPwd).then(() => {
              done();
              onSearch();
            });
          }
        });
      }
    });
  }

  async function handleRole(row) {
    const ids = (await userAPI.getRoleIds(row.id)).data ?? [];
    addDialog({
      title: `分配 ${row.username} 用户的角色`,
      props: {
        formInline: {
          username: row?.username ?? "",
          nickname: row?.nickname ?? "",
          roleOptions: roleOptions.value ?? [],
          ids
        }
      },
      width: "400px",
      draggable: true,
      fullscreen: deviceDetection(),
      fullscreenIcon: true,
      closeOnClickModal: false,
      contentRenderer: () => h(UserRoleForm),
      beforeSure: (done, { options }) => {
        const curData = options.props.formInline as RoleFormItemProps;
        userAPI
          .assignRoleToUser({ id: row.id, roleIDs: curData.ids })
          .then(() => {
            done();
            onSearch();
          });
      }
    });
  }

  onMounted(async () => {
    treeLoading.value = true;
    onSearch();

    const { data } = await deptAPI.getDepts();
    higherDeptOptions.value = handleTree(data);
    treeData.value = handleTree(data);
    treeLoading.value = false;

    roleOptions.value = (await roleAPI.getRolesOption()).data;
  });

  return {
    form,
    loading,
    dataList,
    treeData,
    treeLoading,
    selectedNum,
    pagination,
    switchLoadMap,
    buttonClass,
    deviceDetection,
    onSearch,
    resetForm,
    onbatchDel,
    openDialog,
    onTreeSelect,
    handleDelete,
    handleReset,
    handleRole,
    onChange,
    handleSizeChange,
    onSelectionCancel,
    handleCurrentChange,
    handleSelectionChange
  };
}
