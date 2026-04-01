import { reactive } from "vue";
import type { FormRules } from "element-plus";

export const dictRules = reactive(<FormRules>{
  name: [{ required: true, message: "字典名称必填", trigger: "blur" }],
  type: [{ required: true, message: "字典英文名必填", trigger: "blur" }]
});
