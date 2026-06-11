import type { ApiCallBase } from "@/types/app"

export type ApiCall = Omit<ApiCallBase, 'method' | 'status_code'>