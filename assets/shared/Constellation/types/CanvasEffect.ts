import type { CanvasFxContext } from '@/shared/Constellation/types/CanvasFxContext';

export type CanvasEffect = (context: CanvasFxContext) => () => void;
