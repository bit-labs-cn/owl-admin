export interface DeliveryBatchRow {
  messageId: string;
  title: string;
  kind: number;
  typeName: string;
  total: number;
  successCount: number;
  failedCount: number;
  skippedCount: number;
  pendingCount: number;
  channelCount: number;
  recipientCount: number;
  senderName: string;
  createdAt?: string;
}

export interface DeliveryRow {
  id: string;
  messageId: string;
  title: string;
  userId: string;
  userName: string;
  channel: string;
  status: number;
  attempts: number;
  lastError: string;
  recipientType: number;
  recipientName: string;
  recipientEmail: string;
  sentAt?: string;
  createdAt?: string;
}
