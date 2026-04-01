export interface DeptFormData {
  id: string;
  higherDeptOptions: Record<string, unknown>[];
  parentId: string;
  name: string;
  principal: string;
  phone: string;
  email: string;
  sort: number;
  status: number;
  description: string;
}
