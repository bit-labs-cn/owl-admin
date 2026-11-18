import { createFlexAdmin } from "@bit-labs.cn/owl-ui/bootstrap";
import notificationSubsystem from "./src/index";

createFlexAdmin({
  subsystems: [notificationSubsystem]
}).then(app => app.mount("#app"));
