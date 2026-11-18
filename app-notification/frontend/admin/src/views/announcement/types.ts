import {
  AnnouncementStatusDraft,
  DisplayModeInbox,
  LevelInfo,
  ScopeAll
} from "../../constants";

export interface AnnouncementRow {
  id: string;
  title: string;
  content: string;
  level: number;
  displayMode: number;
  targetScope: string;
  targetValues?: string;
  startAt?: string;
  endAt?: string;
  status: number;
  createdAt?: string;
}

export interface AnnouncementFormData {
  id?: string;
  title: string;
  content: string;
  level: number;
  displayMode: number;
  targetScope: string;
  targetValues?: string;
  startAt: string;
  endAt: string;
  status: number;
}

export function emptyAnnouncementForm(): AnnouncementFormData {
  return {
    title: "",
    content: "",
    level: LevelInfo,
    displayMode: DisplayModeInbox,
    targetScope: ScopeAll,
    startAt: "",
    endAt: "",
    status: AnnouncementStatusDraft
  };
}
