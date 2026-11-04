export interface DiskSnapshot {
  mount: string;
  fsType: string;
  total: number;
  used: number;
  percent: number;
}

export interface HostSnapshot {
  hostname: string;
  os: string;
  platform: string;
  platformVersion: string;
  arch: string;
  uptimeSeconds: number;
  hostError?: string;
  cpuModel: string;
  cpuCores: number;
  cpuPercent: number;
  cpuUser: number;
  cpuSystem: number;
  cpuIOWait: number;
  cpuError?: string;
  memoryTotal: number;
  memoryUsed: number;
  memoryAvailable: number;
  memoryPercent: number;
  memoryError?: string;
  netRecvBytes: number;
  netSentBytes: number;
  netOK: boolean;
  netError?: string;
  diskReadBytes: number;
  diskWriteBytes: number;
  diskIOOK: boolean;
  diskIOError?: string;
  load1: number;
  loadKnown: boolean;
  temperatureC: number;
  temperatureKnown: boolean;
  processCount: number;
  disks: DiskSnapshot[];
  diskError?: string;
}

export interface RuntimeSnapshot {
  goVersion: string;
  numCPU: number;
  goroutines: number;
  heapAlloc: number;
  heapSys: number;
  sys: number;
  numGC: number;
  uptimeSeconds: number;
}

export interface DatabaseSnapshot {
  ok: boolean;
  error?: string;
  open: number;
  inUse: number;
  idle: number;
  waitCount: number;
  maxOpen: number;
  pingMilliseconds: number;
}

export interface RedisSnapshot {
  ok: boolean;
  error?: string;
  version: string;
  uptimeSeconds: number;
  usedMemory: number;
  usedMemoryHuman: string;
  connectedClients: number;
  keyspaceHits: number;
  keyspaceMisses: number;
  hitRate: number;
  hitRateKnown: boolean;
}

export interface ServerSnapshot {
  collectedAt: string;
  host: HostSnapshot;
  runtime: RuntimeSnapshot;
  database: DatabaseSnapshot;
  redis: RedisSnapshot;
}
