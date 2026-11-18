export interface MessageRow {
  id: string;
  kind: number;
  typeCode: string;
  typeName: string;
  title: string;
  senderId: string;
  senderName: string;
  recipientCount: number;
  readCount: number;
  unreadCount: number;
  todoPending: number;
  todoDone: number;
  deliveryFailed: number;
  createdAt?: string;
}

export interface MessageRecipient {
  id: string;
  userId: string;
  userName: string;
  username: string;
  email: string;
  readStatus: number;
  readAt?: string;
  todoStatus: number;
  doneAt?: string;
}

export interface MessageDetail extends MessageRow {
  content: string;
  link: string;
  dueAt?: string;
  recipients: MessageRecipient[];
}
