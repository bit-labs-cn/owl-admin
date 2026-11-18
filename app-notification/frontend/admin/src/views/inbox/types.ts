export interface InboxRow {
  id: string;
  messageId: string;
  kind: number;
  typeCode?: string;
  typeName?: string;
  title: string;
  content: string;
  link?: string;
  bizKey?: string;
  dueAt?: string;
  readStatus: number;
  readAt?: string;
  todoStatus: number;
  doneAt?: string;
  createdAt?: string;
}

export interface InboxCounts {
  unreadNotice: number;
  pendingTodo: number;
  total: number;
  notice?: number;
  todo?: number;
  message?: number;
}
