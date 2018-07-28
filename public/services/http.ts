import { analytics } from "@fider/services";

export interface ErrorItem {
  field?: string;
  message: string;
}

export interface Failure {
  errors?: ErrorItem[];
}

export interface Result<T = void> {
  ok: boolean;
  data: T;
  error?: Failure;
}

async function toResult<T>(response: Response): Promise<Result<T>> {
  const body = await response.json();
  if (response.status < 400) {
    return {
      ok: true,
      data: body as T
    };
  } else {
    return {
      ok: false,
      data: body as T,
      error: {
        errors: body.errors
      }
    };
  }
}
async function request<T>(url: string, method: "GET" | "POST" | "PUT" | "DELETE", body?: any): Promise<Result<T>> {
  const headers = [["Accept", "application/json"], ["Content-Type", "application/json"]];
  const response = await fetch(url, {
    method,
    headers,
    body: JSON.stringify(body),
    credentials: "same-origin"
  });
  return await toResult<T>(response);
}

export const http = {
  get: async <T = void>(url: string): Promise<Result<T>> => {
    return await request<T>(url, "GET");
  },
  post: async <T = void>(url: string, body?: {}): Promise<Result<T>> => {
    return await request<T>(url, "POST", body);
  },
  put: async <T = void>(url: string, body?: {}): Promise<Result<T>> => {
    return await request<T>(url, "PUT", body);
  },
  delete: async <T = void>(url: string, body?: {}): Promise<Result<T>> => {
    return await request<T>(url, "DELETE", body);
  },
  event: (category: string, action: string) => <T>(result: Result<T>): Result<T> => {
    if (result && result.ok) {
      analytics.event(category, action);
    }
    return result;
  }
};
