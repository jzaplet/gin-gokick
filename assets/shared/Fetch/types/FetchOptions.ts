export type FetchOptions<TBody = never> = {
    body?: NoInfer<TBody>;
    headers?: Record<string, string>;
};
