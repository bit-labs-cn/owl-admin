export interface AnnouncementReadRecipient {
  userId: string;
  userName: string;
  username: string;
  email: string;
  readStatus: number;
  readAt?: string;
  isDismissed: number;
}

export interface AnnouncementReadView {
  announcementId: string;
  title: string;
  targetScope: string;
  recipientCount: number;
  readCount: number;
  unreadCount: number;
  dismissedCount: number;
  recipients: AnnouncementReadRecipient[];
}
