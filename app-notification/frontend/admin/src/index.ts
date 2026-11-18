import { defineSubsystem } from "@bit-labs.cn/owl-ui";
import HeaderNotice from "./components/HeaderNotice.vue";
import AnnouncementHost from "./components/AnnouncementHost.vue";

export default defineSubsystem({
  name: "notification",
  viewModulesPathPrefix: "/system/notify",
  viewModules: import.meta.glob("./views/**/*.{vue,tsx}"),
  layoutSlots: {
    header2: HeaderNotice,
    banner: AnnouncementHost
  }
});
