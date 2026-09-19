import type { StatusTone } from "@nexus/ui";
import type { ComponentHealth, OrderStatus } from "./api-types";

export function orderStatusTone(status: OrderStatus): StatusTone {
  switch (status) {
    case "Created":
      return "pending";
    case "Settled":
      return "positive";
    case "Cancelled":
    case "Expired":
      return "negative";
    default:
      return "neutral";
  }
}

export function componentHealthTone(status: ComponentHealth): StatusTone {
  switch (status) {
    case "Operational":
      return "positive";
    case "Degraded":
      return "pending";
    case "Unavailable":
      return "negative";
    case "Unknown":
      return "unknown";
    default:
      return "unknown";
  }
}
