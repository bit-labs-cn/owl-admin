import { ChannelInApp, KindNotice, TypeStatusEnabled } from "../../constants";

export interface TypeRow {
  id: string;
  code: string;
  name: string;
  description: string;
  kind: number;
  defaultChannels: string;
  status: number;
}

export interface TypeFormData {
  id?: string;
  code: string;
  name: string;
  description: string;
  kind: number;
  defaultChannels: string[];
  status: number;
}

export function emptyTypeForm(): TypeFormData {
  return {
    code: "",
    name: "",
    description: "",
    kind: KindNotice,
    defaultChannels: [ChannelInApp],
    status: TypeStatusEnabled
  };
}
