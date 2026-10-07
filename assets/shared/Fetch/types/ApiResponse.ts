import type { ApiError } from '@/shared/Fetch/types/ApiError';
import type { ApiSuccess } from '@/shared/Fetch/types/ApiSuccess';

export type ApiResponse<TData, TError> = ApiSuccess<TData> | ApiError<TError>;
