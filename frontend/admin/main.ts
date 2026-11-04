import { createFlexAdmin } from "@bit-labs.cn/owl-ui/bootstrap";
import adminSubsystem from "@bit-labs.cn/owl-admin-ui";

createFlexAdmin({
  subsystems: [adminSubsystem]
}).then(app => app.mount("#app"));
