import { apiFetch } from "./client";
import { OrderHistory, OrderStatus } from "./types";

interface GetOrderHistoryParams {
  status?: OrderStatus[];
  sort?: "deadline" | "price" | "updated_at";
  order?: "asc" | "desc";
  limit?: number;
  offset?: number;
}

export async function getOrderHistory(
  params?: GetOrderHistoryParams,
): Promise<OrderHistory> {
  const searchParams = new URLSearchParams();

  if (params?.status) {
    params.status.forEach((status) => {
      searchParams.append("status", status);
    });
  }

  if (params?.sort) {
    searchParams.set("sort", params.sort);
  }

  if (params?.order) {
    searchParams.set("order", params.order);
  }

  if (params?.limit !== undefined) {
    searchParams.set("limit", params.limit.toString());
  }

  if (params?.offset !== undefined) {
    searchParams.set("offset", params.offset.toString());
  }

  const query = searchParams.toString();

  return apiFetch<OrderHistory>(`/orders${query ? `?${query}` : ""}`);
}
