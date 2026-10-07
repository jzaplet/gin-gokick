export const errorId = (name: string): string => `${name}-error`;

export const statusId = (name: string): string => `${name}-status`;

export const describedBy = (name: string, error: string | null, status: string | null): string | undefined => {
    if (error !== null) {
        return errorId(name);
    }

    return status === null ? undefined : statusId(name);
};
