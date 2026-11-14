export interface ApiErrorResponse {
  error: {
    code: string
    message: string
    details?: unknown
  }
}

// `details` of a QUERY_FAILED error: PostgreSQL's description of why it
// rejected the statement.
export interface QueryErrorDetails {
  sqlState: string
  detail?: string
  hint?: string
  // 1-based character offset of the error in the statement.
  position?: number
}

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
    public readonly details?: unknown,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}
