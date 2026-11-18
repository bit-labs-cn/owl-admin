export interface DeptOption {
  id: string;
  name: string;
  children?: DeptOption[];
}

export interface RoleOption {
  id: string;
  name: string;
}
