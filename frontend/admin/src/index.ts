import { defineSubsystem } from "@bit-labs.cn/owl-ui";
import { loadPublicSite } from "./loadPublicSite";

export default defineSubsystem({
  name: "admin",
  viewModulesPathPrefix: "/system",
  viewModules: import.meta.glob("./views/**/*.{vue,tsx}"),
  async install() {
    await loadPublicSite();
  }
});
