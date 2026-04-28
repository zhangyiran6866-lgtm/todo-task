// <apifox-generated module="log">
import request from "@/utils/request";

export type LogChannel = "app" | "error" | "audit";
export type LogLevel = "debug" | "info" | "warn" | "error";

export interface GetLogsReq {
  channel?: LogChannel;
  level?: LogLevel;
  module?: string;
  keyword?: string;
  start_at?: string;
  end_at?: string;
  page?: number;
  page_size?: number;
  limit?: number;
  cursor?: string;
}

export interface LogItem {
  id: string;
  channel: string;
  timestamp: string;
  level: string;
  module: string;
  action?: string;
  message: string;
  request_id?: string;
  user_id?: string;
  method?: string;
  path?: string;
  route?: string;
  client_ip?: string;
  status_code?: number;
  latency_ms?: number;
  error?: string;
  raw?: Record<string, unknown>;
}

export interface GetLogsResp {
  items: LogItem[];
  pagination: {
    total: number;
    page: number;
    page_size: number;
    total_pages: number;
    has_next: boolean;
    has_prev: boolean;
  };
  next_cursor: string;
}

export interface GetLogsIdReq {
  channel?: LogChannel;
}

export type GetLogsIdResp = LogItem;

export const logApi = {
  getLogs(params?: GetLogsReq): Promise<GetLogsResp> {
    return request.get("/logs", { params }) as unknown as Promise<GetLogsResp>;
  },

  getLogsId(id: string, params?: GetLogsIdReq): Promise<GetLogsIdResp> {
    return request.get(`/logs/${id}`, { params }) as unknown as Promise<GetLogsIdResp>;
  },
};
// </apifox-generated module="log">
