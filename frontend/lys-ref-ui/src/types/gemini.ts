import type { ApiCallBase } from "@/types/app"

export type ApiCall = Omit<ApiCallBase, 'attempt' | 'method' | 'status_code'>