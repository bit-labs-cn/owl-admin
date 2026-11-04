export interface LegalDocumentRow {
  id: string;
  key: string;
  name: string;
  kind: string;
  kindLabel: string;
  status: string;
  statusLabel: string;
  currentVersion: string;
  sceneNames: string[];
}

export interface LegalDocumentForm {
  id?: string;
  key: string;
  name: string;
  kind: string;
  status: string;
}

export interface LegalOption {
  id?: string;
  key: string;
  name: string;
  kind?: string;
  status?: string;
}

export interface LegalVersionRow {
  id: string;
  documentId: string;
  version: string;
  title: string;
  content: string;
  contentSha256: string;
  changeClass: string;
  changeLabel: string;
  status: string;
  statusLabel: string;
  publishedAt?: string;
  createdAt?: string;
}

export interface LegalVersionForm {
  id?: string;
  documentId: string;
  version: string;
  title: string;
  content: string;
}

export interface LegalSceneItemForm {
  documentKey: string;
  documentName?: string;
  required: boolean;
  groupKey: string;
  wording: string;
  sort: number;
}

export interface LegalSceneRow {
  id: string;
  key: string;
  name: string;
  kind: string;
  kindLabel: string;
  itemSummary: string;
  items: LegalSceneItemForm[];
}

export interface LegalSceneForm {
  id?: string;
  key: string;
  name: string;
  kind: string;
  items: LegalSceneItemForm[];
}

export interface LegalEventRow {
  id: string;
  userLabel: string;
  username: string;
  nickname: string;
  userPresent: boolean;
  documentName: string;
  documentKey: string;
  version: string;
  action: string;
  actionLabel: string;
  sceneKey: string;
  sceneName: string;
  createdAt?: string;
  ip: string;
  userAgent: string;
}
