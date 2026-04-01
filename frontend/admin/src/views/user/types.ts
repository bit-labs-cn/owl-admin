export interface UserFormData {
  id?: number;
  title: string;
  higherDeptOptions: Record<string, unknown>[];
  parentId: number;
  nickname: string;
  username: string;
  password: string;
  phone: string | number;
  email: string;
  sex: string | number;
  status: number;
  avatar?: string;
  dept?: {
    id?: number;
    name?: string;
  };
  remark: string;
  /** 打开裁剪头像弹窗，仅修改用户时由 useUserList 注入 */
  openCropDialog?: (initialAvatar: string, onDone: (url: string) => void) => void;
}

export interface RoleFormItemProps {
  username: string;
  nickname: string;
  roleOptions: any[];
  ids: Record<number, unknown>[];
}
