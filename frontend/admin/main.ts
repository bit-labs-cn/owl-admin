import { createFlexAdmin } from "@bit-labs.cn/owl-ui/bootstrap";
import adminSubsystem from "@bit-labs.cn/owl-admin-ui";
import notificationSubsystem from "@bit-labs.cn/owl-notification-ui";

createFlexAdmin({
  subsystems: [adminSubsystem, notificationSubsystem]
}).then(app => app.mount("#app"));
