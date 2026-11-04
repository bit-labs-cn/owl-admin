import { defineSubsystem } from "@bit-labs.cn/owl-ui";

export default defineSubsystem({
  name: "admin",
  viewModulesPathPrefix: "/system",
  viewModules: import.meta.glob("./views/**/*.{vue,tsx}")
});
