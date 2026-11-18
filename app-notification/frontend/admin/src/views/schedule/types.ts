export interface ScheduleRow {
  id: string;
  kind: number;
  typeCode: string;
  typeName: string;
  title: string;
  channels: string[];
  audienceSummary: string;
  recipientCount: number;
  planAt: string;
  status: number;
  attempts: number;
  lastError: string;
  messageId: string;
  senderId: string;
  senderName: string;
  sentAt?: string;
  createdAt: string;
}
